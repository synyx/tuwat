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
                let lastRefresh = Date.parse(document.querySelector("#last_refresh").dateTime);
                if (Date.now() - lastRefresh > 90000) {
                    console.log("Force reloading, last refresh too old: " + (Date.now() - lastRefresh))
                    location.reload();
                } else if (Date.now() - lastRefresh > 70000) {
                    console.log("Reloading, last refresh too old: " + (Date.now() - lastRefresh))
                    location.reload(); // TODO: make a request to get a partial
                }
                conn.timerId = setTimeout(reload, 10000);
            }
        }, 10000);
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

let eventSource = new URLSearchParams(window.location.search).get("eventSource");
if (eventSource == null) {
    if (window["WebSocket"]) {
        eventSource = "websocket";
    } else if (window["EventSource"]) {
        eventSource = "sse";
    }
}

let fallback = new FallbackConn();
let conn = null;
switch (eventSource) {
    case "websocket": {
        const wsUrl = ((window.location.protocol === "https:") ? "wss://" : "ws://") + window.location.host + "/ws/alerts";
        conn = new WebSocketConn(wsUrl);
        break;
    }
    case "sse": {
        conn = new SSEConn(window.location.protocol + "//" + window.location.host + "/sse/alerts");
        break;
    }
    default:
        conn = fallback;
}
fallback.connect();
conn.connect();

document.addEventListener("DOMContentLoaded", function () {
    console.log('Adding handler for manual disconnect.');
    const csEl = document.getElementById('connection-state');
    csEl.addEventListener("change", function () {
        if (this.checked) {
            console.log('Manually connecting stream.');
            conn.reconnect();
            fallback.reconnect();
        } else {
            console.log('Manually disconnecting stream.');
            conn.disconnect();
            fallback.disconnect();
        }
    });
});
