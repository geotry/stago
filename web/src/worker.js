const websocket = require("./websocket.js");
const webgl = require("./webgl.js");

/**
 * @type {ReturnType<webgl.createContext>}
 */
let webglContext;

/**
 * Rendering worker.
 *
 * @param {MessageEvent<[string, ...any[]]>} e 
 */
self.onmessage = (e) => {
  if (!Array.isArray(e.data)) {
    // if (e.data.type === "webpackOk") {

    // }
    return;
  }

  const [action, ...data] = e.data;

  switch (action) {
    case "setup": {
      webglContext = webgl.createContext(data[0]);
      (async () => {
        await websocket.createInputWebsocket();
      })()
        .then(() => {
          self.postMessage(["ready"]);
        })
        .catch(err => {
          console.error(err);
        });
      break;
    }

    case "start": {
      websocket.createRenderWebSocket(webglContext)
        .then(() => {
          websocket.sendRenderOptions({ fps: data[0], width: data[1], height: data[2] });
        });
      break;
    }

    case "setSize": {
      if (webglContext) {
        webglContext.resize(data[0], data[1]);
        websocket.sendRenderOptions({ width: data[0], height: data[1] });
      }
      break;
    }

    case "setCamera": {
      websocket.sendRenderOptions({ near: data[0], far: data[1], fov: data[2] });
      break;
    }

    case "setFps": {
      websocket.sendRenderOptions({ fps: data[0] });
      break;
    }

    case "stop": {
      websocket.closeRenderWebSocket();
      break;
    }

    case "mouse_move": {
      websocket.sendMouseMoveEvent(data[0], data[1], data[2], data[3]);
      break;
    }

    case "mouse_drag": {
      websocket.sendMouseDragEvent(data[0], data[1]);
      break;
    }

    case "mouse_click": {
      websocket.sendMouseClickEvent(data[0], data[1]);
      break;
    }

    case "scroll": {
      websocket.sendScrollEvent(data[0], data[1], data[2]);
      break;
    }

    case "keydown": {
      websocket.sendKeydownEvent(data[0]);
      break;
    }

    case "keyup": {
      websocket.sendKeyupEvent(data[0]);
      break;
    }

    case "stats": {
      self.postMessage(["stats", ...Object.values(websocket.RenderStatistics)]);
      break;
    }

    default:
      console.log(`[worker] received invalid message `, e.data);
  }
};