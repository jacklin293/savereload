package main

import(
    "github.com/howeyc/fsnotify"
    "log"
    "flag"
    "strings"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "time"
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
        fmt.Printf("Command (%v) has too few args\n", cmd)
        os.Exit(0)
    }
    cmdPtr := exec.Command(splitCmd[0], splitCmd[1:]...)
    cmdPtr.Stdout = os.Stdout
    cmdPtr.Stderr = os.Stderr
    err := cmdPtr.Run()
    if err != nil {
        fmt.Printf("Command failed! %s\n", err.Error())
        os.Exit(0)
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
            prevActionTime = time.Now().Second()
            fmt.Println(prevActionTime-time.Now().Second())
            // Ignore some file extension
            if CheckIgnoreExt(filepath.Ext(ev.Name), strings.Split(args.IgnoreExt, "|")) {
                continue
            }
            log.Println("event:", ev)
            if args.Cmd != "" {
                RunCommand(args.Cmd)
            }
        case err := <-watcher.Error:
            log.Println("error:", err)
            os.Exit(0)
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

    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        fmt.Println(err.Error())
        os.Exit(0)
    }

    done := make(chan bool)
    go args.watch_directory(watcher)

    err = watcher.Watch(args.Path)
    if err != nil {
        fmt.Println(err.Error())
        os.Exit(0)
    }
    <-done
}
