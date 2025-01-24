import { load } from '../index.js';

test('WASM module loads and exports functions', async () => {
    const bstates = await load();
    expect(typeof bstates.createStateQueue).toBe('function');
    expect(typeof bstates.decodeStates).toBe('function');
    expect(typeof bstates.encodeStates).toBe('function');
});

test('createStateQueue initializes a queue with valid schema', async () => {
    const { createStateQueue } = await load();
    const schema = {
        version: "2.0",
        encoderPipeline: "t:z",
        decodedFields: [],
        fields: [
            { name: "testField", type: "int", size: 8 }
        ]
    };

    // Create the state queue with the schema
    const result = createStateQueue(schema);
    expect(result.e).toBeNull();
    expect(result.d).toBeDefined();
    expect(typeof result.d.push).toBe('function');
});

test('encodeStates encodes an array of states', async () => {
    const { encodeStates } = await load();
    const schema = {
        version: "2.0",
        encoderPipeline: "t:z",
        decodedFields: [],
        fields: [
            { name: "testField", type: "int", size: 8 }
        ]
    };
    let states = [{ testField: 123 }];
    let result = encodeStates(schema, states);
    expect(result.e).toBeNull();
    expect(result.d).toBeInstanceOf(Uint8Array);
});

test('decodeStates decodes an array of states', async () => {
    const { createStateQueue, encodeStates, decodeStates } = await load();
    const schema = {
        version: "2.0",
        encoderPipeline: "t:z",
        decodedFields: [],
        fields: [
            { name: "testField", type: "int", size: 8 }
        ]
    };

    // Encode the list of states following the schema
    let states = [{ testField: 123 }, { testField: 124 }];
    let result = encodeStates(schema, states);
    expect(result.e).toBeNull();
    expect(result.d).toBeInstanceOf(Uint8Array);

    // Decode the the list of states
    result = decodeStates(schema, result.d);
    expect(result.e).toBeNull();
    expect(result.d).toEqual([{ testField: 123 }, { testField: 124 }]);
});

test('queue methods', async () => {
    const { createStateQueue, encodeStates, decodeStates } = await load();
    const schema = {
        version: "2.0",
        encoderPipeline: "t:z",
        decodedFields: [],
        fields: [
            { name: "testField", type: "int", size: 8 }
        ]
    };

    // Create the state queue with the schema
    const queue = createStateQueue(schema).d;

    // Push a state to the queue (state must match schema)
    queue.push({ testField: 123 });
    queue.push({ testField: 124 });

    // Get the size of the queue. It must be 2
    let size = queue.size();
    expect(size).toBe(2);

    // Get the states in the queue. It must be an array with two states
    let decodedStates = queue.toArray();
    expect(decodedStates).toEqual([{ testField: 123 }, { testField: 124 }]);
    
    // Pop a state from the queue. It must be the first state pushed
    let state = queue.pop();
    expect(state).toEqual({ testField: 123 });

    // Encode the queue. This will generate a Uint8Array
    const encodedQueue = queue.encode();
    expect(encodedQueue).toBeInstanceOf(Uint8Array);

    // Decode the encoded queue. It must match the states pushed
    const decodedQueue = decodeStates(schema, encodedQueue);
    expect(decodedQueue.d).toEqual([{ testField: 124 }]);

    // Clear the queue
    queue.pop();

    // Get the size of the queue. It must be 0
    size = queue.size();
    expect(size).toBe(0);

    // Reload the queue from the encoded queue
    let res = queue.decode(encodedQueue);
    expect(res.e).toBeNull();

    // Get the size of the queue. It must be 1
    size = queue.size();
    expect(size).toBe(1);

    // Get the unique state in the queue. It must be the second state pushed
    state = queue.pop();
    expect(state).toEqual({ testField: 124 });
});