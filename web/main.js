const canvas = document.querySelector("#canvas");

const worker = new Worker(new URL('./worker.js', import.meta.url));

const fpsSlider = document.querySelector("#fps-slider");
const resolutionSelector = document.querySelector("#resolution-selector");
const stopBtn = document.querySelector("#stop-btn");

/**
 * @returns {number}
 */
const getFps = () => Number(fpsSlider.value);

/**
 * @returns {number[]}
 */
const getResolution = () => resolutionSelector.value.split("x").map(Number);

const updateCamera = () => {
  worker.postMessage(["setCamera",
    Number(document.querySelector(".camera-control[data-option=near]").value),
    Number(document.querySelector(".camera-control[data-option=far]").value),
    Number(document.querySelector(".camera-control[data-option=fov]").value),
  ]);
};

window.addEventListener("resize", () => {
  resizeCanvasToDisplaySize(canvas, true);
});

const main = () => {
  // Setup canvas
  resizeCanvasToDisplaySize(canvas, false);
  canvas.style.cursor = "none";
  canvas.oncontextmenu = (e) => {
    e.preventDefault();
    e.stopPropagation();
  };

  // Setup controls
  fpsSlider.addEventListener("change", event => {
    event.target.nextElementSibling.value = event.target.value + " fps";
    worker.postMessage(["setFps", Number(event.target.value)]);
  });

  resolutionSelector.addEventListener("change", (event) => {
    const [x, y] = event.target.value.split("x").map(Number);
    worker.postMessage(["setSize", x, y]);
  });

  stopBtn.addEventListener("click", (event) => {
    const isStopped = event.target.dataset.stopped === "true";
    if (!isStopped) {
      event.target.dataset.stopped = "true";
      event.target.textContent = "Play";
      canvas.style.cursor = "auto";
      canvas.oncontextmenu = undefined;
      worker.postMessage(["setFps", -1]);
    } else {
      event.target.dataset.stopped = "false";
      event.target.textContent = "Stop";
      canvas.style.cursor = "none";
      canvas.oncontextmenu = (e) => {
        e.preventDefault();
        e.stopPropagation();
      };
      worker.postMessage(["setFps", getFps()]);
    }
  });

  for (const slider of document.querySelectorAll(".camera-control")) {
    slider.addEventListener("input", () => updateCamera());
  }

  // Setup worker
  worker.onmessage = (e) => {
    switch (e.data[0]) {
      case "ready": {
        onReady();
        break;
      }

      case "stats": {
        const fps = e.data[4].toFixed(0);
        const frameTimeMin = e.data[1].toFixed(0);
        const frameTimeMax = e.data[2].toFixed(0);
        const frameTimeAvg = e.data[3].toFixed(0);
        let bandwidthAvg = e.data[4] * e.data[5];
        if (bandwidthAvg < 1024) {
          bandwidthAvg = `${bandwidthAvg.toFixed(0)}b`;
        } else if (bandwidthAvg < 1024 * 1024) {
          bandwidthAvg = `${(bandwidthAvg / 1024).toFixed(0)}kb`;
        } else if (bandwidthAvg < 1024 * 1024 * 1024) {
          bandwidthAvg = `${(bandwidthAvg / 1024 / 1024).toFixed(2)}mb`;
        }
        document.querySelector("#fps").textContent = `fps: ${fps} | min: ${frameTimeMin}ms | max: ${frameTimeMax}ms | avg: ${frameTimeAvg}ms | ↓ ${bandwidthAvg}/s | ↑ n.a/s`;
        break;
      }
    }
  };

  const offscreen = canvas.transferControlToOffscreen();

  worker.postMessage(["setup", offscreen], [offscreen]);

  setInterval(() => {
    worker.postMessage(["stats"]);
  }, 1000);
};

const onReady = () => {
  // Start rendering
  worker.postMessage(["start", getFps(), canvas.width, canvas.height]);

  // Setup DOM events
  setupDOMInputEvents(worker);
  updateCamera();
};

/**
 * 
 * @param {Worker} worker 
 */
const setupDOMInputEvents = (worker) => {
  window.addEventListener("click", e => {
    if (e.target === canvas) {
      const mx = e.offsetX / canvas.clientWidth;
      const my = e.offsetY / canvas.clientHeight;
      worker.postMessage(["mouse_click", mx, my]);
    }
  });

  let drag = false;
  window.addEventListener("mousedown", e => {
    if (e.target === canvas) {
      drag = true;
    }
  });
  window.addEventListener("mouseup", e => {
    drag = false;
  });

  const mouseMouveDebounce = 10;
  let lastMouseEventSent = new Date().getTime();
  let lastMouseMoveX = 0;
  let lastMouseMoveY = 0;
  window.addEventListener("mousemove", e => {
    if (e.target === canvas) {
      const now = new Date().getTime();
      const mx = e.offsetX / canvas.clientWidth;
      const my = e.offsetY / canvas.clientHeight;
      const dx = mx - lastMouseMoveX;
      const dy = my - lastMouseMoveY;

      lastMouseMoveX = mx;
      lastMouseMoveY = mx;

      if (now - lastMouseEventSent > mouseMouveDebounce) {
        if (drag) {
          worker.postMessage(["mouse_drag", mx, my, dx, dy]);
        } else {
          worker.postMessage(["mouse_move", mx, my, dx, dy]);
        }
        lastMouseEventSent = now;
      }
    }
  });

  window.addEventListener("wheel", (e) => {
    if (e.target === canvas) {
      const mx = e.offsetX / canvas.width;
      const my = e.offsetY / canvas.height;
      worker.postMessage(["scroll", mx, my, e.deltaY]);
    }
  });

  // Keyboard inputs
  const filterKey = (code) =>
    code.startsWith("Key") ||
    code.startsWith("Arrow") ||
    code.startsWith("Shift") ||
    code.startsWith("Control") ||
    code.startsWith("Tab") ||
    code.startsWith("Space") ||
    code.startsWith("Escape") ||
    code === "Delete"

  window.addEventListener("keydown", e => {
    if (filterKey(e.code)) {
      worker.postMessage(["keydown", e.code]);
    }
  });
  window.addEventListener("keyup", e => {
    if (filterKey(e.code)) {
      worker.postMessage(["keyup", e.code]);
    }
  });
};

/**
 * 
 * @param {HTMLCanvasElement} canvas
 * @param {boolean} offscreen
 * @returns 
 */
const resizeCanvasToDisplaySize = (canvas, offscreen) => {
  // Lookup the size the browser is displaying the canvas in CSS pixels.
  const displayWidth = canvas.clientWidth;
  const displayHeight = canvas.clientHeight;

  // Check if the canvas is not the same size.
  const needResize = canvas.width !== displayWidth || canvas.height !== displayHeight;

  if (needResize) {
    // Make the canvas the same size
    if (offscreen) {
      worker.postMessage(["setSize", displayWidth, displayHeight]);
    } else {
      canvas.width = displayWidth;
      canvas.height = displayHeight;
    }
  }

  return needResize;
};

main();

