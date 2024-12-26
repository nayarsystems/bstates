package bstates

import (
	"fmt"
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
	// Create states
	fillStates := func(schema *StateSchema) (states []*State) {
		numStates := 16383
		for i := 0; i < numStates; i++ {
			state, err := schema.CreateState()
			require.NoError(t, err)
			err = state.Set("F_COUNTER", i)
			require.NoError(t, err)
			err = state.Set("F_ZEROS", 0)
			require.NoError(t, err)
			states = append(states, state)
		}
		return states
	}

	schema := testPipelineComparativeCreateSchema(t, "")
	zschema := testPipelineComparativeCreateSchema(t, "z")
	zstdschema := testPipelineComparativeCreateSchema(t, "zstd")
	tzschema := testPipelineComparativeCreateSchema(t, "t:z")
	tzstdschema := testPipelineComparativeCreateSchema(t, "t:zstd")

	states := fillStates(schema)
	zstates := fillStates(zschema)
	zstdstates := fillStates(zstdschema)
	tzstates := fillStates(tzschema)
	tzstdstates := fillStates(tzstdschema)

	checkPipeline := func(states []*State, schema *StateSchema) (edata []byte) {
		// Encode
		queue := CreateStateQueue(schema)
		err := queue.PushAll(states)
		require.NoError(t, err)
		edata, err = queue.Encode()
		require.NoError(t, err)

		// Decode
		err = queue.Decode(edata)
		require.NoError(t, err)
		dstates, err := queue.GetStates()
		require.NoError(t, err)

		// Check states == decoded
		testEqualStates(t, states, dstates)
		return
	}

	// No compression
	data := checkPipeline(states, schema)

	// Encode with z
	zdata := checkPipeline(zstates, zschema)

	// Encode with zstd
	zstddata := checkPipeline(zstdstates, zstdschema)

	// Encode with t:z
	tzdata := checkPipeline(tzstates, tzschema)

	// Encode with t:zstd
	tzstddata := checkPipeline(tzstdstates, tzstdschema)

	// -----------------

	fmt.Printf("PipelineComparative: %d states, %d bytes, %d bytes (z), %d bytes (zstd), %d bytes (t:z), %d bytes (t:zstd)\n",
		len(states), len(data), len(zdata), len(zstddata), len(tzdata), len(tzstddata))

	require.Less(t, len(zdata), len(data))
	require.Less(t, len(zstddata), len(data))
	require.Less(t, len(tzdata), len(zdata))
	require.Less(t, len(tzstddata), len(zstddata))

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

func Test_DecodeFromEmpty_NoCompression(t *testing.T) {
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

	queue := CreateStateQueue(schema)

	err = queue.Decode([]byte{})
	require.NoError(t, err)
}

func Test_DecodeFromEmpty_WithCompression(t *testing.T) {
	schema, err := CreateStateSchema(&StateSchemaParams{
		Fields: []StateField{
			{
				Name: "F_COUNTER",
				Type: T_UINT,
				Size: 32,
			},
		},
		EncoderPipeline: "z",
	})
	require.NoError(t, err)

	queue := CreateStateQueue(schema)

	err = queue.Decode([]byte{})
	require.NoError(t, err)
}

func Test_EncodeDecodeEmpty_NoTransposition(t *testing.T) {
	schema, err := CreateStateSchema(&StateSchemaParams{
		Fields: []StateField{
			{
				Name: "F_COUNTER",
				Type: T_UINT,
				Size: 32,
			},
		},
		EncoderPipeline: "z",
	})
	require.NoError(t, err)

	queue := CreateStateQueue(schema)
	data, err := queue.Encode()
	require.NoError(t, err)

	err = queue.Decode(data)
	require.NoError(t, err)
}

func Test_EncodeDecodeEmpty_Transposition(t *testing.T) {
	schema, err := CreateStateSchema(&StateSchemaParams{
		Fields: []StateField{
			{
				Name: "F_COUNTER",
				Type: T_UINT,
				Size: 32,
			},
		},
		EncoderPipeline: "t:z",
	})
	require.NoError(t, err)

	queue := CreateStateQueue(schema)
	data, err := queue.Encode()
	require.NoError(t, err)

	err = queue.Decode(data)
	require.NoError(t, err)
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

func Test_StateBufferIterFrom(t *testing.T) {
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
	queue.StateBufferIterFrom(1, func(stateBuffer []byte) (end bool) {
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
	require.Equal(t, uint64(1), v)

	v, err = states[1].Get("F_COUNTER")
	require.NoError(t, err)
	require.Equal(t, uint64(3), v)

}

func Test_GetNumStates_GetStateAt(t *testing.T) {
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

	numStates := queue.GetNumStates()
	require.Equal(t, 2, numStates)

	state, err := queue.GetStateAt(1)
	require.NoError(t, err)
	v, err := state.Get("F_COUNTER")
	require.NoError(t, err)
	require.Equal(t, uint64(2), v)

	_, err = queue.GetStateAt(2)
	require.Error(t, err)
}

func testPipelineComparativeCreateSchema(t *testing.T, encoderPipeline string) *StateSchema {
	s, err := CreateStateSchema(&StateSchemaParams{
		Fields: []StateField{
			{
				Name: "F_COUNTER",
				Type: T_UINT,
				Size: 14,
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
