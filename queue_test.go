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
	require.NoError(t, err)

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
	require.NoError(t, err)

	maxStates := 1000
	states := []*State{}
	cstates := []*State{}

	for i := 0; i < maxStates; i++ {
		state, err := schema.CreateState()
		require.NoError(t, err)
		err = state.Set("F_BOOL", i&1)
		require.NoError(t, err)
		states = append(states, state)

		state, err = cschema.CreateState()
		require.NoError(t, err)
		err = state.Set("F_BOOL", i&1)
		require.NoError(t, err)
		cstates = append(cstates, state)
	}

	cqueue := CreateStateQueue(cschema)
	err = cqueue.PushAll(cstates)
	require.NoError(t, err)
	cdata, err := cqueue.Encode()
	require.NoError(t, err)

	queue := CreateStateQueue(schema)
	err = queue.PushAll(states)
	require.NoError(t, err)
	data, err := queue.Encode()
	require.NoError(t, err)

	err = cqueue.Decode(cdata)
	require.NoError(t, err)
	err = queue.Decode(data)
	require.NoError(t, err)

	cdstates, err := cqueue.GetStates()
	require.NoError(t, err)
	dstates, err := queue.GetStates()
	require.NoError(t, err)

	v, err := cdstates[0].Get("F_BOOL")
	require.NoError(t, err)
	require.False(t, v.(bool))

	v, err = cdstates[1].Get("F_BOOL")
	require.NoError(t, err)
	require.True(t, v.(bool))

	msicdstates, err := StatesToMsiStates(cdstates)
	require.NoError(t, err)
	msidstates, err := StatesToMsiStates(dstates)
	require.NoError(t, err)
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
		require.NoError(t, err)
		err = state.Set("F_COUNTER", i)
		require.NoError(t, err)
		err = state.Set("F_ZEROS", 0)
		require.NoError(t, err)
		states = append(states, state)

		zstate, err := zschema.CreateState()
		require.NoError(t, err)
		err = zstate.Set("F_COUNTER", i)
		require.NoError(t, err)
		err = zstate.Set("F_ZEROS", 0)
		require.NoError(t, err)
		zstates = append(zstates, zstate)

		tzstate, err := tzschema.CreateState()
		require.NoError(t, err)
		err = tzstate.Set("F_COUNTER", i)
		require.NoError(t, err)
		err = tzstate.Set("F_ZEROS", 0)
		require.NoError(t, err)
		tzstates = append(tzstates, tzstate)
	}

	// -----------------

	// Encode
	queue := CreateStateQueue(schema)
	err := queue.PushAll(states)
	require.NoError(t, err)
	data, err := queue.Encode()
	require.NoError(t, err)

	// Decode
	err = queue.Decode(data)
	require.NoError(t, err)
	dstates, err := queue.GetStates()
	require.NoError(t, err)

	// Check states == decoded
	testEqualStates(t, states, dstates)

	// -----------------

	// Encode with z
	zqueue := CreateStateQueue(zschema)
	err = zqueue.PushAll(zstates)
	require.NoError(t, err)
	zdata, err := zqueue.Encode()
	require.NoError(t, err)

	// Decode with z
	err = zqueue.Decode(zdata)
	require.NoError(t, err)
	zstates, err = zqueue.GetStates()
	require.NoError(t, err)

	// Check states == decoded states with z
	testEqualStates(t, states, zstates)

	// -----------------

	// Encode with t:z
	tzqueue := CreateStateQueue(tzschema)
	err = tzqueue.PushAll(tzstates)
	require.NoError(t, err)
	tzdata, err := tzqueue.Encode()
	require.NoError(t, err)

	// Decode with t:z
	err = tzqueue.Decode(tzdata)
	require.NoError(t, err)
	tzstates, err = tzqueue.GetStates()
	require.NoError(t, err)

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
	require.NoError(t, err)

	state0, err := schema.CreateState()
	require.NoError(t, err)
	err = state0.Set("F_COUNTER", 1)
	require.NoError(t, err)
	state1, err := schema.CreateState()
	require.NoError(t, err)
	err = state1.Set("F_COUNTER", 2)
	require.NoError(t, err)

	queue := CreateStateQueue(schema)
	err = queue.Push(state0)
	require.NoError(t, err)
	err = queue.Push(state1)
	require.NoError(t, err)

	pstate, err := queue.Pop()
	require.NoError(t, err)
	v, err := pstate.Get("F_COUNTER")
	require.NoError(t, err)
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
	require.NoError(t, err)
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
	require.NoError(t, err)

	state, err := eschema.CreateState()
	require.NoError(t, err)

	queue := CreateStateQueue(sschema)
	err = queue.Push(state)
	require.Error(t, err)
}

func Test_StateBufferIter(t *testing.T) {
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
	require.NoError(t, err)

	state0, _ := schema.CreateState()
	_ = state0.Set("F_COUNTER", 1)
	state1, _ := schema.CreateState()
	_ = state1.Set("F_COUNTER", 2)

	queue := CreateStateQueue(schema)
	_ = queue.Push(state0)
	_ = queue.Push(state1)

	tmpState, _ := schema.CreateState()
	queue.StateBufferIter(func(stateBuffer []byte) (end bool) {
		// Decode state
		err = tmpState.Decode(stateBuffer)
		if err != nil {
			return true
		}

		// Increment F_COUNTER value
		vi, err := tmpState.Get("F_COUNTER")
		require.NoError(t, err)
		v := vi.(uint64)
		v++
		err = tmpState.Set("F_COUNTER", v)
		require.NoError(t, err)

		// Encode state
		err = tmpState.EncodeTo(stateBuffer)
		return err != nil
	})

	states, err := queue.GetStates()
	require.NoError(t, err)
	require.Equal(t, 2, len(states))

	v, err := states[0].Get("F_COUNTER")
	require.NoError(t, err)
	require.Equal(t, uint64(2), v)

	v, err = states[1].Get("F_COUNTER")
	require.NoError(t, err)
	require.Equal(t, uint64(3), v)

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
	require.NoError(t, err)
	return s
}

func testEqualStates(t *testing.T, e0 []*State, e1 []*State) {
	msie0, err := StatesToMsiStates(e0)
	require.NoError(t, err)
	msie1, err := StatesToMsiStates(e1)
	require.NoError(t, err)
	require.Equal(t, msie0, msie1)
}
