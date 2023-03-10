package bstates

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Unmarshall_V1Schema(t *testing.T) {
	schemaRaw :=
		`
	{
		"encoderPipeline": "t:z",
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
				"name": "3BITS INT",
				"type": "int",
				"size": 3
			},
			{
				"name": "6BIT_UINT",
				"type": "uint",
				"size": 6
			},
			{
				"name": "323BIT_BUFFER",
				"type": "buffer",
				"size": 323
			},
			{
				"name": "MESSAGE_BUFFER",
				"type": "buffer",
				"size": 96
			}
		]
	}
	`
	var schema StateSchema
	err := json.Unmarshal([]byte(schemaRaw), &schema)
	require.NoError(t, err)
	eSchema := createV1SchemaForJSONTests(t)
	require.Equal(t, eSchema, &schema)

}

func Test_Marshall_V1Schema(t *testing.T) {
	schema := createV1SchemaForJSONTests(t)
	raw, err := json.Marshal(schema)
	require.NoError(t, err)
	var fromRaw StateSchema
	err = json.Unmarshal(raw, &fromRaw)
	require.NoError(t, err)
	require.Equal(t, schema, &fromRaw)
	require.Equal(t, schema.GetSHA256(), fromRaw.GetSHA256())
}

func createV1SchemaForJSONTests(t *testing.T) *StateSchema {
	schema, err := CreateStateSchema(
		&StateSchemaParams{
			EncoderPipeline: "t:z",
			DecoderIntMaps: map[string]map[int64]interface{}{
				"STATE_MAP": {
					0: "IDLE",
					1: "STOPPED",
					2: "RUNNING",
				},
			},
			DecodedFields: []DecodedStateField{
				{
					Name: "MESSAGE",
					Decoder: &BufferToStringDecoder{
						From: "MESSAGE_BUFFER",
					},
				},
				{
					Name: "STATE",
					Decoder: &IntMapDecoder{
						From:  "STATE_CODE",
						MapId: "STATE_MAP",
					},
				},
			},
			Fields: []StateField{
				{
					Name: "STATE_CODE",
					Type: T_INT,
					Size: 2,
				},
				{
					Name: "CHAR",
					Type: T_INT,
					Size: 8,
				},
				{
					Name: "BOOL",
					Type: T_BOOL,
					Size: 1,
				},
				{
					Name: "3BITS INT",
					Type: T_INT,
					Size: 3,
				},
				{
					Name: "6BIT_UINT",
					Type: T_UINT,
					Size: 6,
				},
				{
					Name: "323BIT_BUFFER",
					Type: T_BUFFER,
					Size: 323,
				},
				{
					Name: "MESSAGE_BUFFER",
					Type: T_BUFFER,
					Size: 96,
				},
			},
		},
	)
	require.NoError(t, err)
	return schema
}

func Test_V1Schema_CodeToStringMap(t *testing.T) {
	schemaJson := `
		{
			"encoderPipeline": "t:z",
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
			"fields": [
				{
					"name": "STATE_CODE",
					"type": "int",
					"size": 2
				}
			]
		}
	`
	schema := &StateSchema{}
	err := schema.UnmarshalJSON([]byte(schemaJson))
	require.NoError(t, err)
	state, err := schema.CreateState()
	require.NoError(t, err)

	err = state.Set("STATE_CODE", 2)
	require.NoError(t, err)

	v, err := state.Get("STATE")
	require.NoError(t, err)
	require.Equal(t, "RUNNING", v)
}
