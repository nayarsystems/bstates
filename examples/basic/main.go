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
		"version": "2.0",
		"encoderPipeline": "t:z",
		"decoderIntMaps": 
		{
			"STATE_MAP": {
				"0" : "IDLE",
				"1" : "STOPPED",
				"2" : "RUNNING"
			}
		},
		"decodedFields": [
			{
				"name": "MESSAGE",
				"decoder": "BufferToString",
				"params": {
					"from": "MESSAGE_BUFFER"
				}
			},
			{
				"name": "STATE",
				"decoder": "IntMap",
				"params": {
					"from": "STATE_CODE",
					"mapId": "STATE_MAP"
				}
			},
			{
				"name": "TIMESTAMP_MS",
				"decoder": "NumberToUnixTsMs",
				"params": {
					"from": "48BIT_SECS_FROM_2022",
					"year": "2022",
					"factor": 1000
				}
			}
		],
		"fields": [
			{
				"name": "STATE_CODE",
				"type": "int",
				"size": 2
			},
			{
				"name": "CHAR",
				"type": "int",
				"size": 8
			},
			{
				"name": "BOOL",
				"type": "bool"
			},
			{
				"name": "3BITS_INT",
				"type": "int",
				"size": 3
			},
			{
				"name": "48BIT_SECS_FROM_2022",
				"type": "uint",
				"size": 48
			},
			{
				"name": "MESSAGE_BUFFER",
				"type": "buffer",
				"size": 96
			},
			{		
			"name": "FLOAT32",
			"type": "float32"
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
	if err = state.Set("STATE_CODE", 2); err != nil {
		perrf("can't update value: %v\n", err)
	}
	if err = state.Set("BOOL", true); err != nil {
		perrf("can't update value: %v\n", err)
	}
	if err = state.Set("MESSAGE_BUFFER", "Hello World"); err != nil {
		perrf("can't update value: %v\n", err)
	}
	if err = state.Set("FLOAT32", 12345678.5); err != nil {
		perrf("can't update value: %v\n", err)
	}
	// update time
	//offsetDate := time.Date(2022, time.January, 1, 0, 0, 0, 0, time.UTC)
	state.Set("48BIT_SECS_FROM_2022", 0)

	if err = state.Set("FLOAT32", 12345678.5); err != nil {
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
		perrf("can't decode state: %v\n", err)
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
