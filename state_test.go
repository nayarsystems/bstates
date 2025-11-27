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

func Test_GetDeltaMsiState_WithAliases(t *testing.T) {
	// Test that GetDeltaMsiState includes aliases for changed fields
	schema, err := CreateStateSchema(&StateSchemaParams{
		Fields: []StateField{
			{
				Name:         "temperature",
				Aliases:      []string{"temp", "t"},
				DefaultValue: 0,
				Type:         T_INT,
				Size:         16,
			},
			{
				Name:         "pressure",
				DefaultValue: 0,
				Type:         T_INT,
				Size:         16,
				// No aliases for comparison
			},
		},
		DecodedFields: []DecodedStateField{
			{
				Name:    "temp_celsius",
				Aliases: []string{"temp_c", "celsius"},
				Decoder: &IntMapDecoder{From: "temperature", MapId: "TEMP_MAP"},
			},
		},
		DecoderIntMaps: map[string]map[int64]any{
			"TEMP_MAP": {25: "NORMAL", 30: "HOT"},
		},
	})
	require.Nil(t, err)

	state1, err := CreateState(schema)
	require.Nil(t, err)
	state2, err := CreateState(schema)
	require.Nil(t, err)

	// Change temperature (has aliases) and pressure (no aliases)
	state2.Set("temperature", 25)
	state2.Set("pressure", 1013)

	delta, err := GetDeltaMsiState(state1, state2)
	require.Nil(t, err)

	// Should include original field names
	require.Equal(t, 25, delta["temperature"])
	require.Equal(t, 1013, delta["pressure"])

	// Should include aliases for temperature field
	require.Equal(t, 25, delta["temp"])
	require.Equal(t, 25, delta["t"])

	// Should include decoded field and its aliases
	require.Equal(t, "NORMAL", delta["temp_celsius"])
	require.Equal(t, "NORMAL", delta["temp_c"])
	require.Equal(t, "NORMAL", delta["celsius"])

	// Pressure has no aliases, so no extra keys
	require.NotContains(t, delta, "press")
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

func Test_StateFieldAliases(t *testing.T) {
	// Create schema with fields that have aliases
	schema, err := CreateStateSchema(&StateSchemaParams{
		Fields: []StateField{
			{
				Name:         "temperature",
				Aliases:      []string{"temp", "t"},
				DefaultValue: 0,
				Type:         T_INT,
				Size:         16,
			},
			{
				Name:         "pressure",
				Aliases:      []string{"press", "p"},
				DefaultValue: 0,
				Type:         T_UINT,
				Size:         16,
			},
			{
				Name:         "status",
				DefaultValue: false,
				Type:         T_BOOL,
				// No aliases for this field
			},
			{
				Name:         "dateRaw",
				DefaultValue: false,
				Type:         T_UINT,
				Size:         32,
				Aliases:      []string{"timestampRaw", "datetimeRaw"},
			},
		},
		DecodedFields: []DecodedStateField{
			{
				Name:    "date",
				Aliases: []string{"timestamp", "datetime"},
				Decoder: &NumberToUnixTsMsDecoder{
					From:   "dateRaw",
					Year:   2020,
					Factor: 1,
				},
			},
		},
	})
	require.NoError(t, err)
	state, err := CreateState(schema)
	require.NoError(t, err)

	// Set some values
	err = state.Set("temperature", 25)
	require.NoError(t, err)
	err = state.Set("pressure", 1013)
	require.NoError(t, err)
	err = state.Set("status", true)
	require.NoError(t, err)
	err = state.Set("dateRaw", 0)
	require.NoError(t, err)

	// Convert to MSI and verify aliases are present
	msi, err := state.ToMsi()
	require.NoError(t, err)

	requireEqual := func(t *testing.T, s *State, msi map[string]any, fieldName string, expected any) {
		require.Equal(t, expected, msi[fieldName])
		v, err := s.Get(fieldName)
		require.NoError(t, err)
		require.Equal(t, expected, v)
	}

	// Check that original field names are present
	requireEqual(t, state, msi, "temperature", 25)
	requireEqual(t, state, msi, "pressure", uint64(1013))
	requireEqual(t, state, msi, "status", true)
	requireEqual(t, state, msi, "dateRaw", uint64(0))
	requireEqual(t, state, msi, "date", uint64(time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC).UnixMilli()))

	// Check that aliases are present with the same values
	requireEqual(t, state, msi, "temp", 25)
	requireEqual(t, state, msi, "t", 25)
	requireEqual(t, state, msi, "press", uint64(1013))
	requireEqual(t, state, msi, "p", uint64(1013))
	requireEqual(t, state, msi, "timestampRaw", uint64(0))
	requireEqual(t, state, msi, "datetimeRaw", uint64(0))
	requireEqual(t, state, msi, "timestamp", uint64(time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC).UnixMilli()))
	requireEqual(t, state, msi, "datetime", uint64(time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC).UnixMilli()))

	// Update decoder's raw field and check decoded aliases
	err = state.Set("dateRaw", 1000) // 1000 ms after epoch
	require.NoError(t, err)
	msi, err = state.ToMsi()
	require.NoError(t, err)

	requireEqual(t, state, msi, "date", uint64(time.Date(2020, time.January, 1, 0, 0, 1, 0, time.UTC).UnixMilli()))
	requireEqual(t, state, msi, "timestamp", uint64(time.Date(2020, time.January, 1, 0, 0, 1, 0, time.UTC).UnixMilli()))
	requireEqual(t, state, msi, "datetime", uint64(time.Date(2020, time.January, 1, 0, 0, 1, 0, time.UTC).UnixMilli()))

	// Update decoded field and check decoded aliases
	err = state.Set("date", uint64(time.Date(2020, time.January, 1, 0, 1, 0, 0, time.UTC).UnixMilli())) // 1 minute after epoch
	require.NoError(t, err)
	msi, err = state.ToMsi()
	require.NoError(t, err)

	requireEqual(t, state, msi, "date", uint64(time.Date(2020, time.January, 1, 0, 1, 0, 0, time.UTC).UnixMilli()))
	requireEqual(t, state, msi, "timestamp", uint64(time.Date(2020, time.January, 1, 0, 1, 0, 0, time.UTC).UnixMilli()))
	requireEqual(t, state, msi, "datetime", uint64(time.Date(2020, time.January, 1, 0, 1, 0, 0, time.UTC).UnixMilli()))

	// Check that raw field is updated accordingly
	requireEqual(t, state, msi, "dateRaw", uint64(60000)) // 60 seconds in ms

	// Update an alias of the decoded field and check decoded aliases
	err = state.Set("timestamp", uint64(time.Date(2020, time.January, 1, 1, 0, 0, 0, time.UTC).UnixMilli()))
	require.NoError(t, err)
	msi, err = state.ToMsi()
	require.NoError(t, err)

	requireEqual(t, state, msi, "date", uint64(time.Date(2020, time.January, 1, 1, 0, 0, 0, time.UTC).UnixMilli()))
	requireEqual(t, state, msi, "timestamp", uint64(time.Date(2020, time.January, 1, 1, 0, 0, 0, time.UTC).UnixMilli()))
	requireEqual(t, state, msi, "datetime", uint64(time.Date(2020, time.January, 1, 1, 0, 0, 0, time.UTC).UnixMilli()))

	// Check that raw field is updated accordingly
	requireEqual(t, state, msi, "dateRaw", uint64(3600000)) // 3600 seconds in ms

	// Update an alias of a regular field and check original field
	err = state.Set("timestampRaw", 2000) // 2000 ms after epoch
	require.NoError(t, err)
	msi, err = state.ToMsi()
	require.NoError(t, err)

	requireEqual(t, state, msi, "timestampRaw", uint64(2000))
	requireEqual(t, state, msi, "timestamp", uint64(time.Date(2020, time.January, 1, 0, 0, 2, 0, time.UTC).UnixMilli()))
	requireEqual(t, state, msi, "datetimeRaw", uint64(2000))
	requireEqual(t, state, msi, "datetime", uint64(time.Date(2020, time.January, 1, 0, 0, 2, 0, time.UTC).UnixMilli()))

	// Check total number of keys
	expectedKeys := []string{
		"temperature", "temp", "t", // regular field + aliases
		"pressure", "press", "p", // regular field + aliases
		"status",                                 // field without aliases
		"dateRaw", "timestampRaw", "datetimeRaw", // raw field and aliases
		"date", "timestamp", "datetime", // decoded field + aliases
	}
	require.Len(t, msi, len(expectedKeys))

	for _, key := range expectedKeys {
		require.Contains(t, msi, key)
	}

	// Test that aliases work for empty aliases array
	schemaNoAliases, err := CreateStateSchema(&StateSchemaParams{
		Fields: []StateField{
			{
				Name:         "simple_field",
				Aliases:      []string{}, // Empty aliases
				DefaultValue: 42,
				Type:         T_INT,
				Size:         8,
			},
		},
	})
	require.Nil(t, err)
	stateNoAliases, err := CreateState(schemaNoAliases)
	require.Nil(t, err)

	msiNoAliases, err := stateNoAliases.ToMsi()
	require.Nil(t, err)
	require.Len(t, msiNoAliases, 1)
	require.Equal(t, 42, msiNoAliases["simple_field"])
}

func Test_StateFieldAliases_MsiConversion(t *testing.T) {
	// Test that aliases are properly converted to/from MSI in StateField
	originalField := StateField{
		Name:         "test_field",
		Aliases:      []string{"alias1", "alias2", "test_alias"},
		Size:         8,
		DefaultValue: 100,
		Type:         T_UINT,
	}

	// Normalize the original field first (this is usually done during schema creation)
	err := originalField.normalize()
	require.Nil(t, err)

	// Convert to MSI
	msi, err := originalField.ToMsi()
	require.Nil(t, err)
	require.Equal(t, "test_field", msi["name"])
	require.Equal(t, []string{"alias1", "alias2", "test_alias"}, msi["aliases"])

	// Convert back from MSI
	restoredField := StateField{}
	err = restoredField.FromMsi(msi)
	require.Nil(t, err)
	require.Equal(t, originalField.Name, restoredField.Name)
	require.Equal(t, originalField.Aliases, restoredField.Aliases)
	require.Equal(t, originalField.Size, restoredField.Size)
	require.Equal(t, originalField.Type, restoredField.Type)

	// Test field with no aliases
	fieldNoAliases := StateField{
		Name:         "no_aliases",
		Size:         16,
		DefaultValue: 0,
		Type:         T_INT,
	}
	err = fieldNoAliases.normalize()
	require.Nil(t, err)

	msiNoAliases, err := fieldNoAliases.ToMsi()
	require.Nil(t, err)
	require.NotContains(t, msiNoAliases, "aliases") // Should not include aliases key when empty

	restoredNoAliases := StateField{}
	err = restoredNoAliases.FromMsi(msiNoAliases)
	require.Nil(t, err)
	require.Equal(t, fieldNoAliases.Name, restoredNoAliases.Name)
	require.Nil(t, restoredNoAliases.Aliases) // Should be nil when not present in MSI
}

func Test_StateField_InvalidAliasesType_Errors(t *testing.T) {
	// Test StateField with invalid aliases type (not array)
	invalidMsi := map[string]any{
		"name":    "test_field",
		"type":    "uint",
		"size":    8,
		"aliases": "not_an_array", // Invalid type
	}

	field := StateField{}
	err := field.FromMsi(invalidMsi)
	require.Error(t, err)
	require.Contains(t, err.Error(), "aliases field must be a string array, got string")

	// Test StateField with invalid alias element type (not string)
	// Note: ei.N() can convert numbers to strings, so we use a type that can't be converted
	invalidElementMsi := map[string]any{
		"name":    "test_field",
		"type":    "uint",
		"size":    8,
		"aliases": []any{"valid_alias", map[string]string{"key": "value"}, "another_valid"}, // Element map can't be string
	}

	field2 := StateField{}
	err = field2.FromMsi(invalidElementMsi)
	require.Error(t, err)
	require.Contains(t, err.Error(), "alias at index 1 must be a string")

	// Test StateField with mixed valid and invalid types in array
	// Use a struct which can't be converted to string by ei.N()
	mixedInvalidMsi := map[string]any{
		"name":    "test_field",
		"type":    "uint",
		"size":    8,
		"aliases": []any{struct{}{}, "valid_alias"}, // Element struct{} can't be string
	}

	field3 := StateField{}
	err = field3.FromMsi(mixedInvalidMsi)
	require.Error(t, err)
	require.Contains(t, err.Error(), "alias at index 0 must be a string")
}

func Test_DecodedStateField_Aliases_MsiConversion(t *testing.T) {
	// Test that aliases are properly converted to/from MSI in DecodedStateField
	originalDecodedField := DecodedStateField{
		Name:    "date",
		Aliases: []string{"datetime", "timestamp", "ts"},
		Decoder: &NumberToUnixTsMsDecoder{
			From:   "temp_raw",
			Year:   2020,
			Factor: 1000,
		},
	}

	// Convert to MSI
	msi, err := originalDecodedField.ToMsi()
	require.Nil(t, err)
	require.Equal(t, "date", msi["name"])
	require.Equal(t, []string{"datetime", "timestamp", "ts"}, msi["aliases"])
	require.Equal(t, string(NumberToUnixTsMsDecoderType), msi["decoder"])

	// Convert back from MSI
	restoredDecodedField := DecodedStateField{}
	err = restoredDecodedField.FromMsi(msi)
	require.Nil(t, err)
	require.Equal(t, originalDecodedField.Name, restoredDecodedField.Name)
	require.Equal(t, originalDecodedField.Aliases, restoredDecodedField.Aliases)
	require.Equal(t, originalDecodedField.Decoder.Name(), restoredDecodedField.Decoder.Name())

	// Test decoded field with no aliases
	decodedFieldNoAliases := DecodedStateField{
		Name: "simple_decoded",
		Decoder: &BufferToStringDecoder{
			From: "buffer_field",
		},
	}

	msiNoAliases, err := decodedFieldNoAliases.ToMsi()
	require.Nil(t, err)
	require.NotContains(t, msiNoAliases, "aliases") // Should not include aliases key when empty

	restoredNoAliases := DecodedStateField{}
	err = restoredNoAliases.FromMsi(msiNoAliases)
	require.Nil(t, err)
	require.Equal(t, decodedFieldNoAliases.Name, restoredNoAliases.Name)
	require.Nil(t, restoredNoAliases.Aliases) // Should be nil when not present in MSI
}

func Test_DecodedStateField_InvalidAliasesType_Errors(t *testing.T) {
	// Test DecodedStateField with invalid aliases type (not array)
	invalidMsi := map[string]any{
		"name":    "decoded_field",
		"decoder": "BufferToString",
		"params":  map[string]any{"from": "buffer_field"},
		"aliases": 42, // Invalid type
	}

	decodedField := DecodedStateField{}
	err := decodedField.FromMsi(invalidMsi)
	require.Error(t, err)
	require.Contains(t, err.Error(), "aliases field must be a string array, got int")

	// Test DecodedStateField with invalid alias element type (not string)
	// Use a type that can't be converted to string by ei.N()
	invalidElementMsi := map[string]any{
		"name":    "decoded_field",
		"decoder": "BufferToString",
		"params":  map[string]any{"from": "buffer_field"},
		"aliases": []any{"valid_alias", []int{1, 2, 3}, "another_valid"}, // Element slice can't be string
	}

	decodedField2 := DecodedStateField{}
	err = decodedField2.FromMsi(invalidElementMsi)
	require.Error(t, err)
	require.Contains(t, err.Error(), "alias at index 1 must be a string")

	// Test DecodedStateField with array of non-strings
	nonStringArrayMsi := map[string]any{
		"name":    "decoded_field",
		"decoder": "BufferToString",
		"params":  map[string]any{"from": "buffer_field"},
		"aliases": []any{make(chan int), "valid_alias"}, // Element channel can't be string
	}

	decodedField3 := DecodedStateField{}
	err = decodedField3.FromMsi(nonStringArrayMsi)
	require.Error(t, err)
	require.Contains(t, err.Error(), "alias at index 0 must be a string")
}

func Test_StateField_FromMsi_AliasesReset(t *testing.T) {
	// Test that Aliases field is properly reset when deserializing from MSI
	field := StateField{Aliases: []string{"old_alias1", "old_alias2"}}

	// MSI without aliases field - should reset to nil
	err := field.FromMsi(map[string]any{"name": "test", "type": "int", "size": 8})
	require.Nil(t, err)
	require.Nil(t, field.Aliases)

	// MSI with empty aliases array - should set to empty slice
	err = field.FromMsi(map[string]any{"name": "test2", "type": "int", "size": 8, "aliases": []string{}})
	require.Nil(t, err)
	require.Equal(t, []string{}, field.Aliases)
}

func Test_DecodedStateField_FromMsi_AliasesReset(t *testing.T) {
	// Test that Aliases field is properly reset when deserializing from MSI
	decodedField := DecodedStateField{Aliases: []string{"old_alias1", "old_alias2"}}

	// MSI without aliases field - should reset to nil
	err := decodedField.FromMsi(map[string]any{
		"name": "test", "decoder": "BufferToString", "params": map[string]any{"from": "buffer"},
	})
	require.Nil(t, err)
	require.Nil(t, decodedField.Aliases)

	// MSI with empty aliases array - should set to empty slice
	err = decodedField.FromMsi(map[string]any{
		"name": "test2", "decoder": "BufferToString", "params": map[string]any{"from": "buffer"}, "aliases": []string{},
	})
	require.Nil(t, err)
	require.Equal(t, []string{}, decodedField.Aliases)
}

func Test_StateField_FromMsi_DecimalsReset(t *testing.T) {
	// Test that Decimals field is properly reset when deserializing different field types
	field := StateField{Decimals: 3} // Pre-existing decimals value

	// MSI for non-fixed type - should reset Decimals to 0
	err := field.FromMsi(map[string]any{"name": "test", "type": "int", "size": 8})
	require.Nil(t, err)
	require.Equal(t, uint(0), field.Decimals)

	// MSI for fixed type - should set new decimals value
	err = field.FromMsi(map[string]any{"name": "test", "type": "ufixed", "size": 16, "decimals": 2})
	require.Nil(t, err)
	require.Equal(t, uint(2), field.Decimals)
}

func Test_State_Set_BufferTruncation(t *testing.T) {
	// Test that T_BUFFER fields return error but still write data (truncation happens on encode/decode)
	schema, err := CreateStateSchema(&StateSchemaParams{
		Fields: []StateField{
			{
				Name: "small_buffer",
				Type: T_BUFFER,
				Size: 16, // 2 bytes max
			},
		},
	})
	require.Nil(t, err)

	state, err := CreateState(schema)
	require.Nil(t, err)

	// Try to set 3 bytes (should error but still write the full value)
	oversizedData := []byte{0xAA, 0xBB, 0xCC}
	err = state.Set("small_buffer", oversizedData)
	require.Error(t, err, "Should return error for oversized buffer")
	require.Contains(t, err.Error(), "exceeds field capacity")

	// The value is stored as-is (not truncated yet)
	value, err := state.Get("small_buffer")
	require.Nil(t, err)
	fullBytes, ok := value.([]byte)
	require.True(t, ok)
	require.Equal(t, oversizedData, fullBytes, "Value stored as-is before encode/decode")

	// However, when we encode and decode, truncation should occur
	encodedData, err := state.Encode()
	require.Nil(t, err)

	// Create a new state and decode the data
	newState, err := CreateState(schema)
	require.Nil(t, err)
	err = newState.Decode(encodedData)
	require.Nil(t, err)

	// Now the truncated value should be visible
	value, err = newState.Get("small_buffer")
	require.Nil(t, err)
	truncatedBytes, ok := value.([]byte)
	require.True(t, ok)
	// Should contain only the first 2 bytes due to truncation during encode/decode
	require.Equal(t, []byte{0xAA, 0xBB}, truncatedBytes, "Value truncated after encode/decode")

	// Test with valid size (should not error)
	validData := []byte{0xDD, 0xEE}
	err = state.Set("small_buffer", validData)
	require.Nil(t, err, "Should not error for valid size")

	value, err = state.Get("small_buffer")
	require.Nil(t, err)
	resultBytes, ok := value.([]byte)
	require.True(t, ok)
	require.Equal(t, validData, resultBytes)
}
