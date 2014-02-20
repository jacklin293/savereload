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

func DirExists(path string) (bool, error) {
    fileInfo, err := os.Stat(path)
    if err != nil {
        // no such file or dir
        return false, err
    }
    if fileInfo.IsDir() {
        // it's a directory
        return true, nil
    }
    // it's a file
    return false, nil
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

func CheckIgnoreExt(fileExt string, ignoreExts []string) bool {
    for _, ignoreExt := range ignoreExts {
        if fileExt == "."+ignoreExt {
            return true
        }
    }
    return false
}

func GetAction(e *fsnotify.FileEvent) string {
    var events string = ""
    if e.IsCreate() {
        events += "|" + "CREATE"
    }
    if e.IsDelete() {
        events += "|" + "DELETE"
    }
    if e.IsModify() {
        events += "|" + "MODIFY"
    }
    if e.IsRename() {
        events += "|" + "RENAME"
    }
    if len(events) > 0 {
        events = events[1:]
    }
    return events
}

func (args *Args) watchDirectory(watcher *fsnotify.Watcher) {
    var prevActionTime int
    msg := map[string]interface{}{}
    for {
        select {
        case ev := <-watcher.Event:
            // Prevent the same action output many times.

            if prevActionTime-time.Now().Second() == 0 {
                continue
            }
            // Ignore some file extension
            if CheckIgnoreExt(filepath.Ext(ev.Name), strings.Split(args.IgnoreExt, "|")) {
                continue
            }
            prevActionTime = time.Now().Second()
            log.Println("event:", ev)
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
}

func (args *Args) ExecWatchFlow() {
    // Check path
    isDir, err := DirExists(args.Path)
    if err != nil {
        fmt.Println(err.Error())
        os.Exit(0)
    }
    if isDir {
        fmt.Println("Dir")
    } else {
        fmt.Println("File")
    }

    // Clean Path
    args.Path = filepath.Clean(args.Path)

    // Watch start
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        log.Fatal(err)
    }

    done := make(chan bool)
    go args.watchDirectory(watcher)

    err = watcher.Watch(args.Path)
    if err != nil {
        log.Fatal(err)
    }
    <-done
    watcher.Close()

}

func main() {
    args := Args{}

    flag.StringVar(&args.Path, "p", "", "The file or folder path to watch")
    flag.StringVar(&args.Cmd, "c", "", "The command to run when the folder changes")
    flag.BoolVar(&args.Recurse, "r", true, "Controls whether the watcher should recurse into subdirectories")
    flag.StringVar(&args.IgnoreExt, "ig", "swp|swpx", "Ignore file extension")

    flag.Parse()

    // Listen websocket
    // 54.250.138.78
    http.HandleFunc("/connws/", args.ConnWs)
    err := http.ListenAndServe(":9090", nil)
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
    }
}
