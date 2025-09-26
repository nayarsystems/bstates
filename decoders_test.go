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
		DecoderIntMaps: map[string]map[int64]any{
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

func Test_FlagsDecoder(t *testing.T) {
	schema, err := CreateStateSchema(&StateSchemaParams{
		Fields: []StateField{
			{
				Name:         "status_flags",
				DefaultValue: uint64(0),
				Type:         T_UINT,
				Size:         8,
			},
		},
		DecodedFields: []DecodedStateField{
			{
				Name: "status",
				Decoder: &FlagsDecoder{
					From: "status_flags",
					Flags: map[string]uint8{
						"active":    0,
						"connected": 1,
						"error":     2,
						"synced":    3,
					},
				},
			},
		},
	})
	require.Nil(t, err)
	state, err := CreateState(schema)
	require.Nil(t, err)

	// Test empty flags (default value 0)
	v, err := state.Get("status")
	require.Nil(t, err)
	require.Equal(t, []string{}, v)

	// Test single flag
	err = state.Set("status_flags", uint64(0b0001)) // bit 0 = active
	require.Nil(t, err)
	v, err = state.Get("status")
	require.Nil(t, err)
	require.Equal(t, []string{"active"}, v)

	// Test multiple flags
	err = state.Set("status_flags", uint64(0b0101)) // bits 0,2 = active,error
	require.Nil(t, err)
	v, err = state.Get("status")
	require.Nil(t, err)
	// Order can vary since map iteration is not deterministic
	flags := v.([]string)
	require.Len(t, flags, 2)
	require.Contains(t, flags, "active")
	require.Contains(t, flags, "error")

	// Test all flags
	err = state.Set("status_flags", uint64(0b1111)) // bits 0,1,2,3
	require.Nil(t, err)
	v, err = state.Get("status")
	require.Nil(t, err)
	flags = v.([]string)
	require.Len(t, flags, 4)
	require.Contains(t, flags, "active")
	require.Contains(t, flags, "connected")
	require.Contains(t, flags, "error")
	require.Contains(t, flags, "synced")

	// Test encoding back from flags to raw value
	err = state.Set("status", []string{"connected", "synced"})
	require.Nil(t, err)
	v, err = state.Get("status_flags")
	require.Nil(t, err)
	require.Equal(t, uint64(0b1010), v) // bits 1,3 = connected,synced

	// Test encoding empty flags
	err = state.Set("status", []string{})
	require.Nil(t, err)
	v, err = state.Get("status_flags")
	require.Nil(t, err)
	require.Equal(t, uint64(0), v)
}

func Test_FlagsDecoder_NewDecoder(t *testing.T) {
	// Test creating FlagsDecoder via NewDecoder function
	params := map[string]any{
		"from": "flags_field",
		"flags": map[string]any{
			"flag1": uint8(0),
			"flag2": uint8(1),
			"flag3": uint8(7),
		},
	}

	decoder, err := NewDecoder("Flags", params)
	require.Nil(t, err)
	require.NotNil(t, decoder)

	flagsDecoder, ok := decoder.(*FlagsDecoder)
	require.True(t, ok)
	require.Equal(t, "flags_field", flagsDecoder.From)
	require.Equal(t, map[string]uint8{"flag1": 0, "flag2": 1, "flag3": 7}, flagsDecoder.Flags)
	require.Equal(t, FlagsDecoderType, flagsDecoder.Name())

	// Test GetParams
	returnedParams := flagsDecoder.GetParams()
	require.Equal(t, "flags_field", returnedParams["from"])
	require.Equal(t, map[string]uint8{"flag1": 0, "flag2": 1, "flag3": 7}, returnedParams["flags"])
}

func Test_FlagsDecoder_Errors(t *testing.T) {
	// Test invalid parameters for NewFlagsDecoder

	// Missing "from" parameter
	_, err := NewFlagsDecoder(map[string]any{
		"flags": map[string]any{"flag1": 0},
	})
	require.Error(t, err)

	// Missing "flags" parameter
	_, err = NewFlagsDecoder(map[string]any{
		"from": "test_field",
	})
	require.Error(t, err)

	// Invalid flag bit position type
	_, err = NewFlagsDecoder(map[string]any{
		"from": "test_field",
		"flags": map[string]any{
			"flag1": "invalid",
		},
	})
	require.Error(t, err)

	// Test runtime errors
	schema, err := CreateStateSchema(&StateSchemaParams{
		Fields: []StateField{
			{
				Name:         "flags_field",
				DefaultValue: uint64(0),
				Type:         T_UINT,
				Size:         8,
			},
		},
		DecodedFields: []DecodedStateField{
			{
				Name: "decoded_flags",
				Decoder: &FlagsDecoder{
					From:  "nonexistent_field",
					Flags: map[string]uint8{"flag1": 0},
				},
			},
		},
	})
	require.Nil(t, err)
	state, err := CreateState(schema)
	require.Nil(t, err)

	// Test decode with non-existent field
	_, err = state.Get("decoded_flags")
	require.Error(t, err)

	// Test encode with invalid input type (not []string)
	decoder := &FlagsDecoder{
		From:  "flags_field",
		Flags: map[string]uint8{"flag1": 0},
	}
	err = decoder.Encode(state, "not_a_slice")
	require.Error(t, err)
	require.Contains(t, err.Error(), "expected []string, got string")

	// Test encode with []any instead of []string
	err = decoder.Encode(state, []any{"flag1"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "expected []string, got []interface {}")

	// Test encode with unknown flag name
	err = decoder.Encode(state, []string{"flag1", "unknown_flag"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "unknown flag \"unknown_flag\"")
}

func Test_FlagsDecoder_BitPositionValidation(t *testing.T) {
	// Create schema with a small field (4 bits)
	schema, err := CreateStateSchema(&StateSchemaParams{
		Fields: []StateField{
			{
				Name:         "small_flags",
				DefaultValue: uint64(0),
				Type:         T_UINT,
				Size:         4, // Only 4 bits: positions 0,1,2,3 are valid
			},
		},
		DecodedFields: []DecodedStateField{
			{
				Name: "decoded_small_flags",
				Decoder: &FlagsDecoder{
					From: "small_flags",
					Flags: map[string]uint8{
						"flag0": 0, // Valid
						"flag1": 1, // Valid
						"flag3": 3, // Valid (last valid position)
						"flag4": 4, // Invalid - exceeds 4-bit field
					},
				},
			},
		},
	})
	require.Nil(t, err)
	state, err := CreateState(schema)
	require.Nil(t, err)

	// Test decode with bit position exceeding field size
	err = state.Set("small_flags", uint64(0b0001)) // Set bit 0
	require.Nil(t, err)

	_, err = state.Get("decoded_small_flags")
	require.Error(t, err)
	require.Contains(t, err.Error(), "flag \"flag4\" bit position 4 exceeds field size 4 bits")

	// Test encode with bit position exceeding field size
	err = state.Set("decoded_small_flags", []string{"flag4"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "flag \"flag4\" bit position 4 exceeds field size 4 bits")

	// Test that valid flags still work
	validSchema, err := CreateStateSchema(&StateSchemaParams{
		Fields: []StateField{
			{
				Name:         "valid_flags",
				DefaultValue: uint64(0),
				Type:         T_UINT,
				Size:         4,
			},
		},
		DecodedFields: []DecodedStateField{
			{
				Name: "decoded_valid_flags",
				Decoder: &FlagsDecoder{
					From: "valid_flags",
					Flags: map[string]uint8{
						"flag0": 0, // Valid
						"flag1": 1, // Valid
						"flag3": 3, // Valid (last valid position for 4-bit field)
					},
				},
			},
		},
	})
	require.Nil(t, err)
	validState, err := CreateState(validSchema)
	require.Nil(t, err)

	// Test successful operations with valid bit positions
	err = validState.Set("decoded_valid_flags", []string{"flag0", "flag3"})
	require.Nil(t, err)

	v, err := validState.Get("valid_flags")
	require.Nil(t, err)
	require.Equal(t, uint64(0b1001), v) // bits 0,3 set

	v, err = validState.Get("decoded_valid_flags")
	require.Nil(t, err)
	flags := v.([]string)
	require.Len(t, flags, 2)
	require.Contains(t, flags, "flag0")
	require.Contains(t, flags, "flag3")
}
