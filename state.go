package bstates

import (
	"fmt"

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
	v, err = f.getDecodedField(fieldName)
	if err == nil {
		return v, nil
	}
	return nil, err
}

func (f *State) getDecodedField(fieldName string) (value interface{}, err error) {
	df, ok := f.schema.decodedFields[fieldName]
	if !ok {
		return nil, fmt.Errorf("field \"%s\" not found", fieldName)
	}
	return df.Decoder.Decode(f)
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
	for name := range e.schema.decodedFields {
		v, err := e.Get(name)
		if err != nil {
			return nil, err
		}
		data[name] = v
	}
	return data, nil
}
