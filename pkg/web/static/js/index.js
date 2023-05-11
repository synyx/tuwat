import * as Turbo from '@hotwired/turbo';
import ReconnectingWebSocket from 'reconnecting-websocket';

class SSEConn {
    constructor(socketUrl) {
        this.active = true;
        this.socketUrl = socketUrl;
    }
    connect() {
        this.socket = new EventSource(this.socketUrl);
    }
    disconnect() {
        this.active = false;
        if (this.socket) {
            Turbo.disconnectStreamSource(this.socket);
            this.socket.close();
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
class WebSocketConn {
    constructor(socketUrl) {
        this.socketUrl = socketUrl;
    }
    connect() {
        let conn = this;
        let socket = new ReconnectingWebSocket(this.socketUrl);
        socket.addEventListener("error", function (ev) {
            console.error('Socket encountered error: ', ev.message, 'Closing socket');
        })
        socket.addEventListener("open", function () {
            console.log('Socket is connected.  Enabling turbo.');
            Turbo.connectStreamSource(socket);
        })
        this.socket = socket;
    }
    disconnect() {
        if (this.socket) {
            Turbo.disconnectStreamSource(this.socket);
            this.socket.close(3001, "Human disconnect");
            this.socket = null;
        }
    }
    reconnect() {
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
} else if (window["EventSource"]) {
    conn = new SSEConn(window.location.protocol + "//" + window.location.host + "/sse/alerts");
} else {
    conn = new FallbackConn();
}
conn.connect();

document.addEventListener("DOMContentLoaded", function() {
    console.log('Adding handler for manual disconnect.');
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
