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
	require.Error(t, err)
}

func Test_Unmarshall_Schema(t *testing.T) {
	schemaRaw :=
		`
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
				"name": "3BITS INT",
				"type": "int",
				"size": 3
			},
			{
				"name": "48BIT_SECS_FROM_2022",
				"type": "uint",
				"size": 48
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
	eSchema := createSchemaForJSONTests(t)
	require.Equal(t, eSchema, &schema)

}

func Test_GetSHA256(t *testing.T) {
	schema := createSchemaForJSONTests(t)
	require.Equal(t, schema.GetSHA256(), schema.GetSHA256())
	require.Equal(t, schema.GetSHA256(), schema.GetSHA256())
}

func Test_Marshall(t *testing.T) {
	schema := createSchemaForJSONTests(t)
	raw, err := json.Marshal(schema)
	require.NoError(t, err)
	var fromRaw StateSchema
	err = json.Unmarshal(raw, &fromRaw)
	require.NoError(t, err)
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
				{
					Name: "TIMESTAMP_MS",
					Decoder: &NumberToUnixTsMsDecoder{
						From:   "48BIT_SECS_FROM_2022",
						Year:   2022,
						Factor: 1000,
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
					Name: "48BIT_SECS_FROM_2022",
					Type: T_UINT,
					Size: 48,
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
	require.Error(t, err)
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
	require.Error(t, err)
}

func Test_CodeToStringMap(t *testing.T) {
	schemaJson := `
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
					"name": "STATE",
					"decoder": "IntMap",
					"params": {
						"from": "STATE_CODE",
						"mapId": "STATE_MAP"
					}
				}
			],
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
