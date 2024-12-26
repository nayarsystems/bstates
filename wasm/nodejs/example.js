const fs = require('fs');
const path = require('path');

// Load wasm_exec.js
require('../wasm_exec.js');

const go = new Go(); // `Go` comes from wasm_exec.js
const wasmFilePath = path.join(__dirname, '../bstates.wasm');

(async () => {
    const wasmBuffer = fs.readFileSync(wasmFilePath);
    const { instance } = await WebAssembly.instantiate(wasmBuffer, go.importObject);
    go.run(instance);
    main();
})();


function main() {
    console.log("Creating queue...");
    const schema = 
    {
        "version": "2.0",
        "encoderPipeline": "t:z",
        "decodedFields": [
            {
                "name": "message",
                "decoder": "BufferToString",
                "params": {
                    "from": "message_buf"
                }
            }
        ],
        "fields": 
            [
                {
                    "name": "3bitUnsignedNumTest", 
                    "type": "uint", 
                    "size": 2
                },
                {
                    "name": "boolTest",
                    "type": "bool"
                },
                {
                    "name": "4bitSignedNumTest",
                    "type": "int",
                    "size": 4,
                },
                {
                    "name": "message_buf",
                    "type": "buffer",
                    "size": 1024,
                }
            ]
    }

    const res = global.createStateQueue(schema);
    if (res.e != null) {
        console.log("Error creating queue:", res.e);
        return;
    }
    queue = res.d;
    
    console.log("Queue created:", queue);

    console.log("Pushing states to queue...");
    queue.push({"3bitUnsignedNumTest": 1, "boolTest": true, "4bitSignedNumTest": -7, "message_buf": "Hello, World 1!"});
    queue.push({"3bitUnsignedNumTest": 2, "boolTest": false,"4bitSignedNumTest": -6, "message_buf": "Hello, World 2!"});
    queue.push({"3bitUnsignedNumTest": 3, "boolTest": true, "4bitSignedNumTest": -5, "message_buf": "Hello, World 3!"});

    
    console.log("Trying to push invalid state...");
    console.log(queue.push({"3bitUnsignedNumTest": 4, "THIS_FIELD_DOES_NOT_EXIST": true, "4bitSignedNumTest": -4, "message_buf": "Hello, World 4!"}));
    
    console.log("Queue size:", queue.size());

    states = queue.toArray();
    
    console.log("------------------------");
    states.forEach(state => {
        delete state.message_buf;
        console.log(state);
    });
    console.log("------------------------");

    // Pop states from queue in a for and print them
    while ((state = queue.pop()) != null) {
        delete state.message_buf;
        console.log("Pop: ", state);
    }

    console.log("Queue size:", queue.size());
}