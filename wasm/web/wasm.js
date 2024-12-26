'use strict';

if (!WebAssembly.instantiateStreaming) { // polyfill
    WebAssembly.instantiateStreaming = async (resp, importObject) => {
        const source = await (await resp).arrayBuffer();
        return await WebAssembly.instantiate(source, importObject);
    };
}

const go = new Go();
var mod, wasm;
WebAssembly.instantiateStreaming(fetch("bstates.wasm"), go.importObject).then((result) => {
    mod = result.module;
    wasm = result.instance;
    console.clear();
    go.run(wasm);
}).catch((err) => {
    console.error(err);
});