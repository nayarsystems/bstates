package bstates

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_GetSHA256(t *testing.T) {
	params1 := createSchemaParams(t)
	schema1, err := CreateStateSchema(params1)
	require.NoError(t, err)

	// This assert repetition is ok. Here we check that the
	// obtained hash is always the same for the same input.
	require.Equal(t, schema1.GetSHA256(), schema1.GetSHA256())
	require.Equal(t, schema1.GetSHA256(), schema1.GetSHA256())

	// New schema created with the same parameters must has the same hash
	params2 := createSchemaParams(t)
	schema2, err := CreateStateSchema(params2)
	require.NoError(t, err)
	require.Equal(t, schema1.GetSHA256(), schema2.GetSHA256())

	// New schema created with different parameters must has different hash
	params3 := createSchemaParams(t)
	params3.DecodedFields[0].Name = params3.DecodedFields[0].Name + "_2"
	schema3, err := CreateStateSchema(params3)
	require.NoError(t, err)
	require.NotEqual(t, schema1.GetSHA256(), schema3.GetSHA256())
}

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

func Test_Unmarshall(t *testing.T) {
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
	eSchema := createSchema(t)
	require.Equal(t, eSchema, &schema)

}

func Test_Marshall(t *testing.T) {
	schema := createSchema(t)
	raw, err := json.Marshal(schema)
	require.NoError(t, err)
	var fromRaw StateSchema
	err = json.Unmarshal(raw, &fromRaw)
	require.NoError(t, err)
	require.Equal(t, schema, &fromRaw)
	require.Equal(t, schema.GetSHA256(), fromRaw.GetSHA256())
}

func Test_Meta(t *testing.T) {
	schemaEmptyMeta := createSchemaWithMeta(t, map[string]any{})
	schemaEmptyMetaHash := schemaEmptyMeta.GetHashString()

	schemaNilMeta := createSchemaWithMeta(t, nil)
	schemaNilMetaHash := schemaNilMeta.GetHashString()

	require.Equal(t, schemaEmptyMetaHash, schemaNilMetaHash)

	schemaWithMeta := createSchemaWithMeta(t, map[string]any{"class": "A"})
	schemaWithMetaHash := schemaWithMeta.GetHashString()

	require.NotEqual(t, schemaEmptyMetaHash, schemaWithMetaHash)

	require.Equal(t, schemaWithMeta.GetMeta(), map[string]any{"class": "A"})

	raw, err := json.Marshal(schemaWithMeta)
	require.NoError(t, err)

	var fromRaw StateSchema
	err = json.Unmarshal(raw, &fromRaw)
	require.NoError(t, err)
	require.Equal(t, schemaWithMeta, &fromRaw)
	require.Equal(t, fromRaw.GetMeta(), map[string]any{"class": "A"})
}

func createSchema(t *testing.T) *StateSchema {
	schema, err := CreateStateSchema(createSchemaParams(t))
	require.NoError(t, err)
	return schema
}

func createSchemaWithMeta(t *testing.T, meta map[string]any) *StateSchema {
	params := createSchemaParams(t)
	params.Meta = meta
	schema, err := CreateStateSchema(params)
	require.NoError(t, err)
	return schema
}

func createSchemaParams(t *testing.T) *StateSchemaParams {
	return &StateSchemaParams{
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
	}
}
