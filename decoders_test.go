package bstates

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func Test_IntMapDecoder(t *testing.T) {
	schema, err := CreateStateSchema(&StateSchemaParams{
		Fields: []StateField{
			{
				Name:         "F_INT32",
				DefaultValue: -1,
				Type:         T_INT,
				Size:         3,
			},
		},
		DecodedFields: []DecodedStateField{
			{
				Name: "TYPE",
				Decoder: &IntMapDecoder{
					From:  "F_INT32",
					MapId: "TYPE_MAP",
				},
			},
		},
		DecoderIntMaps: map[string]map[int64]interface{}{
			"TYPE_MAP": {
				-2: "TYPE A",
				-1: "TYPE B",
				0:  "TYPE C",
				1:  "TYPE D",
				2:  "TYPE E",
			},
		},
	})
	require.Nil(t, err)
	state, err := CreateState(schema)
	require.Nil(t, err)

	state.Set("F_INT32", -2)
	v, err := state.Get("TYPE")
	require.Nil(t, err)
	require.Equal(t, "TYPE A", v)

	err = state.Set("TYPE", "TYPE D")
	require.Error(t, err)

	v, err = state.Get("TYPE")
	require.Nil(t, err)
	require.Equal(t, "TYPE A", v) // unchanged
}

func Test_NumberToUnixTsMsDecoder(t *testing.T) {
	schema, err := CreateStateSchema(&StateSchemaParams{
		Fields: []StateField{
			{
				Name:         "DateUnixMs_Int_Raw",
				DefaultValue: 0,
				Type:         T_INT,
				Size:         32,
			},
			{
				Name:         "DateUnixMs_Float_Raw",
				DefaultValue: 0,
				Type:         T_FLOAT32,
			},
		},
		DecodedFields: []DecodedStateField{
			{
				Name: "DateUnixMs_Int",
				Decoder: &NumberToUnixTsMsDecoder{
					From:   "DateUnixMs_Int_Raw",
					Factor: 1,
					Year:   2021,
				},
			}, {
				Name: "DateUnixMs_Float",
				Decoder: &NumberToUnixTsMsDecoder{
					From:   "DateUnixMs_Float_Raw",
					Factor: 1000,
					Year:   2021,
				},
			},
		},
	})
	require.Nil(t, err)
	state, err := CreateState(schema)
	require.Nil(t, err)

	t0 := uint64(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli())
	t1 := uint64(time.Date(2021, 1, 1, 0, 0, 1, 1000000, time.UTC).UnixMilli())

	// Test from int field
	v, err := state.Get("DateUnixMs_Int")
	require.Nil(t, err)
	require.Equal(t, t0, v)

	err = state.Set("DateUnixMs_Int", t1)
	require.NoError(t, err)

	v, err = state.Get("DateUnixMs_Int_Raw")
	require.Nil(t, err)
	require.Equal(t, 1001, v)

	// Test from float field
	v, err = state.Get("DateUnixMs_Float")
	require.Nil(t, err)
	require.Equal(t, t0, v)

	err = state.Set("DateUnixMs_Float", t1)
	require.NoError(t, err)

	v, err = state.Get("DateUnixMs_Float_Raw")
	require.Nil(t, err)
	require.Equal(t, float32(1.001), v)
}

func Test_BufferToStringDecoder(t *testing.T) {
	schema, err := CreateStateSchema(&StateSchemaParams{
		Fields: []StateField{
			{
				Name:         "message_raw",
				DefaultValue: []byte(""),
				Type:         T_BUFFER,
				Size:         512,
			},
		},
		DecodedFields: []DecodedStateField{
			{
				Name:    "message",
				Decoder: &BufferToStringDecoder{From: "message_raw"},
			},
		},
	})
	require.Nil(t, err)
	state, err := CreateState(schema)
	require.Nil(t, err)

	state.Set("message_raw", []byte("hello"))
	v, err := state.Get("message")
	require.Nil(t, err)
	require.Equal(t, "hello", v)

	state.Set("message", "world")
	v, err = state.Get("message")
	require.Nil(t, err)
	require.Equal(t, "world", v)

	v, err = state.Get("message_raw")
	require.Nil(t, err)
	require.Equal(t, []byte("world"), v)
}
