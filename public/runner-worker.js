// runner-worker.js — runs the GoMark in-browser Go runner off the main thread.
//
// Loading the WebAssembly module and executing snippets both happen here, so a
// runaway snippet (e.g. an infinite loop) can be stopped by terminating the
// worker from the page without ever freezing the UI thread.
//
// Protocol (postMessage):
//   in:  { type: 'init' }                      -> { type: 'ready' } | { type: 'error', error }
//   in:  { type: 'run', id, source }           -> { type: 'result', id, result } | { type: 'result', id, error }

self.importScripts('/wasm_exec.js');

var readyPromise = null;

// init loads and starts the wasm module once, resolving when window.runGo (set by
// the module) becomes callable. The module parks on select{} and never returns,
// so go.run is not awaited; a startup trap is surfaced as a rejection.
function init() {
  if (readyPromise) {
    return readyPromise;
  }
  readyPromise = new Promise(function (resolve, reject) {
    try {
      var go = new self.Go();
      fetch('/runner.wasm').then(function (resp) {
        if (!resp.ok) {
          throw new Error('fetch runner.wasm: ' + resp.status);
        }
        return resp.arrayBuffer();
      }).then(function (buf) {
        return WebAssembly.instantiate(buf, go.importObject);
      }).then(function (result) {
        go.run(result.instance).catch(reject);
        var startedAt = Date.now();
        (function waitReady() {
          if (typeof self.runGo === 'function') { resolve(); return; }
          if (Date.now() - startedAt > 8000) { reject(new Error('runtime did not initialize')); return; }
          setTimeout(waitReady, 15);
        })();
      }).catch(reject);
    } catch (err) {
      reject(err);
    }
  });
  return readyPromise;
}

function errorText(err) {
  return String((err && err.message) || err);
}

self.onmessage = function (event) {
  var msg = event.data || {};

  if (msg.type === 'init') {
    init().then(function () {
      self.postMessage({ type: 'ready' });
    }, function (err) {
      self.postMessage({ type: 'error', error: errorText(err) });
    });
    return;
  }

  if (msg.type === 'run') {
    init().then(function () {
      var result;
      try {
        result = self.runGo(msg.source || '') || {};
      } catch (err) {
        self.postMessage({ type: 'result', id: msg.id, error: errorText(err) });
        return;
      }
      self.postMessage({ type: 'result', id: msg.id, result: result });
    }, function (err) {
      self.postMessage({ type: 'result', id: msg.id, error: errorText(err) });
    });
  }
};
