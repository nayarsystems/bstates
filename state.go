package bstates

import (
	"fmt"
	"github.com/jaracil/ei"
	"github.com/nayarsystems/buffer/frame"
	"math"
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
		switch f.Type {
		case T_FIXED:
			f.DefaultValue = toSignedFixedPoint(f.DefaultValue, f.fixedPointCachedFactor)
		case T_UFIXED:
			f.DefaultValue = toUnsignedFixedPoint(f.DefaultValue, f.fixedPointCachedFactor)
		}
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
		field, exists := f.schema.fieldsMap[fieldName]
		if !exists {
			return nil, fmt.Errorf("field \"%s\" not found in schema", fieldName)
		}
		if field.Type == T_FIXED || field.Type == T_UFIXED {
			v = fromFixedPoint(v, field.fixedPointCachedFactor)
		}
		return v, nil
	}
	v, err = f.getDecodedField(fieldName)
	if err == nil {
		return v, nil
	}
	return nil, err
}

// Set updates the value of the specified field in the [State].
//
// It first checks if the field is a decoded field and, if so, uses the schema's encoding logic.
// Otherwise, it updates the field using default encoding logic for the type of the field.
func (f *State) Set(fieldName string, newValue interface{}) error {
	if df, ok := f.schema.decodedFields[fieldName]; ok {
		return df.Decoder.Encode(f, newValue)
	}
	field, ok := f.schema.fieldsMap[fieldName]
	if !ok {
		return fmt.Errorf("field \"%s\" not found in schema", fieldName)
	}
	switch field.Type {
	case T_FIXED:
		newValue = toSignedFixedPoint(newValue, field.fixedPointCachedFactor)
	case T_UFIXED:
		newValue = toUnsignedFixedPoint(newValue, field.fixedPointCachedFactor)
	}

	return f.Frame.Set(fieldName, newValue)
}

func toSignedFixedPoint(v any, factor float64) int64 {
	return int64(math.Round(ei.N(v).Float64Z() * factor))
}

func toUnsignedFixedPoint(v any, factor float64) uint64 {
	return uint64(math.Round(ei.N(v).Float64Z() * factor))
}

func fromFixedPoint(v any, factor float64) float64 {
	return ei.N(v).Float64Z() / factor
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
