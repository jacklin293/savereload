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
        var prevActionSecond int
        msg := map[string]interface{}{}
        for {
            select {
            case ev := <-watcher.Event:
                // Prevent the same action output many times.
                if prevActionSecond-time.Now().Second() == 0 {
                    continue
                }

                // Ignore some file extension
                if len(args.IgnoreExt) > 0 && IsIgnoreExt(filepath.Ext(ev.Name), strings.Split(args.IgnoreExt, "|")) {
                    continue
                }

                // Must be put after ignoring file extension checking, because arise bug if first .fff.swp second fff
                prevActionSecond = time.Now().Second()

                msg["Action"] = "doReload"
                if err := args.Ws.WriteJSON(&msg); err != nil {
                    fmt.Println("watch dir - Write : " + err.Error())
                    return
                }

                fmt.Printf("Notify browser reload : %v\n", msg)

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
    isDir, err := DirExists(args.Path)
    if err != nil {
        fmt.Println(err.Error())
        os.Exit(0)
    }
    if isDir {
        fmt.Println("Path type : Dir")
    } else {
        fmt.Println("Path type : File")
    }

    // Clean Path
    args.Path = filepath.Clean(args.Path)

    // Get all subfolder
    var paths []string
    if args.Recurse {
        paths, err = Walk(args.Path)
        if err != nil {
            log.Fatalln(err)
        }
    } else {
        // Only watch one folder
        paths = append(paths, args.Path)
    }

    // Watch
    args.watch(paths)
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

func main() {
    args := Args{}
    flag.StringVar(&args.Path, "p", DefaultPath, "The file or folder path to watch")
    flag.StringVar(&args.Cmd, "c", "", "The command to run when the folder changes")
    flag.BoolVar(&args.Recurse, "r", true, "Controls whether the watcher should recurse into subdirectories")
    flag.StringVar(&args.IgnoreExt, "ig", "swp", "Ignore file extension")
    flag.Parse()

    http.HandleFunc("/connws/", args.ConnWs)
    err := http.ListenAndServe(":9112", nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
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

// TODO List :
// chrome extension  啟動 save reload 按鈕 分開為 連線及監聽按鈕要分開為兩個checkbox, 結束按鈕就不用了
// watch directory recursive
// UI input directory that i want watching
// - extensions: .html .css .js .png .gif .jpg .php .php5 .py .rb .erb
// - excluding changes in: */.git/* */.svn/* */.hg/*
