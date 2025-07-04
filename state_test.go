package bstates

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
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
			{
				Name:         "F_FIXED",
				DefaultValue: -5.12,
				Type:         T_FIXED,
				Size:         10,
				Decimals:     2,
			},
			{
				Name:         "F_UFIXED",
				DefaultValue: 10.23,
				Type:         T_UFIXED,
				Size:         10,
				Decimals:     2,
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
	state1.Set("F_FIXED", 5.11)
	state1.Set("F_UFIXED", 10.22)

	data, err := StatesToMsiStates([]*State{state0, state1})
	require.Nil(t, err)

	edata := []map[string]interface{}{
		{
			"F_FLOAT32": float32(1.5),
			"F_INT32":   -1,
			"F_FIXED":   -5.12,
			"F_UFIXED":  10.23,
			"TYPE":      "TYPE B",
		},
		{
			"F_FLOAT32": float32(2.7),
			"F_INT32":   -2,
			"F_FIXED":   5.11,
			"F_UFIXED":  10.22,
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
				Name:         "F_FIXED",
				DefaultValue: -5.12,
				Type:         T_FIXED,
				Size:         10,
				Decimals:     2,
			},
			{
				Name:         "F_UFIXED",
				DefaultValue: 10.23,
				Type:         T_UFIXED,
				Size:         10,
				Decimals:     2,
			},
			{
				Name:         "F_STRING_BUFFER",
				DefaultValue: []byte{'h', 'e', 'l', 'l', 'o'},
				Type:         T_BUFFER,
				Size:         40, // 5 bytes = 40 bits
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
	state1.Set("F_FIXED", 5.11)
	data, err = GetDeltaMsiState(state0, state1)
	require.Nil(t, err)
	require.Equal(t, map[string]interface{}{
		"F_FIXED": 5.11,
	}, data)

	state0 = state1.GetCopy()
	state1.Set("F_UFIXED", 10.22)
	data, err = GetDeltaMsiState(state0, state1)
	require.Nil(t, err)
	require.Equal(t, map[string]interface{}{
		"F_UFIXED": 10.22,
	}, data)

	state0 = state1.GetCopy()
	edata = map[string]interface{}{}
	data, err = GetDeltaMsiState(state0, state1)
	require.Nil(t, err)
	require.Equal(t, edata, data)
}

func Test_FixedPoint(t *testing.T) {
	schema, err := CreateStateSchema(&StateSchemaParams{
		Fields: []StateField{
			{
				Name:         "negative",
				DefaultValue: -5.10,
				Type:         T_FIXED,
				Size:         10,
				Decimals:     2,
			},
			{
				Name:         "positive",
				DefaultValue: 5.10,
				Type:         T_FIXED,
				Size:         10,
				Decimals:     2,
			},
			{
				Name:         "unsigned",
				DefaultValue: 10.23,
				Type:         T_UFIXED,
				Size:         10,
				Decimals:     2,
			},
		},
	})
	require.Nil(t, err)
	state0, err := CreateState(schema)
	require.NoError(t, err)

	v, err := state0.Get("negative")
	require.NoError(t, err)
	require.Equal(t, -5.10, v)

	v, err = state0.Get("positive")
	require.NoError(t, err)
	require.Equal(t, 5.10, v)

	v, err = state0.Get("unsigned")
	require.NoError(t, err)
	require.Equal(t, 10.23, v)

	err = state0.Set("negative", -5.12)
	require.NoError(t, err)

	err = state0.Set("positive", 5.11)
	require.NoError(t, err)

	err = state0.Set("unsigned", 10.22)
	require.NoError(t, err)

	v, err = state0.Get("negative")
	require.NoError(t, err)
	require.Equal(t, -5.12, v)

	v, err = state0.Get("positive")
	require.NoError(t, err)
	require.Equal(t, 5.11, v)

	v, err = state0.Get("unsigned")
	require.NoError(t, err)
	require.Equal(t, 10.22, v)

	state0Raw, err := state0.Encode()
	require.NoError(t, err)

	state1, err := CreateState(schema)
	require.NoError(t, err)

	err = state1.Decode(state0Raw)
	require.NoError(t, err)

	v, err = state1.Get("negative")
	require.NoError(t, err)
	require.Equal(t, -5.12, v)

	v, err = state1.Get("positive")
	require.NoError(t, err)
	require.Equal(t, 5.11, v)

	v, err = state1.Get("unsigned")
	require.NoError(t, err)
	require.Equal(t, 10.22, v)

	// Try to encode real numbers within the valid range for fixed point of size 10 and 2 decimals
	// (2's complement range: [-5.12, 5.11])
	err = state1.Set("negative", -5.12)
	require.NoError(t, err)

	err = state1.Set("positive", 5.11)
	require.NoError(t, err)

	// (unsigned range: [0, 10.23])
	err = state1.Set("unsigned", 10.23)
	require.NoError(t, err)

	// If we get the value back, no error is expected since it has not been encoded yet
	v, err = state1.Get("negative")
	require.NoError(t, err)
	require.Equal(t, -5.12, v)

	v, err = state1.Get("positive")
	require.NoError(t, err)
	require.Equal(t, 5.11, v)

	v, err = state1.Get("unsigned")
	require.NoError(t, err)
	require.Equal(t, 10.23, v)

	// Now let's encode it. No error is expected, but wrong values will be retrived when decoding
	state1Raw, err := state1.Encode()
	require.NoError(t, err)

	state2, err := CreateState(schema)
	require.NoError(t, err)
	err = state2.Decode(state1Raw)
	require.NoError(t, err)

	v, err = state2.Get("negative")
	require.NoError(t, err)
	require.NotEqual(t, -5.13, v) // The value could not be encoded, so it should not be equal to -5.13

	v, err = state2.Get("positive")
	require.NoError(t, err)
	require.NotEqual(t, 5.12, v) // The value could not be encoded, so it should not be equal to 5.12

	v, err = state2.Get("unsigned")
	require.NoError(t, err)
	require.NotEqual(t, 10.24, v) // The value could not be encoded, so it should not be equal to 10.24
}

func Test_SameValue(t *testing.T) {
	schema, err := CreateStateSchema(&StateSchemaParams{
		Fields: []StateField{
			{
				Name:         "f_ufixed",
				DefaultValue: 0.020281571796474065,
				Type:         T_UFIXED,
				Size:         16,
				Decimals:     2,
			},
			{
				Name:         "f_fixed",
				DefaultValue: -0.020281571796474065,
				Type:         T_FIXED,
				Size:         16,
				Decimals:     2,
			},
		},
	})
	require.Nil(t, err)
	state0, _ := CreateState(schema)

	same, err := state0.Same("f_ufixed", 0.017905443709534466)
	require.NoError(t, err)
	require.True(t, same) // 0.02 == 0.02

	same, err = state0.Same("f_ufixed", 0.020281571796474065)
	require.NoError(t, err)
	require.True(t, same) // 0.02 == 0.02

	same, err = state0.Same("f_ufixed", 0.030281571796474065)
	require.NoError(t, err)
	require.False(t, same) // 0.02 != 0.03

	same, err = state0.Same("f_fixed", -0.017905443709534466)
	require.NoError(t, err)
	require.True(t, same) // -0.02 == -0.02

	same, err = state0.Same("f_fixed", -0.020281571796474065)
	require.NoError(t, err)
	require.True(t, same) // -0.02 == -0.02

	same, err = state0.Same("f_fixed", -0.030281571796474065)
	require.NoError(t, err)
	require.False(t, same) // -0.02 != -0.03
}
