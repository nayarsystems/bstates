import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

// Load wasm_exec.js
import '../wasm_exec.js';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const wasmFilePath = path.join(__dirname, '../bstates.wasm');

const go = new Go(); // `Go` comes from wasm_exec.js

// Synchronously load and initialize the WASM module
const wasmBuffer = fs.readFileSync(wasmFilePath);
const wasmModule = new WebAssembly.Module(wasmBuffer);
const wasmInstance = new WebAssembly.Instance(wasmModule, go.importObject);

// Run the Go WASM environment
go.run(wasmInstance);

// Ensure functions are available globally
if (typeof global.createStateQueue !== 'function') {
    throw new Error("WASM module did not export 'createStateQueue'.");
}

// Export the WASM functions directly
export const createStateQueue = global.createStateQueue;
export const decodeStates = global.decodeStates;
export const encodeStates = global.encodeStates;
