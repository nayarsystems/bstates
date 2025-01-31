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
        fields: [
            { name: "testField", type: "int", size: 8 }
        ]
    };

    // Create the state queue with the schema
    let queue = createStateQueue(schema);
    expect(queue).toBeDefined();
    expect(typeof queue.push).toBe('function');
});

test('createStateQueue initializes a queue with invalid schema', async () => {
    const { createStateQueue } = await load();
    const schema = {
    };

    // Create the state queue with the schema
    expect(() => createStateQueue(schema)).toThrow();
});

test('encodeStates encodes an array of states', async () => {
    const { encodeStates } = await load();
    const schema = {
        version: "2.0",
        encoderPipeline: "t:z",
        fields: [
            { name: "testField", type: "int", size: 8 }
        ]
    };
    let states = [{ testField: 123 }];
    let encodedStates = encodeStates(schema, states);
    expect(encodedStates).toBeInstanceOf(Uint8Array);
});

test('encodeStates error from invalid schema', async () => {
    const { encodeStates } = await load();
    const schema = {
    };
    let states = [{ testField: 123 }];
    expect(() => encodeStates(schema, states)).toThrow();
});

test('decodeStates decodes an array of states', async () => {
    const { createStateQueue, encodeStates, decodeStates } = await load();
    const schema = {
        version: "2.0",
        encoderPipeline: "t:z",
        fields: [
            { name: "testField", type: "int", size: 8 }
        ]
    };

    // Encode the list of states following the schema
    let states = [{ testField: 123 }, { testField: 124 }];
    let encodedStates = encodeStates(schema, states);
    expect(encodedStates).toBeInstanceOf(Uint8Array);

    // Decode the the list of states
    let decodedStates = decodeStates(schema, encodedStates);
    expect(decodedStates).toEqual([{ testField: 123 }, { testField: 124 }]);
});

test('decodeStates error from wrong schema', async () => {
    const { createStateQueue, encodeStates, decodeStates } = await load();
    let schema = {
        version: "2.0",
        fields: [
            { name: "testField", type: "int", size: 8 }
        ]
    };

    // Encode the list of states following the schema
    let states = [{ testField: 123 }, { testField: 124 }];
    let encodedStates = encodeStates(schema, states);
    expect(encodedStates).toBeInstanceOf(Uint8Array);

    schema = {
        version: "2.0",
        fields: [
            { name: "testField", type: "int", size: 32 },
        ]
    };
    expect(() => decodeStates(schema, encodedStates)).toThrow();
});

test('queue methods', async () => {
    const { createStateQueue, encodeStates, decodeStates } = await load();
    const schema = {
        version: "2.0",
        encoderPipeline: "t:z",
        fields: [
            { name: "testField", type: "int", size: 8 }
        ]
    };

    // Create the state queue with the schema
    const queue = createStateQueue(schema);

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
    expect(decodedQueue).toEqual([{ testField: 124 }]);

    // Clear the queue
    queue.pop();

    // Get the size of the queue. It must be 0
    size = queue.size();
    expect(size).toBe(0);

    // Reload the queue from the encoded queue
    queue.decode(encodedQueue);

    // Get the size of the queue. It must be 1
    size = queue.size();
    expect(size).toBe(1);

    // Get the unique state in the queue. It must be the second state pushed
    state = queue.pop();
    expect(state).toEqual({ testField: 124 });
});