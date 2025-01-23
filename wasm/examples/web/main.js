// import case A: bstates is in "node_modules" directory
import { load, example } from 'bstates';

// import case B: bstates is not in "node_modules" directory
// import { load } from '../../index.js';
// import { example } from '../../example.js';

(async () => {
    try {
        // Load bstates
        // case A: bstates is in "node_modules" directory
        const bs = await load("./node_modules/bstates/dist");

        // case B: bstates is not in "node_modules" directory
        // const bs = await load("../../dist");

        // Run the example automatically
        example(bs);

        // Expose `bs` globally for testing from the console
        window.bstates = bs;
    } catch (error) {
        console.error("An error occurred:", error);
    }
})();
