# Introduction

Detect directory, reload current page when file changed.

# Install

    go get github.com/jex-lin/web_develop_save_reload

# How to use

### [1] Install chrome extension

### [2] Execute the command

    web_develop_save_reload -p /tmp/test

### [3] Open chrome extension and type IP / domain

# Default setting

* Port is 9112, not support other port so far.
* Watch directory in recursion, you can use `-r=false` to disable recursive watching.
* Ignore file extension `.swp` `.swpx`, you can set custom list (`-ig swp|git|swpx`) that you want to ignore the file extension.
