package bstates

import (
	"fmt"
	"github.com/jaracil/ei"
	"github.com/nayarsystems/buffer/buffer"
	"github.com/nayarsystems/buffer/shuffling"
	"reflect"
)

// StateQueue is a queue of states stored in a buffer one after another.
//
// All elements in the same queue must use the [StateSchema] of the queue.
type StateQueue struct {
	StateSchema *StateSchema   // Schemas used to encode each of the states in the queue.
	buffer      *buffer.Buffer // Buffer where the queue is stored.
}

// Creates a [StateQueue] of [States] encoded with the [StateSchema] provided.
func CreateStateQueue(schema *StateSchema) *StateQueue {
	return &StateQueue{
		StateSchema: schema,
		buffer:      &buffer.Buffer{},
	}
}

// GetBitSize return the number of bits used by the internal buffer.
func (s *StateQueue) GetBitSize() int {
	return s.buffer.GetByteSize() * 8
}

// GetByteSize return the number of bytes used by the internal buffer.
func (s *StateQueue) GetByteSize() int {
	return s.buffer.GetByteSize()
}

// Clear deletes all elements in the queue by creating a new internal buffer.
func (s *StateQueue) Clear() {
	s.buffer.Init(0)
}

// PushAll pushes all the [State] objects provided at the end of the queue.
func (s *StateQueue) PushAll(states []*State) error {
	for _, e := range states {
		err := s.Push(e)
		if err != nil {
			return err
		}
	}
	return nil
}

// Push the [State] object provided into the end of the queue.
func (s *StateQueue) Push(state *State) error {
	if state.GetSchema().GetSHA256() != s.StateSchema.GetSHA256() {
		return fmt.Errorf("schema used in new state does not match the schema used by this state queue")
	}
	stateBuff, err := state.Encode()
	if err != nil {
		return err
	}
	s.buffer.Write(stateBuff, state.GetByteSize()*8)
	return nil
}

// Pop the first [State] in the queue removing it from the queue.
func (s *StateQueue) Pop() (*State, error) {
	stateSize := s.StateSchema.GetByteSize() * 8
	outBuffer, err := s.buffer.Read(stateSize)
	if err != nil {
		return nil, err
	}
	outState, err := s.StateSchema.CreateState()
	if err != nil {
		return nil, err
	}
	outState.Decode(outBuffer.GetRawBuffer())
	return outState, nil
}

// ToMsi converts the [StateQueue] into a map[string]interface{} representation.
func (s *StateQueue) ToMsi() (msg map[string]interface{}, err error) {
	msg = map[string]interface{}{}
	msg["schema"] = s.StateSchema.GetHashString()
	msg["payload"], err = s.Encode()
	return msg, err
}

// FromMsi populates the [StateQueue] from a map representation.
func (s *StateQueue) FromMsi(msg map[string]interface{}) error {
	schemaIdStr, err := ei.N(msg).M("schema").String()
	if err != nil {
		return err
	}
	sameSchema := reflect.DeepEqual(schemaIdStr, s.StateSchema.GetHashString())
	if !sameSchema {
		return fmt.Errorf("incompatible state queue message")
	}
	blob, err := ei.N(msg).M("payload").Bytes()
	if err != nil {
		return err
	}
	err = s.Decode(blob)
	return err
}

// Encode runs the encoderPipeline of the schema and outputs a binary blob with the queue compressed.
func (s *StateQueue) Encode() (dataOut []byte, err error) {
	inputBuf := s.buffer
	encPipe := s.StateSchema.GetEncoderPipeline()
	for _, mod := range encPipe {
		switch mod {
		case MOD_GZIP:
			inputBuf, err = GzipEnc(inputBuf)
			if err != nil {
				return
			}
		case MOD_ZSTD:
			inputBuf, err = ZstdEnc(inputBuf)
			if err != nil {
				return
			}
		case MOD_BITTRANS:
			stateBitSize := s.StateSchema.GetByteSize() * 8
			inputBuf, err = shuffling.TransposeBits(inputBuf, stateBitSize)
			if err != nil {
				return
			}
		}
	}
	dataOut = make([]byte, inputBuf.GetByteSize())
	copy(dataOut, inputBuf.GetRawBuffer())
	return
}

// Decode [Clear] the queue and then runs the decoderPipeline of the schema populating the queue.
func (s *StateQueue) Decode(data []byte) (err error) {
	s.Clear()
	inputBuf := &buffer.Buffer{}
	inputBuf.InitFromRawBuffer(data)

	if len(data) == 0 {
		// This want to avoid EOF error when decompressing empty data
		return
	}

	decPipe := s.StateSchema.GetDecoderPipeline()
	for _, mod := range decPipe {
		switch mod {
		case MOD_GZIP:
			inputBuf, err = GzipDec(inputBuf)
			if err != nil {
				return
			}
		case MOD_ZSTD:
			inputBuf, err = ZstdDec(inputBuf)
			if err != nil {
				return
			}
		case MOD_BITTRANS:
			if inputBuf.GetBitSize() == 0 {
				continue
			}
			numStates := inputBuf.GetByteSize() / s.StateSchema.GetByteSize()
			inputBuf, err = shuffling.TransposeBits(inputBuf, numStates)
			if err != nil {
				return
			}
		}
	}
	s.buffer.Write(inputBuf.GetRawBuffer(), inputBuf.GetBitSize())
	_, err = s.GetStates()
	return
}

// GetStates returns a slice with all the events on the queue.
func (s *StateQueue) GetStates() ([]*State, error) {
	queueBitSize := s.GetBitSize()
	stateByteSize := s.StateSchema.GetByteSize()
	stateBitSize := stateByteSize * 8
	states := make([]*State, 0)
	for bit := 0; bit < queueBitSize; bit += stateBitSize {
		state, err := s.StateSchema.CreateState()
		if err != nil {
			return nil, err
		}
		stateBuffer, err := s.buffer.GetBitsToRawBuffer(bit, stateBitSize)
		if err != nil {
			return nil, err
		}
		err = state.Decode(stateBuffer)
		if err != nil {
			return nil, err
		}
		states = append(states, state)
	}
	return states, nil
}

// GetNumStates returns the number of states in the queue.
func (s *StateQueue) GetNumStates() (num int) {
	queueByteSize := s.GetByteSize()
	stateByteSize := s.StateSchema.GetByteSize()
	numStates := queueByteSize / stateByteSize
	return numStates
}

// GetStateAt returns the state at an index.
func (s *StateQueue) GetStateAt(index int) (*State, error) {
	queueByteSize := s.GetByteSize()
	stateByteSize := s.StateSchema.GetByteSize()
	fullRawBuffer := s.buffer.GetRawBuffer()
	b := index * stateByteSize
	if b >= queueByteSize {
		return nil, fmt.Errorf("index out of range")
	}
	stateBuffer := fullRawBuffer[b : b+stateByteSize]
	tmpState, _ := s.StateSchema.CreateState()
	err := tmpState.Decode(stateBuffer)
	if err != nil {
		return nil, err
	}
	return tmpState, nil
}

// StateBufferIter iterates trough the queue executing the callback for each [State]. If the callback returns
// true the iterations ends early.
func (s *StateQueue) StateBufferIter(iterFunc func(stateBuffer []byte) (end bool)) {
	queueByteSize := s.GetByteSize()
	stateByteSize := s.StateSchema.GetByteSize()
	fullRawBuffer := s.buffer.GetRawBuffer()
	for b := 0; b < queueByteSize; b += stateByteSize {
		stateBuffer := fullRawBuffer[b : b+stateByteSize]
		end := iterFunc(stateBuffer)
		if end {
			return
		}
	}
}

// StateBufferIterFrom iterates the queue as [StateBufferIter] but starting from the state index provided.
func (s *StateQueue) StateBufferIterFrom(from int, iterFunc func(stateBuffer []byte) (end bool)) {
	queueByteSize := s.GetByteSize()
	stateByteSize := s.StateSchema.GetByteSize()
	fullRawBuffer := s.buffer.GetRawBuffer()
	for b := from * stateByteSize; b < queueByteSize; b += stateByteSize {
		stateBuffer := fullRawBuffer[b : b+stateByteSize]
		end := iterFunc(stateBuffer)
		if end {
			return
		}
	}
}
