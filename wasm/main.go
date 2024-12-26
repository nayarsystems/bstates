package main

import (
	"encoding/json"
	"fmt"
	"syscall/js"

	"github.com/nayarsystems/bstates"
)

var uint8Array = js.Global().Get("Uint8Array")
var array = js.Global().Get("Array")
var object = js.Global().Get("Object")
var JSON = js.Global().Get("JSON")

func stateToJs(state *bstates.State) (js.Value, error) {
	stateMsi, err := state.ToMsi()
	if err != nil {
		return js.Value{}, err
	}

	stateJs := object.New()

	for k, v := range stateMsi {
		if data, ok := v.([]byte); ok {
			dataJs := uint8Array.New(len(data))
			js.CopyBytesToJS(dataJs, data)
			stateJs.Set(k, dataJs)
		} else {
			stateJs.Set(k, js.ValueOf(v))
		}
	}
	return stateJs, nil
}

func getSchemaFromArgs(args []js.Value) (*bstates.StateSchema, error) {
	var schemaRaw string
	if args[0].Type() == js.TypeString {
		schemaRaw = args[0].String()
	} else {
		schemaRaw = JSON.Call("stringify", args[0]).String()
	}

	var schema bstates.StateSchema
	if err := json.Unmarshal([]byte(schemaRaw), &schema); err != nil {
		return nil, fmt.Errorf("can't parse schema: %v", err)
	}
	return &schema, nil
}

func getStateQueueFromThis(this js.Value) (*bstates.StateQueue, error) {
	queueJs := this
	if queueJs.Type() != js.TypeObject {
		return nil, fmt.Errorf("queue must be an object")
	}

	dataJs := queueJs.Get("data")
	if dataJs.Type() != js.TypeObject {
		return nil, fmt.Errorf("data must be an object")
	}

	var data []byte
	if dataJs.Type() == js.TypeObject {
		data = make([]byte, dataJs.Length())
		js.CopyBytesToGo(data, dataJs)
	} else {
		return nil, fmt.Errorf("data must be a Uint8Array")
	}

	schema, err := getSchemaFromValue(queueJs.Get("schema"))
	if err != nil {
		return nil, fmt.Errorf("can't get schema from queue: %v", err)
	}

	queue := bstates.CreateStateQueue(schema)
	if queue == nil {
		return nil, fmt.Errorf("can't create state queue")
	}

	err = queue.Decode(data)
	if err != nil {
		return nil, fmt.Errorf("can't decode state queue: %v", err)
	}
	return queue, nil
}

func createStateQueue(this js.Value, args []js.Value) any {
	if len(args) != 1 {
		return retE("invalid number of arguments")
	}

	schema, err := getSchemaFromArgs(args)
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
	jsQueue.Set("data", uint8Array.New(0))

	schemaRaw, err := schema.MarshalJSON()
	if err != nil {
		return retE("can't marshal schema: %v", err)
	}

	jsQueue.Set("schema", js.ValueOf(string(schemaRaw)))

	return retD(jsQueue)
}

func getSchemaFromValue(value js.Value) (*bstates.StateSchema, error) {
	var schemaRaw string
	if value.Type() == js.TypeString {
		schemaRaw = value.String()
	} else {
		schemaRaw = JSON.Call("stringify", value).String()
	}

	var schema bstates.StateSchema
	if err := json.Unmarshal([]byte(schemaRaw), &schema); err != nil {
		return nil, fmt.Errorf("can't parse schema: %v", err)
	}
	return &schema, nil
}

func stateQueuePush(this js.Value, args []js.Value) any {
	if len(args) != 1 {
		return retE("invalid number of arguments")
	}

	queue, err := getStateQueueFromThis(this)
	if err != nil {
		return retE("can't get state queue from this: %v", err)
	}

	var stateJson string
	if args[0].Type() == js.TypeString {
		stateJson = args[0].String()
	} else if args[0].Type() == js.TypeObject {
		stateJson = JSON.Call("stringify", args[0]).String()
	}

	var stateMap map[string]any
	err = json.Unmarshal([]byte(stateJson), &stateMap)
	if err != nil {
		return retE("can't parse state: %v", err)
	}

	schema := queue.StateSchema

	state, err := bstates.CreateState(schema)
	if err != nil {
		return retE("can't create state using schema: %v", err)
	}

	for k, v := range stateMap {
		err = state.Set(k, v)
		if err != nil {
			return retE("can't set state value: %v", err)
		}
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
	queue, err := getStateQueueFromThis(this)
	if err != nil {
		return nil
	}

	state, err := queue.Pop()
	if err != nil {
		return nil
	}

	stateJs, err := stateToJs(state)
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
	queue, err := getStateQueueFromThis(this)
	if err != nil {
		return 0
	}

	return js.ValueOf(queue.GetNumStates())
}

func stateQueueToArray(this js.Value, args []js.Value) any {
	queue, err := getStateQueueFromThis(this)
	if err != nil {
		return array.New(0)
	}

	states, err := queue.GetStates()
	if err != nil {
		return array.New(0)
	}

	statesJs := array.New(0)
	for _, state := range states {
		stateJs, err := stateToJs(state)
		if err != nil {
			return retE("can't convert state to js: %v", err)
		}
		statesJs.Call("push", stateJs)
	}

	return statesJs
}

func retD(data any) map[string]any {
	return map[string]any{
		"d": data,
		"e": nil,
	}
}

func retE(format string, args ...interface{}) map[string]any {
	return map[string]any{
		"d": nil,
		"e": fmt.Sprintf(format, args...),
	}
}
func main() {
	wait := make(chan struct{}, 0)
	createStateQueueFunc := js.FuncOf(createStateQueue)
	defer createStateQueueFunc.Release()
	js.Global().Set("createStateQueue", createStateQueueFunc)
	<-wait
}
