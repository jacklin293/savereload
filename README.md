# DEPRECATED

This is an experimental repo and no longer maintained.

Detect directory, reload current page when file changed.

> Ignore all hidden file. (ex: .git, .swp etc....)

# Install

    go get github.com/jex-lin/savereload
    go install github.com/jex-lin/savereload

# How to use

### [1] Install chrome extension

### [2] Execute the command

    web_develop_save_reload -p /tmp/test

### [3] Open chrome extension and type IP / domain

# Option

* `-p` : Watching path. Use `-p /tmp/test` to set watching target.
* `-P` : Listen port.
* `-r` : Watch subfolder under path. Default is recursive. Use `-r=false` to disable recursive watching.
* `-ig` : Ignore file extension changing. Example use `-ig="swp|git|swpx|conf"` to set ignorant list.

# Notice !!

* Port is 9112, not support other port so far. I will add this option as soon as possible.

# Install libsass

Check `g++` that has installed (ubuntu: `sudo apt-get install g++`)

Install libsass

    cd gosass/libsass
    make
    sudo make install

# Todo list

### Hot fix

* Server 回傳監聽的資料夾 path
* 忽略隱藏檔(.*)
* Dynamic add folder, but not monitor.  If folder and not exist then add to monitor.
* Sass compile

### Miscellaneous

* chrome extension  啟動 save reload 按鈕 分開為 連線及監聽按鈕要分開為兩個checkbox, 結束按鈕就不用了
* background.js 定時 update connection status => Can't do that?
* - extensions: .html .css .js .png .gif .jpg .php .php5 .py .rb .erb
* - excluding changes in: */.git/* */.svn/* */.hg/*



