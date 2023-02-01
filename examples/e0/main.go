package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/nayarsystems/bstates"
)

func main() {
	schemaRaw := `
	{
		"decoderIntMaps": 
		{
			"STATE_MAP": {
				"0" : "IDLE",
				"1" : "STOPPED",
				"2" : "RUNNING"
			}
		},
		"mappedFields":
		{
			"STATE": {			
				"from": "STATE_CODE",
				"mapId": "STATE_MAP"
			}
		},
		"decodedFields":
		{
			"MESSAGE": {
				"from": "MESSAGE_BUFFER",
				"decoder": "BufferToString"
			}
		},
		"fields": [
			{
				"name": "3BITS_INT",
				"type": "int",
				"size": 3
			},
			{
				"name": "6BITS_UINT",
				"type": "uint",
				"size": 6
			},
			{
				"name": "STATE_CODE",
				"type": "uint",
				"size": 2
			},			
			{
				"name": "BOOL1",
				"type": "bool"
			},
			{
				"name": "8BITS_INT",
				"type": "int",
				"size": 8
			},
			{
				"name": "BOOL2",
				"type": "bool"
			},
			{
				"name": "MESSAGE_BUFFER",
				"type": "buffer",
				"size": 96
			}
		]
	}`

	var schema bstates.StateSchema
	if err := json.Unmarshal([]byte(schemaRaw), &schema); err != nil {
		perrf("can't parse schema: %v\n", err)
	}

	state, err := bstates.CreateState(&schema)
	if err != nil {
		perrf("can't create state using schema: %v\n", err)
	}

	printState("state 0", state)
	if err = state.Set("3BITS_INT", -3); err != nil {
		perrf("can't update value: %v\n", err)
	}
	if err = state.Set("6BITS_UINT", 0b111111); err != nil {
		perrf("can't update value: %v\n", err)
	}
	if err = state.Set("STATE_CODE", 2); err != nil {
		perrf("can't update value: %v\n", err)
	}
	if err = state.Set("BOOL1", true); err != nil {
		perrf("can't update value: %v\n", err)
	}
	if err = state.Set("8BITS_INT", -127); err != nil {
		perrf("can't update value: %v\n", err)
	}
	if err = state.Set("BOOL2", true); err != nil {
		perrf("can't update value: %v\n", err)
	}
	if err = state.Set("MESSAGE_BUFFER", "Hello World"); err != nil {
		perrf("can't update value: %v\n", err)
	}

	printState("state 1", state)

	raw, err := state.Encode()
	if err != nil {
		perrf("can't encode state: %v\n", err)
	}
	stateFromRaw, err := bstates.CreateState(&schema)
	if err != nil {
		perrf("can't create state from schema: %v\n", err)
	}
	if err = stateFromRaw.Decode(raw); err != nil {
		perrf("can't create state from schema: %v\n", err)
	}

	printState("state 1 from raw", stateFromRaw)

}

func printState(name string, state *bstates.State) {
	msiState, _ := state.ToMsi()
	msiStateStr, _ := json.MarshalIndent(msiState, "", "  ")

	fmt.Printf("%s: %v\n", name, string(msiStateStr))

	raw, _ := state.Encode()

	fmt.Printf("%s (RAW): [ ", name)
	for _, n := range raw {
		fmt.Printf("%08b ", n)
	}
	fmt.Printf("]\n---------------\n")
}

func perrf(msg string, a ...interface{}) {
	fmt.Printf(msg, a...)
	os.Exit(1)
}
