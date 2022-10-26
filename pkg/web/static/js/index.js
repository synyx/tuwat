import * as Turbo from '@hotwired/turbo';

class WebSocketConn {
    constructor(socketUrl) {
        this.active = true;
        this.socketUrl = socketUrl;
    }
    connect() {
        let conn = this;
        let socket = new WebSocket(this.socketUrl);
        socket.addEventListener("close", function (ev) {
            console.log('Socket is closed. Reconnect will be attempted in 1 second: ', ev.code, ev.reason);
            Turbo.disconnectStreamSource(socket);
            conn.socket = null;
            setTimeout(function () {
                if (conn.active) {
                    conn.reconnect();
                }
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
        this.socket = socket;
    }
    disconnect() {
        this.active = false;
        if (this.socket) {
            Turbo.disconnectStreamSource(this.socket);
            this.socket.close(3001, "Human disconnect");
            this.socket = null;
        }
    }
    reconnect() {
        this.active = true;
        if (!this.socket) {
            this.connect();
        }
    }
}
class FallbackConn {
    constructor() {
        this.active = true;
    }
    connect() {
        let conn = this;
        this.timerId = setTimeout(function reload() {
            if (conn.active) {
                location.reload();
                conn.timerId = setTimeout(reload, 60000);
            }
        }, 60000);
    }
    disconnect() {
        this.active = false;
        clearTimeout(this.timerId);
    }
    reconnect() {
        this.active = true;
        this.connect();
    }
}

let conn = null;
if (window["WebSocket"]) {
    const wsUrl = ((window.location.protocol === "https:") ? "wss://" : "ws://") + window.location.host + "/ws/alerts";
    conn = new WebSocketConn(wsUrl);
} else {
    conn = new FallbackConn();
}
conn.connect();

document.addEventListener("DOMContentLoaded", function() {
    const csEl = document.getElementById('connection-state');
    csEl.addEventListener("change", function () {
        if (this.checked) {
            console.log('Manually connecting stream.');
            conn.reconnect();
        } else {
            console.log('Manually disconnecting stream.');
            conn.disconnect();
        }
    });
});
