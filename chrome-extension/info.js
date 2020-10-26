function info() {
    let diff = document.querySelector('div[diff]:not([diff=""])').innerHTML
    let fullTitle = document.querySelector('.css-v3d350').innerHTML
    var res = fullTitle.split(".")

    let url = window.location.href
    var arr = url.split("/").filter(item => item)

    var p = {
        id:     parseInt(res[0].trim()),
        title:  res[1],
        uname:  arr[arr.length - 1],
        difficulty: diff,
        url: url,   
    }

    return JSON.stringify(p)
}

info()