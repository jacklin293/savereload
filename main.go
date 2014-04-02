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
    "path/filepath"
    "strings"
    "time"
    "regexp"
    "savereload/gosass"
)

type Args struct {
    Port      string
    Path      string
    IsDir     bool
    Recurse   bool
    IgnoreExt string
    Ws        *websocket.Conn
    SassChecked bool
    SassSrc     string
    SassDes     string
}

var (
    DefaultPath, _ = filepath.Abs("./")
)

func main() {
    args := Args{}
    flag.StringVar(&args.Path, "p", DefaultPath, "The file or folder path to watch")
    flag.BoolVar(&args.Recurse, "r", true, "Controls whether the watcher should recurse into subdirectories")
    flag.StringVar(&args.IgnoreExt, "ig", "", "Ignore file extension. Ex: -ig=\"swp|swx|swo\"")
    flag.StringVar(&args.Port, "P", "9112", "Listen port");
    flag.Parse()

    http.HandleFunc("/connws/", args.ConnWs)
    err := http.ListenAndServe(":" + args.Port, nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}

func (args *Args) CompileSass(sassFilePath string) error {
    // Get sass source file path
    sassFullPath, err := filepath.Abs(sassFilePath)
    if err != nil {
        return err
    }

    // Assemble css file full path
    re := regexp.MustCompile("scss|sass")
    cssFileName := re.ReplaceAllString(filepath.Base(sassFullPath), "css")
    var cssDirPath string
    if (args.SassSrc == args.SassDes) {
        cssDirPath = filepath.Dir(sassFullPath)
    } else {
        // Create sass file destination path
        cssDirPath = strings.Replace(filepath.Dir(sassFilePath), args.SassSrc, args.SassDes, 1)
        if ! DirExists(cssDirPath) {
            os.MkdirAll(cssDirPath, 0755)
        }
    }
    cssFullPath := cssDirPath + string(os.PathSeparator) + cssFileName

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

                // Ignore all hidden file
                if strings.HasPrefix(filepath.Base(ev.Name), ".") {
                    log.Println("Ignore hidden file : " + ev.Name)
                    continue
                }

                // Ignore some file extension
                if len(args.IgnoreExt) > 0 && IsIgnoreExt(filepath.Ext(ev.Name), strings.Split(args.IgnoreExt, "|")) {
                    log.Println("Ignore " + ev.Name)
                    continue
                }

                // Compile sass file
                if args.SassChecked && filepath.Ext(ev.Name) == ".scss" && strings.HasPrefix(ev.Name, args.SassSrc) {
                    if err := args.CompileSass(ev.Name); err != nil {
                        log.Println("Compile scss error in watching event.")
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
        log.Println(rec)

        // Update sass checked
        if rec["Action"] == "updateSassChecked" {
            args.SassChecked    = rec["SassChecked"].(bool)
            args.SassSrc        = rec["SassSrc"].(string)
            args.SassDes        = rec["SassDes"].(string)

            // Check dir that is existent.
            if ! DirExists(args.SassSrc) {
                rec["SassSrcError"] = "Path doesn't exist."
            } else {
                rec["SassSrcError"] = ""
            }
            if ! DirExists(args.SassDes) {
                rec["SassDesError"] = "Path doesn't exist."
            } else {
                rec["SassDesError"] = ""
            }

            //
            if ! DirExists(args.SassSrc) || ! DirExists(args.SassDes) {
                args.SassChecked = false
                rec["SassChecked"] = false
            }
        }

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
