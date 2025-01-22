import { exit } from 'process';
import { load as bstatesLoad } from '../../index.js';
// import { load as bstatesLoad } from 'bstates'; // This is the import for the published package
import { example } from '../../example.js';

try {
    const bs = await bstatesLoad();
    example(bs);
} catch (error) {
    console.error("An error occurred:", error);
    exit(1);
}