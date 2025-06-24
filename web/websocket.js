const webgl = require("./webgl.js");

const endpoint = "ws://localhost:9090";

const sessionId = (Math.random() + 1).toString(36).substring(7);

/**
 * @type {WebSocket}
 */
let renderWs, inputWs;

const options = {
  session_id: sessionId,
  fps: 60,
};

const RenderHeaders = {
  Palette: 1,
  Texture: 2,
};

export const RenderStatistics = {
  avg: 0,
  min: 0,
  max: 0,
  fps: 0,
  bytesAvg: 0,
};

const frameTimes = Array(60 * 5);
const frameByteLength = Array(60 * 5);

/**
 * Configure webgl context and setup a new websocket connection to render frames.
 * 
 * @param {ReturnType<webgl.createContext>} ctx
 * @returns 
 */
export const createRenderWebSocket = (ctx) => {
  closeRenderWebSocket();

  let frames = 0;
  let messageIndex = 0;
  let time = new Date().getTime();

  return new Promise((resolve) => {
    // Use "render" protocol to receive frames in binary data
    const ws = new WebSocket(endpoint, ["render"]);
    ws.binaryType = "arraybuffer";

    ws.onmessage = event => {
      messageIndex++;

      const header = messageIndex;

      // Headers
      if (header === RenderHeaders.Palette) {
        ctx.createPalette(event.data);
        return;
      }

      if (header === RenderHeaders.Texture) {
        ctx.createTexture(event.data);
        return;
      }

      frameTimes[frames % frameTimes.length] = new Date().getTime();
      frameByteLength[frames % frameByteLength.length] = event.data.byteLength;
      frames++;

      ctx.render(event.data);

      const now = new Date().getTime();
      if ((now - time) > 1000) {
        time = now;
        const max = frameTimes.reduce((prev, curr, index) => {
          const prevTime = frameTimes[index === 0 ? frameTimes.length - 1 : index - 1];
          if (prevTime === undefined || prevTime > curr) {
            return prev;
          }
          return curr - prevTime > prev ? curr - prevTime : prev;
        }, 0);
        const min = frameTimes.reduce((prev, curr, index) => {
          const prevTime = frameTimes[index === 0 ? frameTimes.length - 1 : index - 1];
          if (prevTime === undefined || prevTime > curr) {
            return prev;
          }
          return curr - prevTime < prev ? curr - prevTime : prev;
        }, Number.MAX_SAFE_INTEGER);
        const totalDuration = frameTimes.reduce((prev, curr, index) => {
          const prevTime = frameTimes[index === 0 ? frameTimes.length - 1 : index - 1];
          if (prevTime === undefined || prevTime > curr) {
            return prev;
          }
          return prev + curr - prevTime;
        }, 0);

        const avg = totalDuration / frameTimes.filter(f => f !== undefined).length;

        const totalBytes = frameByteLength.reduce((prev, curr) => prev + curr, 0);
        const avgBytes = totalBytes / frameByteLength.filter(f => f !== undefined).length;

        RenderStatistics.min = min;
        RenderStatistics.max = max;
        RenderStatistics.avg = avg;
        RenderStatistics.fps = avg > 0 ? 1000 / avg : 0;
        RenderStatistics.bytesAvg = avgBytes;
      }
    };

    ws.onclose = event => {
      if (event.code === 1006) {
        // Try to reconnect
        setTimeout(() => resolve(createRenderWebSocket(ctx)), 1000);
      } else {
        console.log("[ws:render] connection closed");
      }
    };

    ws.onopen = event => {
      renderWs = ws;
      console.log("[ws:render] connection open", event);
      sendRenderOptions();
      resolve();
    };
  });
};

export const closeRenderWebSocket = () => {
  if (renderWs?.readyState === WebSocket.OPEN || renderWs?.readyState === WebSocket.CONNECTING) {
    renderWs.close(1000);
  }
};

/**
 * 
 * @param {Record<string, unknown>} newOptions 
 */
export const sendRenderOptions = (newOptions) => {
  if (newOptions) {
    Object.entries(newOptions).forEach(([key, value]) => {
      options[key] = value;
    });
  }
  if (renderWs?.readyState === WebSocket.OPEN) {
    renderWs.send(JSON.stringify(options));
  }
};

/**
 * @returns {Promise<WebSocket>}
 */
export const createInputWebsocket = () => {
  if (inputWs) {
    inputWs.close();
  }

  return new Promise((resolve) => {
    const ws = new WebSocket(endpoint, ["input"]);

    ws.onopen = event => {
      console.log("[ws:input] connection open", event);
      inputWs = ws;
      resolve(ws);
    };

    ws.onclose = (event) => {
      if (event.code === 1006) {
        // Try to reconnect
        setTimeout(() => resolve(createInputWebsocket()), 1000);
      } else {
        console.log("[ws:input] connection closed");
      }
    };
  })
};

/**
 * 
 * @param {number} x
 * @param {number} y 
 */
export const sendMouseMoveEvent = (x, y, dx, dy) => {
  // TODO: Send binary data
  if (inputWs?.readyState === WebSocket.OPEN && renderWs?.readyState === WebSocket.OPEN) {
    inputWs.send(JSON.stringify({ session_id: sessionId, device: 0, x, y, deltaX: dx, deltaY: dy }));
  }
};

/**
 * 
 * @param {number} x
 * @param {number} y 
 */
export const sendMouseDragEvent = (x, y) => {
  // TODO: Send binary data
  if (inputWs?.readyState === WebSocket.OPEN && renderWs?.readyState === WebSocket.OPEN) {
    inputWs.send(JSON.stringify({ session_id: sessionId, device: 0, pressed: true, x, y }));
  }
};

/**
 * 
 * @param {number} x
 * @param {number} y 
 */
export const sendMouseClickEvent = (x, y) => {
  // TODO: Send binary data
  if (inputWs?.readyState === WebSocket.OPEN && renderWs?.readyState === WebSocket.OPEN) {
    inputWs.send(JSON.stringify({ session_id: sessionId, device: 0, pressed: true, released: true, x, y }));
  }
};

/**
 * 
 * @param {number} x
 * @param {number} y 
 * @param {number} deltaY
 */
export const sendScrollEvent = (x, y, deltaY) => {
  // TODO: Send binary data
  if (inputWs?.readyState === WebSocket.OPEN && renderWs?.readyState === WebSocket.OPEN) {
    inputWs.send(JSON.stringify({ session_id: sessionId, device: 0, x, y, scrolled: true, delta: deltaY }));
  }
};


/**
 * 
 * @param {string} key
 */
export const sendKeydownEvent = (key) => {
  // TODO: Send binary data
  if (inputWs?.readyState === WebSocket.OPEN && renderWs?.readyState === WebSocket.OPEN) {
    inputWs.send(JSON.stringify({ session_id: sessionId, device: 1, code: key, pressed: true }));
  }
};

/**
 * 
 * @param {string} key
 */
export const sendKeyupEvent = (key) => {
  // TODO: Send binary data
  if (inputWs?.readyState === WebSocket.OPEN && renderWs?.readyState === WebSocket.OPEN) {
    inputWs.send(JSON.stringify({ session_id: sessionId, device: 1, code: key, pressed: false }));
  }
};