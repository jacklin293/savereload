var wsIsEstablished;

// Show connection status
function showConnStatus(connStatus) {
    document.getElementById("connStatus").innerHTML = connStatus;
    if (connStatus == "connect") {
        document.getElementById("connStatus").className = "success";
    } else {
        document.getElementById("connStatus").className = "fail";
    }

    // If websocket connection is established, disabled url.
    if (wsIsEstablished) {
        document.getElementById("url").disabled = true;
    } else {
        document.getElementById("url").disabled = false;
    }
}

function doConnect(e) {
    document.getElementById('loading').className = "";

    var enabled = document.getElementById('switch').checked;
    var url = document.getElementById('url').value;
    url = extractUrl(url);

    // Do websocket connect
    chrome.runtime.sendMessage({"wsAction": "doConnect","wsConn": enabled, "url": url});

    // Check websocket connection whether establish or not.
    if (enabled) {
        setTimeout(function() {
            checkCount = 0;
            while (checkCount < 14) {
                if (getConnStatus()) {
                    break;
                }
                checkCount++;
            }
            document.getElementById('loading').className = "hide";
        }, 500);
    } else {
        getConnStatus();
        document.getElementById('loading').className = "hide";
    }
}

function doClose(e) {
    chrome.runtime.sendMessage({wsAction: "doClose"});
    setTimeout(function () {
        checkCount = 0;
        while (checkCount < 14) {
            if ( ! getConnStatus()) {
                break;
            }
            checkCount++;
        }
    }, 500);
}

document.addEventListener('DOMContentLoaded', function () {
    // init get connection status
    getConnStatus();

    // checkbox event
    document.getElementById('switch').addEventListener('click', doConnect);
    document.getElementById('close').addEventListener('click', doClose);
})

function getConnStatus() {
    chrome.runtime.sendMessage({wsAction: "getConnStatus"}, function(response) {
        document.getElementById('switch').checked = response.wsIsEstablished;
        wsIsEstablished = response.wsIsEstablished;
        showConnStatus(response.connStatus);

        // show close button
        if (wsIsEstablished) {
            if (response.url != "") {
                document.getElementById('close').className = "";
            } else {
                document.getElementById('close').className = "hide";
            }
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
