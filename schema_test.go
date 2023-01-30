package bstates

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Unmarshall_InvalidType(t *testing.T) {
	schemaRaw :=
		`
	{
		"fields": [
			{
				"name": "A",
				"type": "i23nt",
				"size": 8
			}
		]
	}
	`
	var schema StateSchema
	err := json.Unmarshal([]byte(schemaRaw), &schema)
	require.NotNil(t, err)
}

func Test_Unmarshall(t *testing.T) {
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
	require.Nil(t, err)
	eSchema := createSchemaForJSONTests(t)
	require.Equal(t, eSchema, &schema)

}

func Test_Marshall(t *testing.T) {
	schema := createSchemaForJSONTests(t)
	raw, err := json.Marshal(schema)
	require.Nil(t, err)
	var fromRaw StateSchema
	err = json.Unmarshal(raw, &fromRaw)
	require.Nil(t, err)
	require.Equal(t, schema, &fromRaw)
	require.Equal(t, schema.GetSHA256(), fromRaw.GetSHA256())
}

func createSchemaForJSONTests(t *testing.T) *StateSchema {
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
			MappedFields: map[string]MappedStateField{
				"STATE": {
					From:  "STATE_CODE",
					MapId: "STATE_MAP",
				},
			},
			DecodedFields: map[string]DecodedStateFields{
				"MESSAGE": {
					From:         "MESSAGE_BUFFER",
					FieldDecoder: BufferToString,
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
	require.Nil(t, err)
	return schema
}

func Test_Unmarshall_InvalidBufferSize(t *testing.T) {
	schemaRaw :=
		`
	{
		"fields": [
			{
				"name": "A",
				"type": "buffer",
				"size": 0
			}
		]
	}
	`
	var schema StateSchema
	err := json.Unmarshal([]byte(schemaRaw), &schema)
	require.NotNil(t, err)
}

func Test_Unmarshall_InvalidIntSize(t *testing.T) {
	schemaRaw :=
		`
	{
		"fields": [
			{
				"name": "A",
				"type": "int",
				"size": 0
			}
		]
	}
	`
	var schema StateSchema
	err := json.Unmarshal([]byte(schemaRaw), &schema)
	require.NotNil(t, err)
}

func Test_CodeToStringMap(t *testing.T) {
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
	require.Nil(t, err)
	state, err := schema.CreateState()
	require.Nil(t, err)

	err = state.Set("STATE_CODE", 2)
	require.Nil(t, err)

	v, err := state.Get("STATE")
	require.Nil(t, err)
	require.Equal(t, "RUNNING", v)
}
