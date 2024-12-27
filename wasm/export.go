package main

import (
	"github.com/nayarsystems/bstates"
	"syscall/js"
)

func createStateQueue(this js.Value, args []js.Value) any {
	if len(args) != 1 {
		return retE("invalid number of arguments")
	}

	schema, err := getSchemaFromJsValue(args[0])
	if err != nil {
		return retE("can't get schema from arguments: %v", err)
	}

	queue := bstates.CreateStateQueue(schema)
	if queue == nil {
		return retE("can't create state queue")
	}

	jsQueue := object.New()

	jsQueue.Set("push", js.FuncOf(stateQueuePush))
	jsQueue.Set("pop", js.FuncOf(stateQueuePop))
	jsQueue.Set("size", js.FuncOf(stateQueueSize))
	jsQueue.Set("toArray", js.FuncOf(stateQueueToArray))
	jsQueue.Set("decode", js.FuncOf(stateQueueDecode))
	jsQueue.Set("encode", js.FuncOf(stateQueueEncode))
	jsQueue.Set("data", uint8Array.New(0))

	schemaRaw, err := schema.MarshalJSON()
	if err != nil {
		return retE("can't marshal schema: %v", err)
	}

	jsQueue.Set("schema", js.ValueOf(string(schemaRaw)))

	return retD(jsQueue)
}

func stateQueuePush(this js.Value, args []js.Value) any {
	if len(args) != 1 {
		return retE("invalid number of arguments")
	}

	queue, err := getStateQueueFromJsValue(this)
	if err != nil {
		return retE("can't get state queue from this: %v", err)
	}

	state, err := jsValueToState(queue.StateSchema, args[0])
	if err != nil {
		return retE("can't convert js value to state: %v", err)
	}

	err = queue.Push(state)
	if err != nil {
		return retE("can't push state to queue: %v", err)
	}

	data, err := queue.Encode()
	if err != nil {
		return retE("can't encode state queue: %v", err)
	}

	jsQueueData := uint8Array.New(len(data))
	js.CopyBytesToJS(jsQueueData, data)
	this.Set("data", jsQueueData)
	return nil
}

func stateQueuePop(this js.Value, args []js.Value) any {
	queue, err := getStateQueueFromJsValue(this)
	if err != nil {
		return nil
	}

	state, err := queue.Pop()
	if err != nil {
		return nil
	}

	stateJs, err := stateToJsValue(state)
	if err != nil {
		return nil
	}

	data, err := queue.Encode()
	if err != nil {
		return nil
	}

	jsQueueData := uint8Array.New(len(data))
	js.CopyBytesToJS(jsQueueData, data)
	this.Set("data", jsQueueData)

	return stateJs
}

func stateQueueSize(this js.Value, args []js.Value) any {
	queue, err := getStateQueueFromJsValue(this)
	if err != nil {
		return 0
	}

	return js.ValueOf(queue.GetNumStates())
}

func stateQueueToArray(this js.Value, args []js.Value) any {
	queue, err := getStateQueueFromJsValue(this)
	if err != nil {
		return array.New(0)
	}

	states, err := queue.GetStates()
	if err != nil {
		return array.New(0)
	}

	statesJs := array.New(0)
	for _, state := range states {
		stateJs, err := stateToJsValue(state)
		if err != nil {
			return retE("can't convert state to js: %v", err)
		}
		statesJs.Call("push", stateJs)
	}

	return statesJs
}

func stateQueueDecode(this js.Value, args []js.Value) any {
	if len(args) != 1 {
		return retE("invalid number of arguments: <encoded data in a Uint8Array>")
	}

	jsData := args[0]

	queue, err := getStateQueueFromJsValue(this)
	if err != nil {
		return retE("can't get state queue from this: %v", err)
	}

	var data []byte
	if jsData.Type() == js.TypeObject {
		data = make([]byte, jsData.Length())
		js.CopyBytesToGo(data, jsData)
	} else {
		return retE("data must be a Uint8Array")
	}

	err = queue.Decode(data)
	if err != nil {
		return retE("can't decode state queue: %v", err)
	}

	jsQueueData := uint8Array.New(len(data))
	js.CopyBytesToJS(jsQueueData, data)
	this.Set("data", jsQueueData)

	return retD(nil)
}

func stateQueueEncode(this js.Value, args []js.Value) any {
	return this.Get("data")
}

func decodeStates(this js.Value, args []js.Value) any {
	if len(args) != 2 {
		return retE("invalid number of arguments: <schema>, <encoded data in a Uint8Array>")
	}
	schema, err := getSchemaFromJsValue(args[0])
	if err != nil {
		return retE("can't get schema from arguments: %v", err)
	}

	var data []byte
	if args[1].Type() == js.TypeObject {
		data = make([]byte, args[1].Length())
		js.CopyBytesToGo(data, args[1])
	} else {
		return retE("data must be a Uint8Array")
	}

	queue := bstates.CreateStateQueue(schema)
	if queue == nil {
		return retE("can't create state queue")
	}

	err = queue.Decode(data)
	if err != nil {
		return retE("can't decode state queue: %v", err)
	}

	states, err := queue.GetStates()
	if err != nil {
		return retE("can't get states from queue: %v", err)
	}

	statesJs := array.New(0)
	for _, state := range states {
		stateJs, err := stateToJsValue(state)
		if err != nil {
			return retE("can't convert state to js: %v", err)
		}
		statesJs.Call("push", stateJs)
	}

	return retD(statesJs)
}

func encodeStates(this js.Value, args []js.Value) any {
	if len(args) != 2 {
		return retE("invalid number of arguments: <schema>, <array of states>")
	}

	schema, err := getSchemaFromJsValue(args[0])
	if err != nil {
		return retE("can't get schema from arguments: %v", err)
	}

	queue, err := buildStateQueueFromJsStateArray(schema, args[1])
	if err != nil {
		return retE("can't create state queue from js array: %v", err)
	}

	data, err := queue.Encode()
	if err != nil {
		return retE("can't encode state queue: %v", err)
	}

	jsQueueData := uint8Array.New(len(data))
	js.CopyBytesToJS(jsQueueData, data)

	return retD(jsQueueData)
}
