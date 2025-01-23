(async () => {
    try {
        // import case A: bstates is in "node_modules" directory
        const { load, example } = await import('bstates'); // Dynamic load of ES Module

        // import case B: bstates is not in "node_modules" directory
        // const { load } = await import('../../index.js'); // Dynamic load of ES Module
        // const { example } = await import('../../example.js'); // Dynamic load of ES Module

        console.log("Loading bstates (CommonJS)...");

        const bs = await load();
        example(bs);
    } catch (error) {
        console.error("An error occurred:", error);
        process.exit(1);
    }
})();
