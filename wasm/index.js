// isNode is a boolean that indicates whether the code is running in Node.js or in a browser.
const isNode = typeof process !== 'undefined' && process.versions != null && process.versions.node != null;
import './dist/wasm_exec.js'

// This function is used to load the WASM module
export async function load(customWasmFilesPathPrefix = null) {
    let wasmBuffer;
    if (isNode) {
        let fs, path;
        // Node.js: Use 'fs/promises' to read files
        fs = (await import('fs')).promises;
        path = (await import('path')).default;
        let fileURLToPath = (await import('url')).fileURLToPath;
        const filePath = fileURLToPath(import.meta.url);
        const __dirname = path.dirname(filePath);

        let wasmBinPathPrefix = customWasmFilesPathPrefix ||  path.join(__dirname, 'dist');
        let wasmBinPath = path.join(wasmBinPathPrefix, 'bstates.wasm');
        wasmBuffer = await fs.readFile(wasmBinPath);
    } else {
        if (customWasmFilesPathPrefix !== null) {
            // Browser: Use custom URL for the directory where the .wasm file is located
            let wasmBinPath = customWasmFilesPathPrefix + '/bstates.wasm';
            // Download .wasm file in the browser
            const response = await fetch(wasmBinPath);
            if (!response.ok) {
                throw new Error(`Error loading WASM file at '${wasmBinPath}': ${response.statusText}`);
            }
            wasmBuffer = await response.arrayBuffer();
        } else {
            // Load wasm importing the module that exports the WASM binary as an ArrayBuffer
            wasmBuffer = (await import('./dist/bstates-wasm.js')).wasmBinary;
        }
    }
    
    const go = new Go();

    const wasmModule = await WebAssembly.compile(wasmBuffer);
    const wasmInstance = await WebAssembly.instantiate(wasmModule, go.importObject);

    go.run(wasmInstance);

    // Validate exported functions
    const globalScope = isNode ? global : self; // Use `self` for browsers

    const requiredFunctions = ['createStateQueue', 'decodeStates', 'encodeStates'];
    for (const func of requiredFunctions) {
        if (typeof globalScope[func] !== 'function') {
            throw new Error(`function '${func}' is not available`);
        }
    }
    
    const createStateQueueRaw = globalScope.createStateQueue;
    const decodeStatesRaw = globalScope.decodeStates;
    const encodeStatesRaw = globalScope.encodeStates;

    const createStateQueue = wrapCreateStateQueue(createStateQueueRaw);
    const decodeStates = wrapGoFunc(decodeStatesRaw);
    const encodeStates = wrapGoFunc(encodeStatesRaw);

    return {
        createStateQueue,
        decodeStates,
        encodeStates,
    };    
}

function wrapCreateStateQueue(createStateQueueFn) {
  return (schema) => {
    const res = createStateQueueFn(schema);
  
    const queueObj = checkError(res);

    const wrappedQueue = {
      push: (...args) => checkError(queueObj.push(...args.map(normalizeBinaryArg))),
      pop: (...args) => checkError(queueObj.pop(...args.map(normalizeBinaryArg))),
      size: (...args) => checkError(queueObj.size(...args.map(normalizeBinaryArg))),
      toArray: (...args) => checkError(queueObj.toArray(...args.map(normalizeBinaryArg))),
      decode: (...args) => checkError(queueObj.decode(...args.map(normalizeBinaryArg))),
      encode: (...args) => checkError(queueObj.encode(...args.map(normalizeBinaryArg))),

      get schema() {
        return queueObj.schema;
      },
      get data() {
        return queueObj.data;
      },
    };

    return wrappedQueue;
  };
}
  
function checkError(res) {
  if (!res) {
    throw new Error("No result returned from Go function.");
  }
  if (res.e) {
    throw new Error(res.e);
  }
  return res.d;
}

function wrapGoFunc(goFunc) {
  return (...args) => {
    const normalizedArgs = args.map(normalizeBinaryArg);
    const result = goFunc(...normalizedArgs);
    return checkError(result); // lanza excepción si e != null
  };
}

function normalizeBinaryArg(arg) {
  // Node.js Buffer -> Uint8Array
  if (typeof Buffer !== 'undefined' && arg instanceof Buffer) {
    return new Uint8Array(arg.buffer, arg.byteOffset, arg.byteLength);
  }

  // ArrayBuffer -> Uint8Array
  if (arg instanceof ArrayBuffer) {
    return new Uint8Array(arg);
  }

  // if (
  //   arg instanceof Uint8Array ||
  //   arg instanceof Uint8ClampedArray ||
  //   arg instanceof Int8Array ||
  //   arg instanceof Uint16Array ||
  //   arg instanceof Int16Array ||
  //   arg instanceof Uint32Array ||
  //   arg instanceof Int32Array ||
  //   arg instanceof Float32Array ||
  //   arg instanceof Float64Array
  // ) {
      
  // }

  return arg;
}

export * from './example.js'; // Export the example function