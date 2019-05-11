if ('serviceWorker' in navigator) {
    window.addEventListener('load', function () {
        navigator.serviceWorker.register('swDummy.js').then(function (registration) {
            // console.log('ServiceWorker registration successful with scope: ', registration.scope);
        }, function (err) {
            // console.log('ServiceWorker registration failed: ', err);
        });
    });
}

function removeClass(node, className) {
    node.className = node.className.replace(className, '');
}
function addClass(node, className) {
    node.className += " " + className;
}

function buzzer(door, csrfToken) {
    var errorSnack = document.getElementById('errorSnack');
    var infoSnack = document.getElementById('infoSnack');

    removeClass(errorSnack, "show");
    removeClass(infoSnack, "show");

    var dooButtons = document.getElementById("doorButtons");
    addClass(dooButtons, "sending");
    var buttons = document.getElementsByTagName("button");
    for (var i = 0; i < buttons.length; i++) {
        buttons.item(i).disabled = true;
    }

    sendRequest('/buzzer?door=' + door, csrfToken,
        function (serverError, response) {
            removeClass(dooButtons, "sending");
            var buttons = document.getElementsByTagName('button');
            for (var i = 0; i < buttons.length; i++) {
                if (buttons.item(i).onclick) {
                    buttons.item(i).disabled = false;
                }
            }

            if (serverError) {
                addClass(errorSnack, "show");
            } else {
                if (response === 'OK') {
                    addClass(infoSnack, "show");
                } else if (response === 'LOGIN') {
                    window.location = '/login';
                } else {
                    addClass(errorSnack, "show");
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
        removeClass(document.getElementById('errorSnack'), "show");
        removeClass(document.getElementById('infoSnack'), "show");
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


function showLoginWaiting() {
    var lBtn = document.getElementById('loginButton');
    addClass(lBtn, 'waiting');
    lBtn.disabled = true;
}