import { exit } from 'process';

// import case A: bstates is in "node_modules" directory
import { load, example } from 'bstates';

// import case B: bstates is not in "node_modules" directory
//import { load as bstatesLoad } from '../../index.js';
//import { example } from '../../example.js';

try {
    const bs = await load();
    example(bs);
} catch (error) {
    console.error("An error occurred:", error);
    exit(1);
}