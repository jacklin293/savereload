package main

import(
    "code.google.com/p/go.net/websocket"
    "github.com/howeyc/fsnotify"
    "log"
    "flag"
    "strings"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "time"
    "errors"
    "net/http"
)

type Args struct {
    Path string
    IsDir bool
    Cmd string
    Recurse bool
    IgnoreExt string
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
        if fileExt == "." + ignoreExt {
            return true
        }
    }
    return false
}

func GetAction(e *fsnotify.FileEvent) string {
    var events string = ""
    if e.IsCreate() { events += "|" + "CREATE" }
    if e.IsDelete() { events += "|" + "DELETE" }
    if e.IsModify() { events += "|" + "MODIFY" }
    if e.IsRename() { events += "|" + "RENAME" }
    if e.IsAttrib() { events += "|" + "ATTRIB" }
    if len(events) > 0 { events = events[1:] }
    return events
}

func (args *Args) watch_directory(watcher *fsnotify.Watcher) {
    var prevActionTime int
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
            if args.Cmd != "" {
                RunCommand(args.Cmd)
            }
        case err := <-watcher.Error:
            log.Fatal(err)
        }
    }
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
    http.HandleFunc("/", Home)
    http.Handle("/wat/", websocket.Handler(Wat))
    err := http.ListenAndServe(":9090", nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }

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
    go args.watch_directory(watcher)

    err = watcher.Watch(args.Path)
    if err != nil {
        log.Fatal(err)
    }
    <-done
    watcher.Close()
}

func Wat(ws *websocket.Conn) {
    var err error
    var rec string

    for {
        err = websocket.JSON.Receive(ws, &rec)
        if err != nil {
            fmt.Println(err.Error())
            break
        }
        rec = "Server receives : " + rec
        fmt.Println(rec)

        rec = "Server send : " + rec
        if err = websocket.JSON.Send(ws, rec); err != nil {
            fmt.Println("Can't send")
            break
        }
    }
}

func Home(w http.ResponseWriter, r *http.Request) {
    fmt.Println("hello~")
    fmt.Fprintf(w, "hello")
}
