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

    // createStateQueue accepts an either an object or a string representing the schema
    res = createStateQueue(schema);
    if (res.e != null) {
        console.log("Error creating queue:", res.e);
        return;
    }
    queue = res.d;
    
    console.log("Queue created:", queue);

    console.log("Pushing states to queue...");
    queue.push({"3bitUnsignedNumTest": 1, "boolTest": true, "4bitSignedNumTest": -7, "message": "Hello, World 1!"});
    queue.push({"3bitUnsignedNumTest": 2, "boolTest": false,"4bitSignedNumTest": -6, "message": "Hello, World 2!"});
    queue.push({"3bitUnsignedNumTest": 3, "boolTest": true, "4bitSignedNumTest": -5, "message": "Hello, World 3!"});
    // raw "message_buf" field can be used instead of "message", which is a decoded field, to encode the value in the state
    queue.push({"3bitUnsignedNumTest": 4, "boolTest": false, "4bitSignedNumTest": -4, "message_buf": "Hello, World 4!"});

    console.log("Queue size:", queue.size());

    console.log("Encoding queue...");
    encodedQueue = queue.encode();
    console.log("Encoded queue:", encodedQueue);

    states = queue.toArray();
    
    console.log("------------------------");
    console.log("toArray():");
    states.forEach(state => {
        delete state.message_buf;
        console.log(state);
    });
    console.log("------------------------");

    while ((state = queue.pop()) != null) {
        delete state.message_buf;
        console.log("Pop:");
        console.log(state);
        console.log("------------------------");
    }

    console.log("Queue size:", queue.size());

    console.log("Decoding queue...");
    queue.decode(encodedQueue);
    console.log("Queue size:", queue.size());

    console.log("Direct decoding to array of states...");
    res = decodeStates(schema, encodedQueue);
    if (res.e != null) {
        console.log("Error decoding queue:", res.e);
        return;
    }
    states = res.d;
    console.log("------------------------");
    states.forEach(state => {
        stateCopy = JSON.parse(JSON.stringify(state));
        delete stateCopy.message_buf;
        console.log(stateCopy);
    });

    console.log("Direct encoding from array of states...");
    res = encodeStates(schema, states);
    if (res.e != null) {
        console.log("Error encoding queue:", res.e);
        return;
    }
    encodedQueue = res.d;
    console.log("Encoded queue:", encodedQueue);

    res = queue.decode(encodedQueue);
    if (res.e != null) {
        console.log("Error decoding queue:", res.e);
        return;
    }
    console.log("Queue size:", queue.size());
    console.log("toArray():");
    states = queue.toArray();
    states.forEach(state => {
        delete state.message_buf;
        console.log(state);
    });

}