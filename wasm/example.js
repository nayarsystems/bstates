export async function example(bs) {
    console.log("Creating queue...");
    const schema = {
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
        "fields": [
            { "name": "3bitUnsignedNumTest", "type": "uint", "size": 2 },
            { "name": "boolTest", "type": "bool" },
            { "name": "4bitSignedNumTest", "type": "int", "size": 4 },
            { "name": "message_buf", "type": "buffer", "size": 1024 }
        ]
    };

    // Create the state queue
    let res = bs.createStateQueue(schema);
    if (res.e != null) {
        console.error("Error creating queue:", res.e);
        exit(1);
    }
    let queue = res.d;

    console.log("Queue created:", queue);

    console.log("Pushing states to queue...");
    queue.push({ "3bitUnsignedNumTest": 1, "boolTest": true, "4bitSignedNumTest": -7, "message": "Hello, World 1!" });
    queue.push({ "3bitUnsignedNumTest": 2, "boolTest": false, "4bitSignedNumTest": -6, "message": "Hello, World 2!" });
    queue.push({ "3bitUnsignedNumTest": 3, "boolTest": true, "4bitSignedNumTest": -5, "message": "Hello, World 3!" });
    queue.push({ "3bitUnsignedNumTest": 4, "boolTest": false, "4bitSignedNumTest": -5, "message_buf": new TextEncoder().encode("Hello, World 4!") });
    // Alternative to the previous line only in Node.js:
    // queue.push({ "3bitUnsignedNumTest": 4, "boolTest": false, "4bitSignedNumTest": -5, "message_buf": new Uint8Array(Buffer.from("Hello, World 4!", 'utf-8')) });
    queue.push({ "3bitUnsignedNumTest": 5, "boolTest": false, "4bitSignedNumTest": -4, "message_buf": "Hello, World 5!" });

    console.log("Queue size:", queue.size());

    console.log("Encoding queue...");
    let encodedQueue = queue.encode();
    console.log("Encoded queue:", encodedQueue);

    let states = queue.toArray();

    console.log("------------------------");
    console.log("toArray():");
    states.forEach(state => {
        delete state.message_buf;
        console.log(state);
    });
    console.log("------------------------");

    let state;
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
    res = bs.decodeStates(schema, encodedQueue);
    if (res.e != null) {
        console.error("Error decoding queue:", res.e);
        exit(1);
    }
    states = res.d;
    console.log("------------------------");
    states.forEach(state => {
        let stateCopy = JSON.parse(JSON.stringify(state));
        delete stateCopy.message_buf;
        console.log(stateCopy);
    });

    console.log("Direct encoding from array of states...");
    res = bs.encodeStates(schema, states);
    if (res.e != null) {
        console.error("Error encoding queue:", res.e);
        exit(1);
    }
    encodedQueue = res.d;
    console.log("Encoded queue:", encodedQueue);

    res = queue.decode(encodedQueue);
    if (res.e != null) {
        console.error("Error decoding queue:", res.e);
        exit(1);
    }
    console.log("Queue size:", queue.size());
    console.log("toArray():");
    states = queue.toArray();
    states.forEach(state => {
        delete state.message_buf;
        console.log(state);
    });
}
