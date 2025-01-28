import { load, example } from './node_modules/bstates/index.js';

(async () => {
    try {
        // Load bstates passing the directory where the bstates.wasm file is hosted.
         const bs = await load("./node_modules/bstates/dist");

        // Run the example automatically
        example(bs);

        // Expose `bs` globally for testing from the console
        window.bstates = bs;
    } catch (error) {
        console.error("An error occurred:", error);
    }
})();
