console.log("=== background starting ===");

// Global variable
var ws = new WebSocket("ws://54.250.138.78:9090/connws/");
var wsEnabled = false;

ws.onopen = function() {
  console.log("[onopen] connect ws uri.");
  var data = {
    "Action" : "requireConnect"
  };
  ws.send(JSON.stringify(data));
  wsEnabled = true;
}

ws.onmessage = function(e) {
    var res = JSON.parse(e.data);
    if (wsEnabled && res["Action"] == "doReload") {
      pageReload();
    }
}

ws.onclose = function(e) {
    console.log("[onclose] connection closed (" + e.code + ")");
    delete ws;
    wsEnabled = false;
}

ws.onerror = function (e) {
    console.log("[onerror] error!");
    wsEnabled = false;
}

function wsConnect() {
    
    if (ws.readyState == 1) {
        wsEnabled = true;
    } else {
        wsEnabled = false;
    }
    /* 
    Check ws whether initial or not.
    ==================================
    CONNECTING 0 The connection is not yet open.
    OPEN  1 The connection is open and ready to communicate.
    CLOSING 2 The connection is in the process of closing.
    CLOSED  3 The connection is closed or couldn't be opened.
    
    if (ws.readyState != 1) {
        wsEnabled = false;
        return;
    }*/
}

function wsDisconnect() {
    var data = {
        "Action" : "requireDisconnect"
    };
    ws.send(JSON.stringify(data));
    wsEnabled = false;
}

function pageReload() {
    chrome.tabs.query({active: true, currentWindow: true}, function(tabs) {
        lastTabID = tabs[0].id;
        chrome.tabs.sendMessage(lastTabID, "Tab " + lastTabID + " do reload.");
    });
}

chrome.runtime.onMessage.addListener(
  function(request, sender, sendResponse) {
    if (request.wsAction == "getConnStatus") {
       var connStatus = (wsEnabled) ? "connect" : "disconnect";
       changeBrowserActionIcon();
       sendResponse({"wsEnabled": wsEnabled, "connStatus": connStatus});
    }

    if (request.wsAction == "checkboxEvent") {
        if (request.wsConn) {
          wsConnect();
        } else {
          wsDisconnect();
        }
        var connStatus = (wsEnabled) ? "connect" : "disconnect";
        changeBrowserActionIcon();
        sendResponse({"connStatus": connStatus});
    }
});

// Change browser action icon
function changeBrowserActionIcon() {
    if (wsEnabled) {
        chrome.browserAction.setIcon({
              path : "img/browser_action_icon_enabled_19.png"
          });
    } else {
        chrome.browserAction.setIcon({
              path : "img/browser_action_icon_disabled_19.png"
          });
    }
}
