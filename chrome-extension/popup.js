let done = document.getElementById('done');
let home = document.getElementById('home');
let cancel = document.getElementById('cancel');

cancel.onclick = function(element) {
    window.close();
};

home.onclick = function(element) {
    var newURL = "http://localhost:8080/";
    chrome.tabs.create({ url: newURL });
};

chrome.tabs.query({active: true, currentWindow: true}, function(tabs) {
    chrome.tabs.executeScript(
        tabs[0].id,
        {
            file: "info.js",
            runAt: "document_idle"
        },
        function(p) {
            p = JSON.parse(String(p));
            document.getElementById("problem").innerHTML = p.id;
            done.onclick = function(element) {
                element.preventDefault();
                p.tags = document.getElementById("tags").value.split(",");
                fetch(
                    "http://localhost:8080/problem", 
                    {
                        method: "post",
                        body:  JSON.stringify(p)
                    }
                )
                .then(
                    response => {
                        if (response.status == 200) {
                            window.location.href = "succeed.html";
                        } else {
                            window.location.href = "fail.html";
                        }
                    }
                )
                .catch(err => {
                    alert(err);
                    window.location.href = "fail.html";
                });
            };
    })
})