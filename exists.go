package main

import (
    "log"
    "os"
    "path/filepath"
)

func FileExists(path string) (bool) {
    fileInfo, err := os.Stat(path)
    if err != nil {
        // no such file or dir
        return false
    }
    if fileInfo.IsDir() {
        // it's a directory
        return false
    }
    // it's a file
    return true
}

func DirExists(path string) (bool) {
    fileInfo, err := os.Stat(path)
    if err != nil {
        // no such file or dir
        return false
    }
    if fileInfo.IsDir() {
        // it's a directory
        return true
    }
    // it's a file
    return false
}

func Subfolders(path string) (paths []string) {
    filepath.Walk(path, func(newPath string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        if info.IsDir() {
            name := info.Name()
            // skip folders that begin with a dot
            hidden := filepath.HasPrefix(name, ".") && name != "." && name != ".."
            if hidden {
                return filepath.SkipDir
            } else {
                paths = append(paths, newPath)
            }
        }
        return nil
    })
    return paths
}

func FilterExistPaths(paths []string) []string {
    var result []string
    for _, path := range paths {
        if FileExists(path) {
            result = append(result, path)
        } else {
            log.Printf("Invalid path: '%v'", path)
        }
    }
    return result
}
