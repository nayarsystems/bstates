import fs from 'fs/promises';
import path from 'path';
import { fileURLToPath } from 'url';

// Load wasm_exec.js
import '../wasm_exec.js';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const wasmFilePath = path.join(__dirname, '../bstates.wasm');

export async function init() {
    const go = new Go(); // `Go` comes from wasm_exec.js

    // Asynchronously load and initialize the WASM module
    const wasmBuffer = await fs.readFile(wasmFilePath);
    const wasmModule = await WebAssembly.compile(wasmBuffer);
    const wasmInstance = await WebAssembly.instantiate(wasmModule, go.importObject);

    // Run the Go WASM environment
    go.run(wasmInstance);

    // Validate exports
    const requiredExports = ['createStateQueue', 'decodeStates', 'encodeStates'];
    requiredExports.forEach((fn) => {
        if (typeof global[fn] !== 'function') {
            throw new Error(`WASM module did not export '${fn}'.`);
        }
    });

    return {
        createStateQueue: global.createStateQueue,
        decodeStates: global.decodeStates,
        encodeStates: global.encodeStates,
    };
}
