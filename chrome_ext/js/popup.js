// Default connection setting
var wsIsEstablished = false,
    watchCheckedStatus = false,
    defaultHost = "127.0.0.1",
    defaultPort = "9112";

// Sass
var doSassCount= 0,
    sassChecked = false,
    sassServerReply = false,
    sassSrcError = "",
    sassDesError = "";

// Backgound check status count
var doConnectCount = 0,
    doWatchFolderCount = 0;

// Show connection status
function showConnStatus() {
    // If websocket connection is established, disabled url.
    if (wsIsEstablished) {
        // url & port
        document.getElementById("url").disabled = true;
        document.getElementById("port").disabled = true;

        // websocket connection status
        document.getElementById("wsIsEstablished").innerHTML = "connect";
        document.getElementById("wsIsEstablished").className = "success";

        // Reload switch status
        if (watchFolderStatus) {
            document.getElementById("watchFolderStatus").innerHTML = "enabled";
            document.getElementById("watchFolderStatus").className = "success";
        } else {
            document.getElementById("watchFolderStatus").innerHTML = "disabled";
            document.getElementById("watchFolderStatus").className = "fail";
        }
    } else {
        // url & port
        document.getElementById("url").disabled = false;
        document.getElementById("port").disabled = false;

        // websocket connection status
        document.getElementById("wsIsEstablished").innerHTML = "disconnect";
        document.getElementById("wsIsEstablished").className = "fail";

        // Reload switch status
        document.getElementById("watchFolderStatus").innerHTML = "disabled";
        document.getElementById("watchFolderStatus").className = "fail";
    }
}

function getConnStatus() {
    chrome.runtime.sendMessage({wsAction: "getConnStatus"}, function(response) {
        // sass status
        document.getElementById('sassChecked').checked = response.sassChecked;
        document.getElementById('sassSrc').value = response.sassSrc;
        document.getElementById('sassDes').value = response.sassDes;

        // url & port
        document.getElementById('url').value = (response.url == "") ? defaultHost : response.url;
        document.getElementById('port').value = (response.port == "") ? defaultPort : response.port;

        // websocket status
        wsIsEstablished = response.wsIsEstablished;
        watchFolderStatus = response.watchFolderStatus;
        showConnStatus();

        // sass status
        showSassOptions();

        // show close websocket button
        if (wsIsEstablished) {
            document.getElementById('close').className = "";
        } else {
            document.getElementById('close').className = "hide";
        }
    });
    return wsIsEstablished;
}

function doWatchFolder(e) {
    document.getElementById('watchFolderLoading').className = "";
    var watchFolderChecked = document.getElementById('watchFolderChecked').checked;

    chrome.runtime.sendMessage({
        "wsAction"      : "watchFolder",
        "watchFolder"   : watchFolderChecked
    });
    if (watchFolderChecked) {
        var timer = setInterval(function() {
            doWatchFolderCount++;
            console.log("Count : " + doWatchFolderCount);
            getConnStatus();
            document.getElementById("watchFolderChecked").checked = true;
            if (watchFolderStatus || doWatchFolderCount > 5) {
                document.getElementById('connectLoading').className = "hide";
                document.getElementById("watchFolderChecked").checked = false;
                clearInterval(timer);
                doWatchFolder = 0;
            }
        }, 1000);
    } else {
        getConnStatus();
        document.getElementById('connectLoading').className = "hide";
        document.getElementById("watchFolderChecked").checked = false;
    }
}

function doConnect(e) {
    document.getElementById('connectLoading').className = "";
    var connChecked = document.getElementById('connChecked').checked;

    var port = document.getElementById('port').value;
    var url = document.getElementById('url').value;
    url = extractUrl(url);

    // Do websocket connect
    chrome.runtime.sendMessage({
            "wsAction"    : "doConnect",
            "url"         : url,
            "port"        : port
    });

    // Check websocket connection whether establish or not.
    if (connChecked) {
        var timer = setInterval(function() {
            doConnectCount++;
            console.log("Count : " + doConnectCount);
            getConnStatus();
            if (wsIsEstablished || doConnectCount > 5) {
                document.getElementById('connectLoading').className = "hide";
                document.getElementById("connChecked").checked = false;
                clearInterval(timer);
                doConnectCount = 0;
            }
        }, 1000);
    } else {
        getConnStatus();
        document.getElementById('connectLoading').className = "hide";
        document.getElementById("connChecked").checked = false;
    }
}

function doClose(e) {
    chrome.runtime.sendMessage({wsAction: "doClose"}, function(response) {
            console.log(response.log);
    });
    setTimeout(function () {
        getConnStatus();
    }, 1000);
}

document.addEventListener('DOMContentLoaded', function () {
    // init get connection status
    getConnStatus();

    // checkbox event
    document.getElementById('connChecked').addEventListener('click', doConnect);
    document.getElementById('watchFolderChecked').addEventListener('click', doWatchFolder);
    document.getElementById('close').addEventListener('click', doClose);
    document.getElementById('sassChecked').addEventListener('click', updateSassChecked);
})

function extractUrl(url) {
    if (url.indexOf("http://") == 0) {
         return url.substr(7);
    }
    return url;
}

function showSassOptions() {
    // Show sass options
    var sassOptions = document.getElementsByClassName("sassOptions");
    count = sassOptions.length;
    if (wsIsEstablished) {
        document.getElementById("sassCheckedContainer").style.display = "table-row";
        while (count--) {
            sassOptions[count].style.display = "table-row";
        }
    } else {
        document.getElementById("sassCheckedContainer").style.display = "none";
        while (count--) {
            sassOptions[count].style.display = "none";
        }
    }
}

function getSassStatus() {
    sassServerReply= true;
    sassChecked = false;
    sassSrcError = true;
    sassDesError = true;


    document.getElementById('sassChecked').checked = sassChecked;

    if (sassSrcError) {
        document.getElementById('sassSrcError').className = "";
    }
    if (sassDesError) {
        document.getElementById('sassDesError').className = "";
    }
}

function updateSassChecked() {
    document.getElementById('sassLoading').className = "";

    // Send sass status to background
    sassChecked = document.getElementById('sassChecked').checked;
    var sassSrc = document.getElementById('sassSrc').value;
    var sassDes = document.getElementById('sassDes').value;
    chrome.runtime.sendMessage({
        "wsAction"    : "updateSassChecked",
        "sassChecked" : sassChecked,
        "sassSrc"     : sassSrc,
        "sassDes"     : sassDes
    });

    // Check websocket connection whether establish or not.
    if (sassChecked) {
        var timer = setInterval(function() {
            doSassCount++;
            console.log("Count : " + doSassCount);
            getSassStatus();
            document.getElementById('switch').checked = true;
            if (sassServerReply || doSassCount > 5) {
                document.getElementById('sassLoading').className = "hide";
                clearInterval(timer);
                doSassCount = 0;
                document.getElementById('switch').checked = false;
            }
        }, 1000);
    } else {
        getSassStatus();
        document.getElementById('sassLoading').className = "hide";
    }
}
