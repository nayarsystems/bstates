package bstates

import (
	"fmt"

	"github.com/nayarsystems/buffer/frame"
)

// State represents a system state in a point of time.
// It holds data in the [frame.Frame] as defined by the provided [StateSchema].
type State struct {
	*frame.Frame              // Underlying binary data of the state
	schema       *StateSchema // Schema used for decoding the binary data of the state
}

// CreateState initializes a new empty [State] based on the provided [StateSchema].
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

// GetCopy returns a deep copy of the current [State].
// This includes a copy of the underlying [Frame] and retains the original [StateSchema].
func (e *State) GetCopy() *State {
	fcopy := e.Frame.GetCopy()
	ecopy := &State{
		Frame:  fcopy,
		schema: e.schema,
	}
	return ecopy
}

// GetSchema returns the [StateSchema] associated with the current [State].
func (e *State) GetSchema() *StateSchema {
	return e.schema
}

// Get retrieves the value of the specified field from the [State].
//
// It first tries to retrieve the raw value from the [frame.Frame] and, if unsuccessful,
// attempts to decode the field using the schema's decoding logic.
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

// ToMsi converts the [State] into a map[string]interface{} representation,
// where each field's name is a key, and its corresponding value is the field's value.
// It includes both regular fields and decoded fields.
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
