console.log("=== background starting ===");

// Global variable
var ws,
    wsIsEstablished     = false,
    watchFolderStatus  = false,
    url                 = "",
    port                = "",
    sassChecked         = false;
    sassSrc             = "",
    sassDes             = "",
    sassServerReply     = false,
    sassSrcError        = "",
    sassDesError        = ""

function wsConnect() {
    ws = new WebSocket("ws://" + url + ":" + port + "/connws/");

    ws.onopen = function() {
      console.log("[onopen] connect ws uri.");
      var data = {
        "Action" : "requireConnect"
      };
      ws.send(JSON.stringify(data));
      wsIsEstablished = true;
    }

    ws.onmessage = function(e) {
        var res = JSON.parse(e.data);
        if (watchFolderStatus && res["Action"] == "doReload") {
            pageReload();
        }
        if (wsIsEstablished && res["Action"] == "requireClose") {
            wsIsEstablished = false;
        }
        if (wsIsEstablished && res["Action"] == "updateSassChecked") {
            sassServerReply = true;
            sassSrcError = (res["SassSrcError"] == "undefined") ? "" : res["SassSrcError"];
            sassDesError = (res["SassDesError"] == "undefined") ? "" : res["SassDesError"];
        }
    }

    ws.onclose = function(e) {
        console.log("[onclose] connection closed (" + e.code + ")");
        delete ws;
        wsIsEstablished = false;
    }

    ws.onerror = function (e) {
        console.log("[onerror] error!");
        wsIsEstablished = false;
    }

    /*
    Check ws whether initial or not.
    ==================================
    CONNECTING 0 The connection is not yet open.
    OPEN  1 The connection is open and ready to communicate.
    CLOSING 2 The connection is in the process of closing.
    CLOSED  3 The connection is closed or couldn't be opened.

    if (ws.readyState != 1) {
        wsIsEstablished = false;
        return;
    }*/
}

function updateSassChecked() {
    var data = {
        "Action"        : "updateSassChecked",
        "SassChecked"   : sassChecked,
        "SassSrc"       : sassSrc,
        "SassDes"       : sassDes
    };
    ws.send(JSON.stringify(data));
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
        changeBrowserActionIcon();
        sendResponse({
            "wsIsEstablished"   : wsIsEstablished,
            "watchFolderStatus" : watchFolderStatus,
            "url"               : url,
            "port"              : port,
            "sassChecked"       : sassChecked,
            "sassSrc"           : sassSrc,
            "sassDes"           : sassDes,
            "sassServerReply"   : sassServerReply,
            "sassSrcError"      : sassSrcError,
            "sassDesError"      : sassDesError
        });
    }

    if (request.wsAction == "connChecked") {
        if (request.connChecked) {
            url         = request.url;
            port        = request.port;
            wsConnect();
            console.log("Do Close done.");
        } else {
            if (wsIsEstablished) {
                var data = {
                    "Action" : "requireClose"
                };
                ws.send(JSON.stringify(data));
                console.log("Do Close done.");
            } else {
                console.log("Websocket isn't established.");    
            }

        }
    }

    if (request.wsAction == "updateWatchFolderChecked") {
        watchFolderStatus = request.watchFolderChecked;
    }

    if (request.wsAction == "updateSassChecked") {
        if (wsIsEstablished) {
            sassChecked = request.sassChecked;
            sassSrc     = request.sassSrc;
            sassDes     = request.sassDes;
            updateSassChecked();
            console.log("Sass checked : " + request.sassChecked);
        } else {
            console.log("Websocket isn't established.");
        }
    }
});

// Change browser action icon
function changeBrowserActionIcon() {
    if (wsIsEstablished) {
        chrome.browserAction.setIcon({
              path : "img/browser_action_icon_enabled_19.png"
        });
    } else {
        chrome.browserAction.setIcon({
              path : "img/browser_action_icon_disabled_19.png"
        });
    }
}
