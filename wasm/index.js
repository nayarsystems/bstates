// isNode is a boolean that indicates whether the code is running in Node.js or in a browser.
const isNode = typeof process !== 'undefined' && process.versions != null && process.versions.node != null;
import './dist/wasm_exec.js'

// This function is used to load the WASM module
export async function load(customWasmFilesPathPrefix = null) {
    let wasmPath;
    let fs, path, fileURLToPath;
    let wasmBuffer;
    if (isNode) {
        // Node.js: Use 'fs/promises' to read files
        fs = (await import('fs')).promises;
        path = (await import('path')).default;
        fileURLToPath = (await import('url')).fileURLToPath;
        const __filename = fileURLToPath(import.meta.url);
        const __dirname = path.dirname(__filename);

        let wasmBinPathPrefix = customWasmFilesPathPrefix ||  path.join(__dirname, 'dist');
        let wasmBinPath = path.join(wasmBinPathPrefix, 'bstates.wasm');
        wasmBuffer = await fs.readFile(wasmBinPath);
    } else {
        // Browser: Use relative URL to load .wasm file
        let wasmBinPathPrefix = customWasmFilesPathPrefix || './dist';
        let wasmBinPath = wasmBinPathPrefix + '/bstates.wasm';
        // Download .wasm file in the browser
        const response = await fetch(wasmBinPath);
        if (!response.ok) {
            throw new Error(`Error loading WASM file at '${wasmBinPath}': ${response.statusText}`);
        }
        wasmBuffer = await response.arrayBuffer();
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
