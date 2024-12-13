package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"syscall/js"

	"github.com/nayarsystems/bstates"
)

var schemas = make(map[string]*bstates.StateSchema)

func logI(format string, args ...interface{}) {
	fmt.Fprintf(os.Stdout, format+"\n", args...)
}

func logE(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
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

func createSchema(this js.Value, args []js.Value) any {
	if len(args) != 1 {
		return retE("invalid number of arguments")
	}

	var schemaRaw string
	if args[0].Type() == js.TypeString {
		schemaRaw = args[0].String()
	} else {
		schemaRaw = js.Global().Get("JSON").Call("stringify", args[0]).String()
	}

	var schema bstates.StateSchema
	if err := json.Unmarshal([]byte(schemaRaw), &schema); err != nil {
		return retE("can't parse schema: %v", err)
	}

	hash := schema.GetSHA256()
	hashStr := hex.EncodeToString(hash[:])
	schemas[hashStr] = &schema

	return retD(hashStr)
}

func decodeState(this js.Value, args []js.Value) any {
	if len(args) != 2 {
		return retE("invalid number of arguments")
	}

	hash := args[0].String()
	var raw []byte
	// raw cannot be a string, it must be a Uint8Array
	if args[1].Type() == js.TypeObject {
		raw = make([]byte, args[1].Length())
		js.CopyBytesToGo(raw, args[1])
	} else {
		return retE("raw must be a Uint8Array")
	}

	schema, ok := schemas[hash]
	if !ok {
		return retE("schema not found")
	}

	state, err := bstates.CreateState(schema)
	if err != nil {
		return retE("can't create state using schema: %v", err)
	}

	if err = state.Decode([]byte(raw)); err != nil {
		return retE("can't decode state: %v", err)
	}

	obj, err := state.ToMsi()
	if err != nil {
		return retE("can't convert state to msi: %v", err)
	}

	return retD(obj)
}

func main() {
	wait := make(chan struct{}, 0)
	addSchemaFunc := js.FuncOf(createSchema)
	defer addSchemaFunc.Release()
	decodeStateFunc := js.FuncOf(decodeState)
	defer decodeStateFunc.Release()
	js.Global().Set("createSchema", addSchemaFunc)
	js.Global().Set("decodeState", decodeStateFunc)
	<-wait
}
