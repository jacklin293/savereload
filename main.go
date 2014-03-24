package main

import (
    "errors"
    "flag"
    "fmt"
    "github.com/gorilla/websocket"
    "github.com/howeyc/fsnotify"
    "log"
    "net/http"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "time"
    "regexp"
    "savereload/gosass"
)

type Args struct {
    Path      string
    IsDir     bool
    Cmd       string
    Recurse   bool
    IgnoreExt string
    Ws        *websocket.Conn
}

var (
    DefaultPath, _ = filepath.Abs("./")
)

func CompileSass(sassFilePath string) error {
    // Get sass source file path
    sassFullPath, err := filepath.Abs(sassFilePath)
    if err != nil {
        return err
    }
    sassDirPath := filepath.Dir(sassFullPath)

    // Assemble css file full path
    re := regexp.MustCompile("scss|sass")
    cssFileName := re.ReplaceAllString(filepath.Base(sassFullPath), "css")
    cssFullPath := sassDirPath + string(os.PathSeparator) + cssFileName

    ctx := gosass.FileContext {
        Options: gosass.Options{
            OutputStyle: gosass.NESTED_STYLE,
            SourceComments: false,
            IncludePaths: make([]string, 0),
        },
        InputPath: sassFullPath,
        OutputString: "",
        ErrorStatus: 0,
        ErrorMessage: "",
    }
    gosass.CompileFile(&ctx)

    if ctx.ErrorStatus != 0 {
        if ctx.ErrorMessage != "" {
            return errors.New(ctx.ErrorMessage)
        } else {
            return errors.New("Sass compile : Unknow error.")
        }
    } else {
        // Create css file
        var fi *os.File
        if FileExists(cssFullPath) {
            os.Remove(cssFullPath)
        }
        fi, err = os.Create(cssFullPath)
        if err != nil {
            return err
        }
        defer fi.Close()

        if _, err = fi.Write([]byte(ctx.OutputString)); err != nil {
            return err
        }
        return err
    }
}

func main() {
    args := Args{}
    flag.StringVar(&args.Path, "p", DefaultPath, "The file or folder path to watch")
    flag.StringVar(&args.Cmd, "c", "", "The command to run when the folder changes")
    flag.BoolVar(&args.Recurse, "r", true, "Controls whether the watcher should recurse into subdirectories")
    flag.StringVar(&args.IgnoreExt, "ig", "swp|swpx|swx", "Ignore file extension")
    flag.Parse()

    http.HandleFunc("/connws/", args.ConnWs)
    err := http.ListenAndServe(":9112", nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}

func RunCommand(cmd string) {
    splitCmd := strings.Split(cmd, " ")
    if strings.TrimSpace(splitCmd[0]) == "" {
        log.Fatal(errors.New(fmt.Sprintf("Command (%v) has too few args\n", cmd)))
    }
    cmdPtr := exec.Command(splitCmd[0], splitCmd[1:]...)
    cmdPtr.Stdout = os.Stdout
    cmdPtr.Stderr = os.Stderr
    err := cmdPtr.Run()
    if err != nil {
        log.Fatal(err)
    }
}

func IsIgnoreExt(fileExt string, ignoreExts []string) bool {
    for _, ignoreExt := range ignoreExts {
        if fileExt == "."+ignoreExt {
            return true
        }
    }
    return false
}

func (args *Args) watch(paths []string) {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        log.Fatal(err)
    }
    done := make(chan bool)

    go func() {
        var prevActionSecond, duration int
        msg := map[string]interface{}{}
        for {
            select {
            case ev := <-watcher.Event:
                // Prevent the same action output many times.
                duration = prevActionSecond-time.Now().Second()
                if duration == 0 {
                    continue
                }

                // Ignore some file extension
                if len(args.IgnoreExt) > 0 && IsIgnoreExt(filepath.Ext(ev.Name), strings.Split(args.IgnoreExt, "|")) {
                    fmt.Println("Ignore " + ev.Name)
                    continue
                }

                // Compile sass file
                if filepath.Ext(ev.Name) == ".scss" {
                    if err := CompileSass(ev.Name); err != nil {
                        fmt.Println("Compile scss error in watching event.")
                        continue
                    }
                }

                // Must be put after ignoring file extension checking, because arise bug if first .fff.swp second fff
                prevActionSecond = time.Now().Second()

                msg["Action"] = "doReload"
                if err := args.Ws.WriteJSON(&msg); err != nil {
                    fmt.Println("watch dir - Write : " + err.Error())
                    return
                }

                fmt.Printf("Detect %s changing, notify browser reload : %v\n", ev.Name, msg)

                if args.Cmd != "" {
                    RunCommand(args.Cmd)
                }
            case err := <-watcher.Error:
                log.Fatal(err)
            }
        }
    }()

    for _, path := range paths {
        err = watcher.Watch(path)
        if err != nil {
            log.Fatalln(err)
        }
    }
    <-done
    watcher.Close()
}

func (args *Args) ExecWatchFlow() {
    // Check path
    if ! DirExists(args.Path) {
        fmt.Printf("%s doesn't exist.\n", args.Path)
        os.Exit(0)
    }

    // Clean Path
    args.Path = filepath.Clean(args.Path)
    fmt.Printf("Watch %s ...\n", args.Path)

    // Get all subfolder
    if args.Recurse {
        paths, err := Walk(args.Path)
        if err != nil {
            log.Fatal("Walk path error")
        }

        // Watch
        args.watch(paths)
    } else {
        // Only watch one folder
        var paths []string
        paths = append(paths, args.Path)

        // Watch
        args.watch(paths)
    }

}

func Walk(rootDir string) (paths []string, err error) {
    err = filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
        if !info.IsDir() || strings.Contains(path, ".git") {
            return nil
        }
        paths = append(paths, path)
        return nil
    })
    if err != nil {
        return
    }
    return
}

func (args *Args) ConnWs(w http.ResponseWriter, r *http.Request) {
    ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
    if _, ok := err.(websocket.HandshakeError); ok {
        http.Error(w, "Not a websocket handshake", 400)
        return
    } else if err != nil {
        log.Println(err)
        return
    }

    args.Ws = ws
    go args.ExecWatchFlow()

    rec := map[string]interface{}{}
    for {
        if err = ws.ReadJSON(&rec); err != nil {
            if err.Error() == "EOF" {
                return
            }
            // ErrShortWrite means that a write accepted fewer bytes than requested but failed to return an explicit error.
            if err.Error() == "unexpected EOF" {
                return
            }
            fmt.Println("Read : " + err.Error())
            return
        }
        rec["ServerResponse"] = "Server received."
        fmt.Println(rec)

        if err = ws.WriteJSON(&rec); err != nil {
            fmt.Println("watch dir - Write : " + err.Error())
            return
        }

        // close
        if rec["Action"] == "requireClose" {
            os.Exit(0)
        }
    }
}
