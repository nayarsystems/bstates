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
    
    return {
        createStateQueue: globalScope.createStateQueue,
        decodeStates: globalScope.decodeStates,
        encodeStates: globalScope.encodeStates,
    };
}

export * from './example.js'; // Export the example function