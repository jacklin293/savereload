var wsIsEstablished = false;
var defaultHost = "127.0.0.1";
var defaultPort = "9112";
var doConnectCount = 0;

// Show connection status
function showConnStatus(wsIsEstablished, connSwitchStatus) {
    // If websocket connection is established, disabled url.
    if (wsIsEstablished) {
        // url & port
        document.getElementById("url").disabled = true;
        document.getElementById("port").disabled = true;

        // websocket connection status
        document.getElementById("wsIsEstablished").innerHTML = "connect";
        document.getElementById("wsIsEstablished").className = "success";

        // Reload switch status
        if (connSwitchStatus) {
            document.getElementById("connSwitchStatus").innerHTML = "enabled";
            document.getElementById("connSwitchStatus").className = "success";
        } else {
            document.getElementById("connSwitchStatus").innerHTML = "disabled";
            document.getElementById("connSwitchStatus").className = "fail";
        }
    } else {
        // url & port
        document.getElementById("url").disabled = false;
        document.getElementById("port").disabled = false;

        // websocket connection status
        document.getElementById("wsIsEstablished").innerHTML = "disconnect";
        document.getElementById("wsIsEstablished").className = "fail";

        // Reload switch status
        document.getElementById("connSwitchStatus").innerHTML = "disabled";
        document.getElementById("connSwitchStatus").className = "fail";
    }
}

function doConnect(e) {
    document.getElementById('loading').className = "";

    var switchStatus = document.getElementById('switch').checked;
    var port = document.getElementById('port').value;
    var url = document.getElementById('url').value;
    url = extractUrl(url);

    // Do websocket connect
    chrome.runtime.sendMessage({
            "wsAction"    : "doConnect",
            "wsConn"      : switchStatus,
            "url"         : url,
            "port"        : port
    });

    // Check websocket connection whether establish or not.
    if (switchStatus) {
        var timer = setInterval(function() {
            doConnectCount++;
            console.log("timeout : " + doConnectCount);
            getConnStatus();
            if (wsIsEstablished || doConnectCount > 5) {
                document.getElementById('loading').className = "hide";
                clearInterval(timer);
                doConnectCount = 0;
            }
        }, 1000);
    } else {
        getConnStatus();
        document.getElementById('loading').className = "hide";
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
    document.getElementById('switch').addEventListener('click', doConnect);
    document.getElementById('close').addEventListener('click', doClose);
    document.getElementById('sassChecked').addEventListener('click', updateSassChecked);
})

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
        showConnStatus(response.wsIsEstablished, response.connSwitchStatus);
        wsIsEstablished = response.wsIsEstablished;

        // sass status
        showSassStatus();

        // show close websocket button
        if (wsIsEstablished) {
            document.getElementById('switch').checked = response.connSwitchStatus;
            document.getElementById('close').className = "";
        } else {
            document.getElementById('switch').checked = false;
            document.getElementById('close').className = "hide";
        }
    });
    return wsIsEstablished;
}

function extractUrl(url) {
    if (url.indexOf("http://") == 0) {
         return url.substr(7);
    }
    return url;
}

function showSassStatus() {
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

function updateSassChecked() {
    // Send sass status to background
    var sassChecked = document.getElementById('sassChecked').checked;
    var sassSrc = document.getElementById('sassSrc').value;
    var sassDes = document.getElementById('sassDes').value;
    chrome.runtime.sendMessage({
        "wsAction"    : "updateSassChecked",
        "sassChecked" : sassChecked,
        "sassSrc"     : sassSrc,
        "sassDes"     : sassDes
    });
}
