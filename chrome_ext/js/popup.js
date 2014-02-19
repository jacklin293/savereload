console.log("=== popup starting ===");

$("#qq").click(function(){
    chrome.runtime.sendMessage({greeting: "hello"}, function(response) {
      console.log(response.farewell);
    });
});

