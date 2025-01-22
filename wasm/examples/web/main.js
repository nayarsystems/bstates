import { load as bstatesLoad } from '../../index.js';
// import { load as bstatesLoad } from 'bstates'; // This is the import for the published package
import { example } from '../../example.js';

(async () => {
    try {
        // Load bstates
        const bs = await bstatesLoad("../../dist");

        // Run the example automatically
        example(bs);

        // Expose `bs` globally for testing from the console
        window.bstates = bs;
    } catch (error) {
        console.error("An error occurred:", error);
    }
})();
