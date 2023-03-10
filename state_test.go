package bstates

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_StatesToMsiStates(t *testing.T) {
	schema, err := CreateStateSchema(&StateSchemaParams{
		Fields: []StateField{
			{
				Name:         "F_FLOAT32",
				DefaultValue: 1.5,
				Type:         T_FLOAT32,
			},
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
	state0, err := CreateState(schema)
	require.Nil(t, err)
	state1, err := CreateState(schema)
	require.Nil(t, err)

	state1.Set("F_FLOAT32", 2.7)
	state1.Set("F_INT32", -2)

	data, err := StatesToMsiStates([]*State{state0, state1})
	require.Nil(t, err)

	edata := []map[string]interface{}{
		{
			"F_FLOAT32": float32(1.5),
			"F_INT32":   -1,
			"TYPE":      "TYPE B",
		},
		{
			"F_FLOAT32": float32(2.7),
			"F_INT32":   -2,
			"TYPE":      "TYPE A",
		},
	}
	require.Equal(t, edata, data)
}

func Test_GetDeltaMsiState(t *testing.T) {
	schema, err := CreateStateSchema(&StateSchemaParams{
		Fields: []StateField{
			{
				Name:         "F_FLOAT32",
				DefaultValue: 1.5,
				Type:         T_FLOAT32,
			},
			{
				Name:         "F_INT32",
				DefaultValue: -1,
				Type:         T_INT,
				Size:         3,
			},
			{
				Name:         "F_STRING_BUFFER",
				DefaultValue: []byte{'h', 'e', 'l', 'l', 'o'},
				Type:         T_BUFFER,
				Size:         5,
			},
			{
				Name:         "F_FLOAT64_SECS_FROM_2022",
				DefaultValue: 1.234,
				Type:         T_FLOAT64,
			},
			{
				Name:         "F_UINT64_MICROS_FROM_2022",
				DefaultValue: 1234000,
				Type:         T_UINT,
				Size:         64,
			},
		},
		DecodedFields: []DecodedStateField{
			{
				Name:    "STRING",
				Decoder: &BufferToStringDecoder{From: "F_STRING_BUFFER"},
			},
			{
				Name:    "TYPE",
				Decoder: &IntMapDecoder{From: "F_INT32", MapId: "TYPE_MAP"},
			},
			{
				Name: "TIMESTAMP_MS",
				Decoder: &NumberToUnixTsMsDecoder{
					From:   "F_FLOAT64_SECS_FROM_2022",
					Year:   2022,
					Factor: 1000,
				},
			},
			{
				Name: "TIMESTAMP_MS_2",
				Decoder: &NumberToUnixTsMsDecoder{
					From:   "F_UINT64_MICROS_FROM_2022",
					Year:   2022,
					Factor: 0.001,
				},
			},
		},
		DecoderIntMaps: map[string]map[int64]interface{}{
			"TYPE_MAP": {
				-1: "TYPE A",
				-2: "TYPE B",
			},
		},
	})
	require.Nil(t, err)
	state0, err := CreateState(schema)
	require.Nil(t, err)
	state1 := state0.GetCopy()

	state1.Set("F_INT32", -2)
	data, err := GetDeltaMsiState(state0, state1)
	require.Nil(t, err)
	edata := map[string]interface{}{
		"F_INT32": -2,
		"TYPE":    "TYPE B",
	}
	require.Equal(t, edata, data)

	state0 = state1.GetCopy()
	state1.Set("F_FLOAT32", 2.7)
	edata = map[string]interface{}{
		"F_FLOAT32": float32(2.7),
	}
	data, err = GetDeltaMsiState(state0, state1)
	require.Nil(t, err)
	require.Equal(t, edata, data)

	state0 = state1.GetCopy()
	state1.Set("F_INT32", -1)
	data, err = GetDeltaMsiState(state0, state1)
	require.Nil(t, err)
	edata = map[string]interface{}{
		"F_INT32": -1,
		"TYPE":    "TYPE A",
	}
	require.Equal(t, edata, data)

	state0 = state1.GetCopy()
	state1.Set("F_STRING_BUFFER", []byte{'h', 'i'})
	data, err = GetDeltaMsiState(state0, state1)
	require.Nil(t, err)
	edata = map[string]interface{}{
		"F_STRING_BUFFER": []byte{'h', 'i'},
		"STRING":          "hi",
	}
	require.Equal(t, edata, data)

	offsetDate := time.Date(2022, time.January, 1, 0, 0, 0, 0, time.UTC)
	oldTimestampMs := offsetDate.UnixMilli() + 1234
	newTimestampMs := offsetDate.UnixMilli() + 10987
	require.Equal(t, "2022-01-01 00:00:01.234 +0000 UTC", time.UnixMilli(oldTimestampMs).UTC().String())
	require.Equal(t, "2022-01-01 00:00:10.987 +0000 UTC", time.UnixMilli(newTimestampMs).UTC().String())

	// Let's check a change in F_FLOAT64_SECS_FROM_2022
	state0 = state1.GetCopy()
	state1.Set("F_FLOAT64_SECS_FROM_2022", 10.987)
	data, err = GetDeltaMsiState(state0, state1)
	require.Nil(t, err)

	v, err := state0.Get("TIMESTAMP_MS")
	require.NoError(t, err)
	require.Equal(t, uint64(oldTimestampMs), v)

	edata = map[string]interface{}{
		"F_FLOAT64_SECS_FROM_2022": 10.987,
		"TIMESTAMP_MS":             uint64(newTimestampMs),
	}
	require.Equal(t, edata, data)

	// Let's check a change in F_UINT64_MICROS_FROM_2022
	state0 = state1.GetCopy()
	state1.Set("F_UINT64_MICROS_FROM_2022", 10987000)
	data, err = GetDeltaMsiState(state0, state1)
	require.Nil(t, err)

	v, err = state0.Get("TIMESTAMP_MS_2")
	require.NoError(t, err)
	require.Equal(t, uint64(oldTimestampMs), v)

	edata = map[string]interface{}{
		"F_UINT64_MICROS_FROM_2022": uint64(10987000),
		"TIMESTAMP_MS_2":            uint64(newTimestampMs),
	}
	require.Equal(t, edata, data)

	state0 = state1.GetCopy()
	edata = map[string]interface{}{}
	data, err = GetDeltaMsiState(state0, state1)
	require.Nil(t, err)
	require.Equal(t, edata, data)
}
