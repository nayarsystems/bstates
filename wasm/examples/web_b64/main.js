import { load, example } from './node_modules/bstates/index.js';

(async () => {
    try {
        // Load bstates wasm file by importing the module that exports the WASM binary.
        // The module uses the Base64 string of the WASM binary to create an ArrayBuffer.
        // This means that the fetched resource is a JS module that occupies more space than the original WASM file.
        const bs = await load();

        // Run the example automatically
        example(bs);

        // Expose `bs` globally for testing from the console
        window.bstates = bs;
    } catch (error) {
        console.error("An error occurred:", error);
    }
})();
