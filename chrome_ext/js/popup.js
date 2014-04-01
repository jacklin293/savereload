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
var doConnectCount = 0;

document.addEventListener('DOMContentLoaded', function () {
    // init get connection status
    getConnStatus();

    // checkbox event
    document.getElementById('connChecked').addEventListener('click', connChecked);
    document.getElementById('watchFolderChecked').addEventListener('click', updateWatchFolderChecked);
    document.getElementById('sassChecked').addEventListener('click', updateSassChecked);
})

function connChecked(e) {
    document.getElementById('connChecked').className = "hide";
    document.getElementById('connectLoading').className = "";
    var connChecked = document.getElementById('connChecked').checked;

    var port = document.getElementById('port').value;
    var url = document.getElementById('url').value;
    url = _extractUrl(url);

    // Do websocket connect
    chrome.runtime.sendMessage({
            "wsAction"    : "connChecked",
            "connChecked" : connChecked,
            "url"         : url,
            "port"        : port
    });

    // Check websocket connection whether establish or not.
    if (connChecked) {
        var timer = setInterval(function() {
            doConnectCount++;
            console.log("Count : " + doConnectCount);
            getConnStatus();
            if (wsIsEstablished || doConnectCount > 3) {
                document.getElementById('connChecked').className = "";
                document.getElementById('connectLoading').className = "hide";
                clearInterval(timer);
                doConnectCount = 0;
                // If connect successfully, watch folder by default.
                document.getElementById("watchFolderChecked").checked = true;
                updateWatchFolderChecked();
            }
        }, 1000);
    } else {
        setTimeout(function () {
            getConnStatus();
            document.getElementById('connectLoading').className = "hide";
            document.getElementById("connChecked").checked = false;
            document.getElementById('connChecked').className = "";
        }, 1000);
    }
}

function updateWatchFolderChecked(e) {
    var watchFolderChecked = document.getElementById('watchFolderChecked').checked;
    chrome.runtime.sendMessage({
        "wsAction"              : "updateWatchFolderChecked",
        "watchFolderChecked"    : watchFolderChecked
    });
    getConnStatus();
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
    var timer = setInterval(function() {
        doSassCount++;
        console.log("Count : " + doSassCount);
        getConnStatus();
        document.getElementById('sassChecked').checked = true;
        if (sassServerReply || doSassCount > 3) {
            document.getElementById('sassLoading').className = "hide";
            clearInterval(timer);
            doSassCount = 0;
            showSassStatus();
        }
    }, 1000);

}

function showWatchFolderStatus(isEnabled) {
    document.getElementById('watchFolderChecked').checked = isEnabled;
    // Reload switch status
    if (isEnabled) {
        document.getElementById("watchFolderStatus").innerHTML = "enabled";
        document.getElementById("watchFolderStatus").className = "success";
    } else {
        document.getElementById("watchFolderStatus").innerHTML = "disabled";
        document.getElementById("watchFolderStatus").className = "fail";
    }
}

// Show connection status
function showConnStatus() {
    document.getElementById("connChecked").checked = wsIsEstablished;

    // If websocket connection is established, disabled url.
    if (wsIsEstablished) {
        // url & port
        document.getElementById("url").disabled = true;
        document.getElementById("port").disabled = true;

        // websocket connection status
        document.getElementById("wsIsEstablished").innerHTML = "connect";
        document.getElementById("wsIsEstablished").className = "success";

        // watch folder
        document.getElementById("watchFolderCheckedContainer").className = "";
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

        // watch folder
        document.getElementById("watchFolderCheckedContainer").className = "hide";
    }
}

function getConnStatus() {
    chrome.runtime.sendMessage({wsAction: "getConnStatus"}, function(response) {
        // url & port
        document.getElementById('url').value = (response.url == "") ? defaultHost : response.url;
        document.getElementById('port').value = (response.port == "") ? defaultPort : response.port;

        // websocket status
        wsIsEstablished = response.wsIsEstablished;
        showConnStatus();

        // watch folder status
        watchFolderStatus = response.watchFolderStatus;
        showWatchFolderStatus(watchFolderStatus);

        // sass status
        document.getElementById('sassChecked').checked = response.sassChecked;
        document.getElementById('sassSrc').value = response.sassSrc;
        document.getElementById('sassDes').value = response.sassDes;
        sassServerReply = response.sassServerReply;
        sassSrcError = response.sassSrcError;
        sassDesError = response.sassDesError;
        showSassStatus();
    });
}

function showSassStatus() {
    // Show sass options
    var sassOptions = document.getElementsByClassName("sassOptions");
    count = sassOptions.length;

    // Connect successfully.
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

    // Check path that is existent.
    if (sassSrcError != "") {
        document.getElementById('sassSrcError').className = "";
    } else {
        document.getElementById('sassSrcError').className = "hide";
    }
    if (sassDesError != "") {
        document.getElementById('sassDesError').className = "";
    } else {
        document.getElementById('sassDesError').className = "hide";
    }
}

function _extractUrl(url) {
    if (url.indexOf("http://") == 0) {
         return url.substr(7);
    }
    return url;
}