console.log("=== background starting ===");

var ws = new WebSocket("ws://54.250.138.78:9090/connws/");
var wsEnabled = false;
ws.onopen = function() {
    console.log("[onopen] connect ws uri.");
}
ws.onmessage = function(e) {
    var res = JSON.parse(e.data);
    if (res["Enabled"] == "true") {
      pageReload();
    }
}
ws.onclose = function(e) {
    console.log("[onclose] connection closed (" + e.code + ")");
    delete ws;
}
ws.onerror = function (e) {
    console.log("[onerror] error!");
}

function wsConnect() {
    var data = {
        "Enabled" : "true"
    };
    ws.send(JSON.stringify(data));
    wsEnabled = true;
}

function wsDisconnect() {
    var data = {
        "Enabled" : "false"
    };
    ws.send(JSON.stringify(data));
    wsEnabled = false;
}

function pageReload() {
    chrome.tabs.query({active: true, currentWindow: true}, function(tabs) {
      lastTabId = tabs[0].id;
      chrome.tabs.sendMessage(lastTabId, "Tab " + lastTabId + " do reload.");
    });
}

chrome.runtime.onMessage.addListener(
  function(request, sender, sendResponse) {
    console.log(sender.tab ?
              "from a content script:" + sender.tab.url :
              "from the extension");
    if (request.wsAction == "getStatus") {
       sendResponse({"wsEnabled": wsEnabled});
    }

    if (request.wsConn == true) {
      wsConnect();
      sendResponse({connStatus: "connect"});
    } else if (request.wsConn == false) {
      wsDisconnect();
      sendResponse({connStatus: "disconnect"});
    }
});
