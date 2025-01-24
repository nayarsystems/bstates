(async () => {
    try {
        const { load, example } = await import('bstates'); // Dynamic load of ES Module
        console.log("Loading bstates (CommonJS)...");
        const bs = await load();
        example(bs);
    } catch (error) {
        console.error("An error occurred:", error);
        process.exit(1);
    }
})();
