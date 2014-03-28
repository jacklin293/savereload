var wsIsEstablished = false;

// Show connection status
function showConnStatus(wsIsEstablished, connSwitchStatus) {
    // If websocket connection is established, disabled url.
    if (wsIsEstablished) {
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

    // Compile sass
    compileSass = document.getElementById('compileSass').checked;
    var sassSrc = (compileSass) ? document.getElementById('sassSrc').value : "";
    var sassDes = (compileSass) ? document.getElementById('sassDes').value : "";
    var switchStatus = document.getElementById('switch').checked;
    var port = document.getElementById('port').value;
    var url = document.getElementById('url').value;
    url = extractUrl(url);

    // Do websocket connect
    chrome.runtime.sendMessage({
            "wsAction" : "doConnect",
            "wsConn"   : switchStatus,
            "url"      : url,
            "port"     : port,
            "sassSrc"  : sassSrc,
            "sassDes"  : sassDes
    });

    // Check websocket connection whether establish or not.
    if (switchStatus) {
        setTimeout(function() {
            getConnStatus();
            document.getElementById('loading').className = "hide";
        }, 1500);
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
    document.getElementById('compileSass').addEventListener('click', doCompileSass);
})

function getConnStatus() {
    chrome.runtime.sendMessage({wsAction: "getConnStatus"}, function(response) {
        document.getElementById('url').value = response.url;
        document.getElementById('port').value = response.port;
        showConnStatus(response.wsIsEstablished, response.connSwitchStatus);
        wsIsEstablished = response.wsIsEstablished;

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

function doCompileSass() {
    var elem1 = document.getElementById('sassOption1');
    var elem2 = document.getElementById('sassOption2');
    if (document.getElementById('compileSass').checked) {
        elem1.style.display = "table-row";
        elem2.style.display = "table-row";
    } else {
        elem1.style.display = "none";
        elem2.style.display = "none";
    }
}
