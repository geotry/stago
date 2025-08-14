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

export const RenderStatistics = {
  avg: 0,
  min: 0,
  max: 0,
  fps: 0,
  render: 0,
  bytesDownAvg: 0,
  bytesUpAvg: 0,
};

let frame = 0;
const frameTimes = Array(60 * 5);
const frameDownByteLength = Array(60 * 5).fill(0);
const frameUpByteLength = Array(60 * 5).fill(0);

/**
 * Configure webgl context and setup a new websocket connection to render frames.
 * 
 * @param {Awaited<ReturnType<webgl.createContext>>} ctx
 * @param {boolean} reconnecting
 * @returns 
 */
export const createRenderWebSocket = (ctx, reconnecting) => {
  closeRenderWebSocket();

  frame = 0;
  let messageIndex = 0;
  let time = new Date().getTime();

  return new Promise((resolve) => {
    // Use "render" protocol to receive frames in binary data
    const ws = new WebSocket(endpoint, ["render"]);
    ws.binaryType = "arraybuffer";

    ws.onmessage = event => {
      messageIndex++;

      frameTimes[frame % frameTimes.length] = new Date().getTime();
      frameDownByteLength[frame % frameDownByteLength.length] = event.data.byteLength;
      frame++;

      const beforeRender = new Date().getTime();

      ctx.handle(event.data, frame);
      ctx.render(frame);

      const now = new Date().getTime();

      // Compute statistics every second
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

        const totalDownBytes = frameDownByteLength.reduce((prev, curr) => prev + curr, 0);
        const avgDownBytes = totalDownBytes / frameDownByteLength.filter(f => f !== undefined).length;
        const totalUpBytes = frameUpByteLength.reduce((prev, curr) => prev + curr, 0);
        const avgUpBytes = totalUpBytes / frameUpByteLength.filter(f => f !== undefined).length;

        frameUpByteLength.fill(0);

        RenderStatistics.min = min;
        RenderStatistics.max = max;
        RenderStatistics.avg = avg;
        RenderStatistics.fps = avg > 0 ? 1000 / avg : 0;
        RenderStatistics.render = now - beforeRender;
        RenderStatistics.bytesDownAvg = avgDownBytes;
        RenderStatistics.bytesUpAvg = avgUpBytes;
      }
    };

    ws.onclose = event => {
      if (event.code === 1006) {
        setTimeout(() => resolve(createRenderWebSocket(ctx, true)), 1000);
      } else {
        console.log("[ws:render] connection closed");
      }
    };

    ws.onopen = event => {
      if (reconnecting) {
        ctx.reset();
      }
      renderWs = ws;
      console.log("[ws:render] connection open");
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
 * @param {boolean} reconnecting
 * @returns {Promise<WebSocket>}
 */
export const createInputWebsocket = (reconnecting) => {
  if (inputWs?.readyState === WebSocket.OPEN || inputWs?.readyState === WebSocket.CONNECTING) {
    inputWs.close(1000);
  }

  return new Promise((resolve) => {
    const ws = new WebSocket(endpoint, ["input"]);

    ws.onopen = event => {
      console.log("[ws:input] connection open");
      inputWs = ws;
      resolve(ws);
    };

    ws.onclose = (event) => {
      if (event.code === 1006) {
        setTimeout(() => resolve(createInputWebsocket(true)), 1000);
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
    const message = JSON.stringify({ session_id: sessionId, device: 0, x, y, deltaX: dx, deltaY: dy });
    frameUpByteLength[frame % frameDownByteLength.length] += message.length;
    inputWs.send(message);
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
    const message = JSON.stringify({ session_id: sessionId, device: 0, pressed: true, x, y });
    frameUpByteLength[frame % frameDownByteLength.length] += message.length;
    inputWs.send(message);
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
    const message = JSON.stringify({ session_id: sessionId, device: 0, pressed: true, released: true, x, y });
    frameUpByteLength[frame % frameDownByteLength.length] += message.length;
    inputWs.send(message);
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
    const message = JSON.stringify({ session_id: sessionId, device: 0, x, y, scrolled: true, delta: deltaY });
    frameUpByteLength[frame % frameDownByteLength.length] += message.length;
    inputWs.send(message);
  }
};

/**
 * 
 * @param {string} key
 */
export const sendKeydownEvent = (key) => {
  // TODO: Send binary data
  if (inputWs?.readyState === WebSocket.OPEN && renderWs?.readyState === WebSocket.OPEN) {
    const message = JSON.stringify({ session_id: sessionId, device: 1, code: key, pressed: true });
    frameUpByteLength[frame % frameDownByteLength.length] += message.length;
    inputWs.send(message);
  }
};

/**
 * 
 * @param {string} key
 */
export const sendKeyupEvent = (key) => {
  // TODO: Send binary data
  if (inputWs?.readyState === WebSocket.OPEN && renderWs?.readyState === WebSocket.OPEN) {
    const message = JSON.stringify({ session_id: sessionId, device: 1, code: key, pressed: false });
    frameUpByteLength[frame % frameDownByteLength.length] += message.length;
    inputWs.send(message);
  }
};