package bstates

import (
	"fmt"
	"reflect"

	"github.com/jaracil/ei"
	"github.com/nayarsystems/idefix/libraries/buffer/buffer"
	"github.com/nayarsystems/idefix/libraries/buffer/shuffling"
)

type StateQueue struct {
	StateSchema *StateSchema
	buffer      *buffer.Buffer
}

func CreateStateQueue(schema *StateSchema) *StateQueue {
	return &StateQueue{
		StateSchema: schema,
		buffer:      &buffer.Buffer{},
	}
}

func (s *StateQueue) GetBitSize() int {
	return s.buffer.GetByteSize() * 8
}

func (s *StateQueue) GetByteSize() int {
	return s.buffer.GetByteSize()
}

func (s *StateQueue) Clear() {
	s.buffer.Init(0)
}

func (s *StateQueue) PushAll(states []*State) error {
	for _, e := range states {
		err := s.Push(e)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *StateQueue) Push(state *State) error {
	if state.GetSchema().GetSHA256() != s.StateSchema.GetSHA256() {
		return fmt.Errorf("schema used in the new state does not match the schema used by this state stack")
	}
	stateBuff, err := state.Encode()
	if err != nil {
		return err
	}
	s.buffer.Write(stateBuff, state.GetByteSize()*8)
	return nil
}

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

func (s *StateQueue) ToMsi() (msg map[string]interface{}, err error) {
	msg = map[string]interface{}{}
	msg["schema"] = s.StateSchema.GetHashString()
	msg["payload"], err = s.Encode()
	return msg, err
}

func (s *StateQueue) FromMsi(msg map[string]interface{}) error {
	schemaIdStr, err := ei.N(msg).M("schema").String()
	if err != nil {
		return err
	}
	sameSchema := reflect.DeepEqual(schemaIdStr, s.StateSchema.GetHashString())
	if !sameSchema {
		return fmt.Errorf("incompatible state stack message")
	}
	blob, err := ei.N(msg).M("payload").Bytes()
	if err != nil {
		return err
	}
	err = s.Decode(blob)
	return err
}

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

func (s *StateQueue) Decode(data []byte) (err error) {
	s.Clear()
	inputBuf := &buffer.Buffer{}
	inputBuf.InitFromRawBuffer(data)

	decPipe := s.StateSchema.GetDecoderPipeline()
	for _, mod := range decPipe {
		switch mod {
		case MOD_GZIP:
			inputBuf, err = GzipDec(inputBuf)
			if err != nil {
				return
			}
		case MOD_BITTRANS:
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
