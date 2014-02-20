function pageReload(){
    location.reload();
}
chrome.runtime.onMessage.addListener(function(msg, _, sendResponse) {
    console.log(msg);
    pageReload();
});
