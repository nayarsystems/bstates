import { load, example } from 'bstates';
try {
    const bs = await load();
    example(bs);
} catch (error) {
    console.error("An error occurred:", error);
}