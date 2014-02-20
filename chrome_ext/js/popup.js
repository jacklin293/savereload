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

// init get connection status
document.addEventListener('DOMContentLoaded', function () {
    chrome.runtime.sendMessage({wsAction: "getConnStatus"}, function(response) {
        document.getElementById('switch').checked = response.wsEnabled;
        showConnStatus(response.connStatus);
    });    
})

function clickHandler(e) {
    var enabled = document.getElementById('switch').checked;
    var url = document.getElementById('url').value;
    chrome.runtime.sendMessage({"wsAction": "checkboxEvent","wsConn": enabled, "url": url}, function(response) {
        showConnStatus(response.connStatus);
    });

}

document.addEventListener('DOMContentLoaded', function () {
    document.getElementById('switch').addEventListener('click', clickHandler);
})

