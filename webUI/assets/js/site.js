if ('serviceWorker' in navigator) {
    window.addEventListener('load', function () {
        navigator.serviceWorker.register('swDummy.js').then(function (registration) {
            // console.log('ServiceWorker registration successful with scope: ', registration.scope);
        }, function (err) {
            // console.log('ServiceWorker registration failed: ', err);
        });
    });
}


function buzzer(door, csrfToken) {
    var errorBox = document.getElementById('errorBox');
    var successBox = document.getElementById('successBox');
    errorBox.style.display = 'none';
    successBox.style.display = 'none';

    var dooButtons = document.getElementById("doorButtons");
    dooButtons.className += " sending";
    var buttons = document.getElementsByTagName("button");
    for (var i = 0; i < buttons.length; i++) {
        buttons.item(i).disabled = true;
    }

    sendRequest('/buzzer?door=' + door, csrfToken,
        function (serverError, response) {
            dooButtons.className = dooButtons.className.replace('sending', '');
            var buttons = document.getElementsByTagName('button');
            for (var i = 0; i < buttons.length; i++) {
                buttons.item(i).disabled = false;
            }

            if (serverError) {
                errorBox.style.display = 'block';
            } else {
                if (response === 'OK') {
                    successBox.style.display = 'block';
                } else if (response === 'LOGIN') {
                    window.location = '/login';
                } else {
                    errorBox.style.display = 'block';
                }
            }
            hideBoxWithTimeout();
        }
    );
}

var timeoutHandle;

function hideBoxWithTimeout() {
    if (timeoutHandle) {
        window.clearTimeout(timeoutHandle);
    }
    timeoutHandle = window.setTimeout(function () {
        document.getElementById('errorBox').style.display = 'none';
        document.getElementById('successBox').style.display = 'none';
    }, 3000);
}

function sendRequest(url, csrfToken, callback) {
    var xhr = new XMLHttpRequest();
    xhr.open('PUT', url);
    xhr.setRequestHeader('X-CSRF-TOKEN', csrfToken);
    xhr.send(null);
    xhr.onreadystatechange = function () {
        var DONE = 4; // readyState 4 means the request is done.
        var OK = 200; // status 200 is a successful return.
        if (xhr.readyState === DONE) {
            if (xhr.status === OK) {
                callback(false, xhr.responseText);
            } else {
                callback(true, xhr);
            }
        }
    };
}