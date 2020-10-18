let done = document.getElementById('done');
let home = document.getElementById('home');

home.onclick = function(element) {
    var newURL = "http://localhost:8080/";
    chrome.tabs.create({ url: newURL });
}

done.onclick = function(element) {
    chrome.tabs.query({active: true, currentWindow: true}, function(tabs) {
        chrome.tabs.executeScript(
            tabs[0].id,
            {
                file: "info.js",
                runAt: "document_idle"
            },
            function(p) {
                p = JSON.parse(String(p))
                var tags = document.getElementById('tags').value;
                var tag_arr = tags.split(";").filter(item => item);
                p.tags = tag_arr;
                fetch("http://localhost:8080/attempt", {
                    method: "post",
                    body:  JSON.stringify(p)
                }).then(response => {
                    if (response.status == 200) {
                        alert("Succeed!")
                    } else {
                        alert("Failed!")
                    }
                });
            })
  });
}