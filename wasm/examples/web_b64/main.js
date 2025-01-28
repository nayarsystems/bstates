import { load, example } from './node_modules/bstates/index.js';

(async () => {
    try { 
        // Load bstates wasm
        const bs = await load();

        // Run the example automatically
        example(bs);

        // Expose `bs` globally for testing from the console
        window.bstates = bs;
    } catch (error) {
        console.error("An error occurred:", error);
    }
})();
