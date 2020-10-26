window.onload = function() {
let home = document.getElementById('home');
let close = document.getElementById('close');

home.onclick = function(element) {
    var newURL = "http://localhost:8080/";
     chrome.tabs.create({ url: newURL });
}

close.onclick = function(e) {
    window.close();
}
}