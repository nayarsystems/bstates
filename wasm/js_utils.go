package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"syscall/js"

	"github.com/jaracil/ei"
	"github.com/nayarsystems/bstates"
)

var uint8Array = js.Global().Get("Uint8Array")
var array = js.Global().Get("Array")
var object = js.Global().Get("Object")
var JSON = js.Global().Get("JSON")

func stateToJsValue(state *bstates.State) (js.Value, error) {
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

func getSchemaFromJsValue(value js.Value) (*bstates.StateSchema, error) {
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

func getStateQueueFromJsValue(queueJs js.Value) (*bstates.StateQueue, error) {
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

func buildStateQueueFromJsStateArray(schema *bstates.StateSchema, stateJs js.Value) (*bstates.StateQueue, error) {
	var statesJson string
	if stateJs.Type() == js.TypeString {
		statesJson = stateJs.String()
	} else if stateJs.Type() == js.TypeObject {
		statesJson = JSON.Call("stringify", stateJs).String()
	}
	var statesArrayRaw []map[string]any
	if err := json.Unmarshal([]byte(statesJson), &statesArrayRaw); err != nil {
		return nil, fmt.Errorf("can't parse state: %v", err)
	}

	queue := bstates.CreateStateQueue(schema)
	if queue == nil {
		return nil, fmt.Errorf("can't create state queue")
	}

	state, err := bstates.CreateState(schema)
	if err != nil {
		return nil, fmt.Errorf("can't create state using schema: %v", err)
	}

	for _, stateMap := range statesArrayRaw {
		err = fillStateFromMap(state, stateMap)
		if err != nil {
			return nil, fmt.Errorf("can't fill state from map: %v", err)
		}
		queue.Push(state)
	}

	return queue, nil
}

func fillStateFromMap(state *bstates.State, stateMap map[string]any) error {
	for k, vRaw := range stateMap {
		switch v := vRaw.(type) {
		case map[string]any:
			// (?) Array of bytes can be unmarshalled as a map[string]any
			// from a stringified JSON object.
			data := make([]byte, len(v))
			for indexStr, valueRaw := range v {
				index, err := strconv.Atoi(indexStr)
				if err != nil {
					continue
				}
				data[index], err = ei.N(valueRaw).Uint8()
				if err != nil {
					continue
				}
			}
			vRaw = data
		default:
		}
		state.Set(k, vRaw)
	}
	return nil
}

func jsValueToState(schema *bstates.StateSchema, stateJs js.Value) (*bstates.State, error) {
	var stateJson string
	if stateJs.Type() == js.TypeString {
		stateJson = stateJs.String()
	} else if stateJs.Type() == js.TypeObject {
		stateJson = JSON.Call("stringify", stateJs).String()
	}

	var stateMap map[string]any
	err := json.Unmarshal([]byte(stateJson), &stateMap)
	if err != nil {
		return nil, fmt.Errorf("can't parse state: %v", err)
	}

	state, err := bstates.CreateState(schema)
	if err != nil {
		return nil, fmt.Errorf("can't create state using schema: %v", err)
	}

	err = fillStateFromMap(state, stateMap)
	if err != nil {
		return nil, fmt.Errorf("can't fill state from map: %v", err)
	}

	return state, nil
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
