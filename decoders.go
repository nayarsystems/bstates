package bstates

import (
	"fmt"

	"github.com/jaracil/ei"
)

type FieldDecoderType string

const (
	BufferToStringDecoderType FieldDecoderType = "BufferToString"
	IntMapDecoderType         FieldDecoderType = "IntMap"
)

type Decoder interface {
	Name() FieldDecoderType
	Decode(s *State) (interface{}, error)
	GetParams() map[string]interface{}
}

func NewDecoder(dtype string, params map[string]interface{}) (d Decoder, err error) {
	switch FieldDecoderType(dtype) {
	case BufferToStringDecoderType:
		d, err = NewBufferToStringDecoder(params)
	case IntMapDecoderType:
		d, err = NewIntMapDecoder(params)
	default:
		err = fmt.Errorf("unknown decoder \"%s\"", dtype)
	}
	return
}

type BufferToStringDecoder struct {
	From string
}

func (d *BufferToStringDecoder) GetParams() map[string]interface{} {
	m := map[string]interface{}{}
	m["from"] = d.From
	return m
}

func NewBufferToStringDecoder(params map[string]interface{}) (d *BufferToStringDecoder, err error) {
	d = &BufferToStringDecoder{}
	d.From, err = ei.N(params).M("from").String()
	if err != nil {
		return nil, err
	}
	return
}

func (d *BufferToStringDecoder) Name() FieldDecoderType {
	return BufferToStringDecoderType
}

func (d *BufferToStringDecoder) Decode(s *State) (interface{}, error) {
	fromValueI, err := s.Get(d.From)
	if err != nil {
		return nil, err
	}
	fromValue, err := ei.N(fromValueI).Bytes()
	if err != nil {
		return nil, err
	}
	i := 0
	for ; i < len(fromValue); i++ {
		if fromValue[i] == 0 {
			break
		}
	}
	return string(fromValue[:i]), nil
}

type IntMapDecoder struct {
	From  string
	MapId string
}

func (d *IntMapDecoder) GetParams() map[string]interface{} {
	m := map[string]interface{}{}
	m["from"] = d.From
	m["mapId"] = d.MapId
	return m
}

func NewIntMapDecoder(params map[string]interface{}) (d *IntMapDecoder, err error) {
	d = &IntMapDecoder{}
	d.From, err = ei.N(params).M("from").String()
	if err != nil {
		return nil, err
	}
	d.MapId, err = ei.N(params).M("mapId").String()
	if err != nil {
		return nil, err
	}
	return
}

func (d *IntMapDecoder) Name() FieldDecoderType {
	return IntMapDecoderType
}

func (d *IntMapDecoder) Decode(s *State) (interface{}, error) {
	fromValueI, err := s.Get(d.From)
	if err != nil {
		return nil, err
	}
	fromValue, err := ei.N(fromValueI).Int64()
	if err != nil {
		return nil, err
	}
	intMap, ok := s.schema.decoderIntMaps[d.MapId]
	if !ok {
		return nil, fmt.Errorf("map \"%s\" not found", d.MapId)
	}
	toValue, ok := intMap[fromValue]
	if !ok {
		//return nil, fmt.Errorf("value \"%d\" not in map", fromValue)
		return "UNKNOWN", nil
	}
	return toValue, nil
}
