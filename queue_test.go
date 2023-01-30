package bstates

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Compression(t *testing.T) {
	cschema, err := CreateStateSchema(
		&StateSchemaParams{
			Fields: []StateField{
				{
					Name: "F_BOOL",
					Type: T_BOOL,
				},
			},
			EncoderPipeline: "t:z",
		},
	)
	require.Nil(t, err)

	schema, err := CreateStateSchema(
		&StateSchemaParams{
			Fields: []StateField{
				{
					Name: "F_BOOL",
					Type: T_BOOL,
				},
			},
			EncoderPipeline: "",
		},
	)
	require.Nil(t, err)

	maxStates := 1000
	states := []*State{}
	cstates := []*State{}

	for i := 0; i < maxStates; i++ {
		state, err := schema.CreateState()
		require.Nil(t, err)
		err = state.Set("F_BOOL", i&1)
		require.Nil(t, err)
		states = append(states, state)

		state, err = cschema.CreateState()
		require.Nil(t, err)
		err = state.Set("F_BOOL", i&1)
		require.Nil(t, err)
		cstates = append(cstates, state)
	}

	cstack := CreateStateQueue(cschema)
	err = cstack.PushAll(cstates)
	require.Nil(t, err)
	cdata, err := cstack.Encode()
	require.Nil(t, err)

	stack := CreateStateQueue(schema)
	err = stack.PushAll(states)
	require.Nil(t, err)
	data, err := stack.Encode()
	require.Nil(t, err)

	err = cstack.Decode(cdata)
	require.Nil(t, err)
	err = stack.Decode(data)
	require.Nil(t, err)

	cdstates, err := cstack.GetStates()
	require.Nil(t, err)
	dstates, err := stack.GetStates()
	require.Nil(t, err)

	v, err := cdstates[0].Get("F_BOOL")
	require.Nil(t, err)
	require.False(t, v.(bool))

	v, err = cdstates[1].Get("F_BOOL")
	require.Nil(t, err)
	require.True(t, v.(bool))

	msicdstates, err := StatesToMsiStates(cdstates)
	require.Nil(t, err)
	msidstates, err := StatesToMsiStates(dstates)
	require.Nil(t, err)
	require.Equal(t, msicdstates, msidstates)

	require.Less(t, len(cdata), len(data))
}

func Test_PipelineComparative(t *testing.T) {

	schema := testPipelineComparativeCreateSchema(t, "")
	zschema := testPipelineComparativeCreateSchema(t, "z")
	tzschema := testPipelineComparativeCreateSchema(t, "t:z")

	numStates := 2048 // 2^11 - 1
	states := []*State{}
	zstates := []*State{}
	tzstates := []*State{}

	// Create states

	for i := 0; i < numStates; i++ {
		state, err := schema.CreateState()
		require.Nil(t, err)
		err = state.Set("F_COUNTER", i)
		require.Nil(t, err)
		err = state.Set("F_ZEROS", 0)
		require.Nil(t, err)
		states = append(states, state)

		zstate, err := zschema.CreateState()
		require.Nil(t, err)
		err = zstate.Set("F_COUNTER", i)
		require.Nil(t, err)
		err = zstate.Set("F_ZEROS", 0)
		require.Nil(t, err)
		zstates = append(zstates, zstate)

		tzstate, err := tzschema.CreateState()
		require.Nil(t, err)
		err = tzstate.Set("F_COUNTER", i)
		require.Nil(t, err)
		err = tzstate.Set("F_ZEROS", 0)
		require.Nil(t, err)
		tzstates = append(tzstates, tzstate)
	}

	// -----------------

	// Encode
	stack := CreateStateQueue(schema)
	err := stack.PushAll(states)
	require.Nil(t, err)
	data, err := stack.Encode()
	require.Nil(t, err)

	// Decode
	err = stack.Decode(data)
	require.Nil(t, err)
	dstates, err := stack.GetStates()
	require.Nil(t, err)

	// Check states == decoded
	testEqualStates(t, states, dstates)

	// -----------------

	// Encode with z
	zstack := CreateStateQueue(zschema)
	err = zstack.PushAll(zstates)
	require.Nil(t, err)
	zdata, err := zstack.Encode()
	require.Nil(t, err)

	// Decode with z
	err = zstack.Decode(zdata)
	require.Nil(t, err)
	zstates, err = zstack.GetStates()
	require.Nil(t, err)

	// Check states == decoded states with z
	testEqualStates(t, states, zstates)

	// -----------------

	// Encode with t:z
	tzstack := CreateStateQueue(tzschema)
	err = tzstack.PushAll(tzstates)
	require.Nil(t, err)
	tzdata, err := tzstack.Encode()
	require.Nil(t, err)

	// Decode with t:z
	err = tzstack.Decode(tzdata)
	require.Nil(t, err)
	tzstates, err = tzstack.GetStates()
	require.Nil(t, err)

	// Check states == decoded states with t:z
	testEqualStates(t, states, tzstates)

	// ----------------

	// Check sizes of encoded data with no compression, "z" and "t:z"
	require.Less(t, len(zdata), len(data))
	require.Less(t, len(tzdata), len(zdata))

}

func Test_PushPop(t *testing.T) {
	schema, err := CreateStateSchema(&StateSchemaParams{
		Fields: []StateField{
			{
				Name: "F_COUNTER",
				Type: T_UINT,
				Size: 32,
			},
		},
		EncoderPipeline: "",
	})
	require.Nil(t, err)

	state0, err := schema.CreateState()
	require.Nil(t, err)
	err = state0.Set("F_COUNTER", 1)
	require.Nil(t, err)
	state1, err := schema.CreateState()
	require.Nil(t, err)
	err = state1.Set("F_COUNTER", 2)
	require.Nil(t, err)

	stack := CreateStateQueue(schema)
	err = stack.Push(state0)
	require.Nil(t, err)
	err = stack.Push(state1)
	require.Nil(t, err)

	pstate, err := stack.Pop()
	require.Nil(t, err)
	v, err := pstate.Get("F_COUNTER")
	require.Nil(t, err)
	require.Equal(t, uint64(1), v)

}

func Test_InvalidPushState(t *testing.T) {
	eschema, err := CreateStateSchema(
		&StateSchemaParams{
			Fields: []StateField{
				{
					Name: "F_BOOL",
					Type: T_BOOL,
				},
			},
			EncoderPipeline: "t:z",
		},
	)
	require.Nil(t, err)
	sschema, err := CreateStateSchema(
		&StateSchemaParams{
			Fields: []StateField{
				{
					Name: "F_BOOL",
					Type: T_BOOL,
				},
			},
			EncoderPipeline: "z",
		},
	)
	require.Nil(t, err)

	state, err := eschema.CreateState()
	require.Nil(t, err)

	stack := CreateStateQueue(sschema)
	err = stack.Push(state)
	require.NotNil(t, err)
}

func testPipelineComparativeCreateSchema(t *testing.T, encoderPipeline string) *StateSchema {
	s, err := CreateStateSchema(&StateSchemaParams{
		Fields: []StateField{
			{
				Name: "F_COUNTER",
				Type: T_UINT,
				Size: 11,
			},
			{
				Name: "F_ZEROS",
				Type: T_UINT,
				Size: 64,
			},
		},
		EncoderPipeline: encoderPipeline,
	})
	require.Nil(t, err)
	return s
}

func testEqualStates(t *testing.T, e0 []*State, e1 []*State) {
	msie0, err := StatesToMsiStates(e0)
	require.Nil(t, err)
	msie1, err := StatesToMsiStates(e1)
	require.Nil(t, err)
	require.Equal(t, msie0, msie1)
}
