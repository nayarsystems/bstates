package bstates

import (
	"testing"

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
		MappedFields: map[string]MappedStateField{
			"TYPE": {
				From:  "F_INT32",
				MapId: "TYPE_MAP",
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
	edata = map[string]interface{}{}
	data, err = GetDeltaMsiState(state0, state1)
	require.Nil(t, err)
	require.Equal(t, edata, data)
}
