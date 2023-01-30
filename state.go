package bstates

import (
	"fmt"

	"github.com/jaracil/ei"
	"github.com/nayarsystems/buffer/frame"
)

type State struct {
	*frame.Frame
	schema *StateSchema
}

func CreateState(schema *StateSchema) (*State, error) {
	f := frame.CreateFrame()
	fields := []*frame.FieldDesc{}
	for _, f := range schema.GetFields() {
		fd := &frame.FieldDesc{
			Name:         f.Name,
			Size:         f.Size,
			DefaultValue: f.DefaultValue,
		}
		fields = append(fields, fd)
	}
	err := f.AddFields(fields)
	if err != nil {
		return nil, err
	}
	state := &State{
		Frame:  f,
		schema: schema,
	}
	return state, nil
}

func (e *State) GetCopy() *State {
	fcopy := e.Frame.GetCopy()
	ecopy := &State{
		Frame:  fcopy,
		schema: e.schema,
	}
	return ecopy
}

func (e *State) GetSchema() *StateSchema {
	return e.schema
}

func (f *State) Get(fieldName string) (value interface{}, err error) {
	v, err := f.Frame.Get(fieldName)
	if err == nil {
		return v, nil
	}
	v, err = f.getMappedField(fieldName)
	if err == nil {
		return v, nil
	}
	v, err = f.getDecodedField(fieldName)
	if err == nil {
		return v, nil
	}
	return nil, err
}

func (f *State) getMappedField(fieldName string) (value interface{}, err error) {
	mf, ok := f.schema.mappedFields[fieldName]
	if !ok {
		return nil, fmt.Errorf("field \"%s\" not found", fieldName)
	}
	fromValueI, err := f.Get(mf.From)
	if err != nil {
		return nil, err
	}
	fromValue, err := ei.N(fromValueI).Int64()
	if err != nil {
		return nil, err
	}
	intMap, ok := f.schema.decoderIntMaps[mf.MapId]
	if !ok {
		return nil, fmt.Errorf("map \"%s\" not found", mf.MapId)
	}
	toValue, ok := intMap[fromValue]
	if !ok {
		//return nil, fmt.Errorf("value \"%d\" not in map", fromValue)
		return "UNKNOWN", nil
	}
	return toValue, nil
}

func (f *State) getDecodedField(fieldName string) (value interface{}, err error) {
	df, ok := f.schema.decodedFields[fieldName]
	if !ok {
		return nil, fmt.Errorf("field \"%s\" not found", fieldName)
	}
	switch df.FieldDecoder {
	case BufferToString:
		fromValueI, err := f.Get(df.From)
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
	default:
		return nil, fmt.Errorf("unknown decoder for field \"%s\"", fieldName)
	}
}

func (e *State) ToMsi() (map[string]interface{}, error) {
	data := map[string]interface{}{}
	fields := e.GetFieldsDesc()
	for _, f := range fields {
		v, err := e.Get(f.Name)
		if err != nil {
			return nil, err
		}
		data[f.Name] = v
	}
	for name := range e.schema.mappedFields {
		v, err := e.Get(name)
		if err != nil {
			return nil, err
		}
		data[name] = v
	}

	for name := range e.schema.decodedFields {
		v, err := e.Get(name)
		if err != nil {
			return nil, err
		}
		data[name] = v
	}
	return data, nil
}
