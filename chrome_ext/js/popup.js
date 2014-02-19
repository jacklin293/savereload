// init
document.addEventListener('DOMContentLoaded', function () {
    chrome.runtime.sendMessage({wsAction: "getStatus"}, function(response) {
        document.getElementById('switch').checked = response.wsEnabled;
    });    
})

function clickHandler(e) {
    var enabled = document.getElementById('switch').checked;
    chrome.runtime.sendMessage({wsConn: enabled}, function(response) {
        alert(response.connStatus);
    });

}

document.addEventListener('DOMContentLoaded', function () {
    document.getElementById('switch').addEventListener('click', clickHandler);
})

