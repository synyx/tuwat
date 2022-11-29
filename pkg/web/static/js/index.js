import * as Turbo from '@hotwired/turbo';

function getSocket() {
    if (window["WebSocket"]) {
        const schema = ((window.location.protocol === "https:") ? "wss://" : "ws://");
        const wsUrl =  schema + window.location.host + "/ws/alerts";
        return new WebSocket(wsUrl);
    } else if (window["EventSource"]) {
        return new EventSource("/sse/alerts");
    }
    return null;
}

function connect() {
    const socket = getSocket();

    if (socket) {
        socket.addEventListener("close", function (ev) {
            console.log('Socket is closed. Reconnect will be attempted in 1 second.', ev.reason);
            Turbo.disconnectStreamSource(socket);
            setTimeout(function() {
                connect();
            }, 1000);
        })
        socket.addEventListener("error", function (ev) {
            console.error('Socket encountered error: ', ev.message, 'Closing socket');
            socket.close()
        })
        socket.addEventListener("open", function () {
            console.log('Socket is connected.  Enabling turbo.');
            Turbo.connectStreamSource(socket);
        })

        document.addEventListener("DOMContentLoaded", function() {
            const csEl = document.getElementById('connection-state');
            csEl.addEventListener("change", function (ev) {
                if (this.checked) {
                    console.log('Manually connecting stream.');
                    Turbo.disconnectStreamSource(socket);
                } else {
                    console.log('Manually disconnecting stream.');
                    Turbo.connectStreamSource(socket);
                }
            });
        });
    } else {
        setTimeout(function () {
            location.reload();
        }, 60000);
    }
};

connect();
