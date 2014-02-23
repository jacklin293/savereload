// Show connection status
function showConnStatus(connStatus) {
    document.getElementById("connStatusContainer").style.display = "block";
    document.getElementById("connStatus").innerHTML = connStatus;
    if (connStatus == "connect") {
        document.getElementById("connStatus").className = "success";
        document.getElementById("url").disabled = true;
    } else {
        document.getElementById("connStatus").className = "fail";
        document.getElementById("url").disabled = false;
    }
}

function doConnect(e) {
    var enabled = document.getElementById('switch').checked;
    var url = document.getElementById('url').value;
    url = extractUrl(url);
    chrome.runtime.sendMessage({"wsAction": "doConnect","wsConn": enabled, "url": url}, function(response) {
        //showConnStatus(response.connStatus);
    });
    setInterval(getConnStatus(), 2000);
}

document.addEventListener('DOMContentLoaded', function () {
    // init get connection status
    getConnStatus();

    // checkbox event
    document.getElementById('switch').addEventListener('click', doConnect);
})

function getConnStatus() {
    chrome.runtime.sendMessage({wsAction: "getConnStatus"}, function(response) {
        document.getElementById('switch').checked = response.wsEnabled;
        document.getElementById('url').value = response.url;
        showConnStatus(response.connStatus);
    });
}

function extractUrl(url) {
    if (url.indexOf("http://") == 0) {
         return url.substr(7);
    }
    return url;
}

