
function getData(url, async, callback) {
    $.ajax({
        url: url,
        async: async,
        type: "GET",
        success: callback,
        dataType: "json"
    })
}

function postData(url, async, data, callback) {
    $.ajax({
        url: url,
        async: async,
        type: "POST",
        contentType: "application/json",
        data: JSON.stringify(data),
        success: callback,
        dataType: "json"
    });
}




