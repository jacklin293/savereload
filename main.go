package main

import(
    "github.com/howeyc/fsnotify"
    "log"
    "flag"
    "strings"
    "fmt"
    "os"
    "os/exec"
)

type Args struct {
    Folder string
    Cmd string
    Recurse bool
}

func runCommand(cmd string) {
    splitCmd := strings.Split(cmd, " ")
    if strings.TrimSpace(splitCmd[0]) == "" {
        fmt.Fprintf(os.Stdout, "Command (%v) has too few args\n", cmd)
        return
    }
    cmdPtr := exec.Command(splitCmd[0], splitCmd[1:]...)
    cmdPtr.Stdout = os.Stdout
    cmdPtr.Stderr = os.Stderr
    err := cmdPtr.Run()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Command failed! %v\n", err)
    }
}

func watch_directory(watcher *fsnotify.Watcher, args *Args) {
    for {
        select {
        case ev := <-watcher.Event:
            log.Println("event:", ev)
            if args.Cmd != "" {
                runCommand(args.Cmd)
            }
        case err := <-watcher.Error:
            log.Println("error:", err)
        }
    }
}

func main() {
    args := Args{}

	flag.StringVar(&args.Folder, "f", "", "The folder to watch")
	flag.StringVar(&args.Cmd, "c", "", "The command to run when the folder changes")
	flag.BoolVar(&args.Recurse, "r", true, "Controls whether the watcher should recurse into subdirectories")
	flag.Parse()

    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        log.Fatal(err)
    }

    done := make(chan bool)

    go watch_directory(watcher, &args)

    err = watcher.Watch(args.Folder)
    if err != nil {
        log.Fatal(err)
    }
    <-done

}
