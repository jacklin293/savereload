function getDomainFromUrl(url){
      var host = "null" ;
      if ( typeof url == "undefined" || null == url)
          url = window.location.href;
      var regex = /.*\:\/\/([^\/]*).*/ ;
      var match = url.match(regex);
      if ( typeof match != "undefined " && null != match)
          host = match[1 ];
      return host;
}
function checkForValidUrl(tabId, changeInfo, tab) {
      if (getDomainFromUrl(tab.url).toLowerCase()=="www.google.com.tw" ){
          chrome.pageAction.show(tabId);
     }
}
// Called when the user clicks on the browser action.
/*
chrome.browserAction.onClicked.addListener(function(tab) {
  // No tabs or host permissions needed!
  console.log('Turning ' + tab.url + ' red!');
  chrome.tabs.executeScript({
    code: 'document.body.style.backgroundColor="red"'
  });
});
var oldOnload = window.onload || function () {};
window.onload = function ()
{
    console.log("Chrome extension start...");
    var websocketEnabled = false;
    //chrome.tabs.onUpdated.addListener(checkForValidUrl);
    oldOnload();
    ws = new WebSocket("ws://127.0.0.1:9090/connws/");
    ws.onopen = function() {
        console.log("[onopen] connect ws uri.");
        var data = {
            "Enabled" : "true"
        };
        ws.send(JSON.stringify(data));
    }
    ws.onmessage = function(e) {
        var res = JSON.parse(e.data);
        console.log(res);
    }
    ws.onclose = function(e) {
        console.log("[onclose] connection closed (" + e.code + ")");
        delete ws;
    }
    ws.onerror = function (e) {
        console.log("[onerror] error!");
    }
}*/
