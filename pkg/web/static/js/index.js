import * as Turbo from '@hotwired/turbo';
import ReconnectingWebSocket from 'reconnecting-websocket';
import { toggleFilteredStatus } from "./toggle-filtered";

class SSEConn {
    constructor(socketUrl) {
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
                let lastRefreshNode = document.querySelector("#last_refresh");
                let lastRefresh = Date.now();
                if (lastRefreshNode) {
                    lastRefresh = Date.parse(lastRefreshNode.dateTime);
                }

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
        const protocol = ((window.location.protocol === "https:") ? "wss://" : "ws://");
        const dashboard = window.location.pathname;
        const wsUrl = protocol + window.location.host + "/ws" + dashboard;
        conn = new WebSocketConn(wsUrl);
        break;
    }
    case "sse": {
        const protocol = window.location.protocol;
        const dashboard = window.location.pathname;
        const sseUrl = protocol + "//" + window.location.host + "/sse" + dashboard;
        conn = new SSEConn(sseUrl);
        break;
    }
    default:
        conn = fallback;
}
fallback.connect();
conn.connect();

document.addEventListener("DOMContentLoaded", function () {
    toggleFilteredStatus();

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
