import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

// Input: Path to the .wasm file
const wasmFilePath = path.resolve(
    path.dirname(fileURLToPath(import.meta.url)),
    './dist/bstates.wasm'
);

// Output: Path to the generated JS module
const outputFilePath = path.resolve(
    path.dirname(fileURLToPath(import.meta.url)),
    './dist/bstates-wasm.js'
);

// Read the .wasm file and convert it to Base64
const wasmBuffer = fs.readFileSync(wasmFilePath);
const wasmBase64 = wasmBuffer.toString('base64');

// Generate a JavaScript module that exports the WASM content as an ArrayBuffer
const jsModuleContent = `
/**
 * This module exports the WASM binary as an ArrayBuffer.
 * It is automatically generated from the original .wasm file.
 */

// Function to decode Base64 into an ArrayBuffer (works in both Node.js and browsers)
function base64ToArrayBuffer(base64) {
    const binaryString = typeof atob === 'function'
        ? atob(base64) // Browser
        : Buffer.from(base64, 'base64').toString('binary'); // Node.js

    const len = binaryString.length;
    const buffer = new ArrayBuffer(len);
    const view = new Uint8Array(buffer);
    for (let i = 0; i < len; i++) {
        view[i] = binaryString.charCodeAt(i);
    }
    return buffer;
}

export const wasmBinary = base64ToArrayBuffer('${wasmBase64}');
`;

fs.writeFileSync(outputFilePath, jsModuleContent);

console.log(`WASM module generated at: ${outputFilePath}`);
