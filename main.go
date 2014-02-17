package main

import(
    "github.com/gorilla/websocket"
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
    "html/template"
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
    http.HandleFunc("/connws/", ConnWs)
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

func ConnWs(w http.ResponseWriter, r *http.Request) {
    ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
    if _, ok := err.(websocket.HandshakeError); ok {
        http.Error(w, "Not a websocket handshake", 400)
        return
    } else if err != nil {
        log.Println(err)
        return
    }

    rec := map[string] interface{}{}
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
        rec["Test"] = "tt"
        fmt.Println(rec)
        if err = ws.WriteJSON(&rec); err != nil {
            fmt.Println("Write : " + err.Error())
            return
        }
    }

}

func Home(w http.ResponseWriter, r *http.Request) {
    t := template.Must(template.New("connWebsocket").Parse(tmpl))
    v := map[string]interface{}{
        "host" : "54.250.138.78:9090",
    }
    t.Execute(w, v)
}


const tmpl = `
<!DOCTYPE html>
<head>
    <title>Test~</title>
</head>
<body>
</body>
<script type="text/javascript">
    ws = new WebSocket("ws://{{.host}}/connws/");
    ws.onopen = function() {
        console.log("[onopen] connect ws uri.");
        var data = {
            "Enabled" : "true"
        };
        ws.send(JSON.stringify(data));
    }
    ws.onmessage = function(e) {
        var res = JSON.parse(e.data);
        console.log(res);
    }
    ws.onclose = function(e) {
        console.log("[onclose] connection closed (" + e.code + ")");
        delete ws;
    }
    ws.onerror = function (e) {
        console.log("[onerror] error!");
    }
</script>
</html>
`
