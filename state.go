package bstates

import (
	"errors"
	"fmt"
	"math"
	"reflect"

	"github.com/jaracil/ei"
	"github.com/nayarsystems/buffer/frame"
)

// State represents a system state in a point of time.
// It holds data in the [frame.Frame] as defined by the provided [StateSchema].
type State struct {
	*frame.Frame              // Underlying binary data of the state
	schema       *StateSchema // Schema used for decoding the binary data of the state
	aliasMap     map[string]string
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

	// Build alias map for quick lookup to original field names
	// (look for both regular and decoded fields)
	aliasMap := make(map[string]string)
	for _, f := range schema.GetFields() {
		for _, alias := range f.Aliases {
			aliasMap[alias] = f.Name
		}
	}
	for _, df := range schema.GetDecodedFields() {
		for _, alias := range df.Aliases {
			aliasMap[alias] = df.Name
		}
	}

	state := &State{
		Frame:    f,
		schema:   schema,
		aliasMap: aliasMap,
	}
	return state, nil
}

// GetCopy returns a deep copy of the current [State].
// This includes a copy of the underlying [Frame] and retains the original [StateSchema].
func (e *State) GetCopy() *State {
	fcopy := e.Frame.GetCopy()
	ecopy := &State{
		Frame:    fcopy,
		schema:   e.schema,
		aliasMap: e.aliasMap,
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
func (f *State) Get(fieldName string) (value any, err error) {
	// Look for an alias first
	if originalName, ok := f.aliasMap[fieldName]; ok {
		fieldName = originalName
	}
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

func (f *State) Same(fieldName string, newValue any) (same bool, err error) {
	if originalName, ok := f.aliasMap[fieldName]; ok {
		fieldName = originalName
	}
	field, ok := f.schema.fieldsMap[fieldName]
	if !ok {
		// Check if it's a decoded field
		if _, ok := f.schema.decodedFields[fieldName]; ok {
			oldValue, err := f.Get(fieldName)
			if err != nil {
				return false, err
			}
			return reflect.DeepEqual(oldValue, newValue), nil
		}
	}
	// Its a regular field.
	switch field.Type {
	case T_FIXED:
		oldValue, err := f.Get(fieldName)
		if err != nil {
			return false, err
		}
		newValue := fromFixedPoint(toSignedFixedPoint(newValue, field.fixedPointCachedFactor), field.fixedPointCachedFactor)
		return reflect.DeepEqual(oldValue, newValue), nil
	case T_UFIXED:
		oldValue, err := f.Get(fieldName)
		if err != nil {
			return false, err
		}
		newValue := fromFixedPoint(toUnsignedFixedPoint(newValue, field.fixedPointCachedFactor), field.fixedPointCachedFactor)
		return reflect.DeepEqual(oldValue, newValue), nil
	}
	return f.Frame.Same(fieldName, newValue)
}

// Set updates the value of the specified field in the [State].
//
// It first checks if the field is a decoded field and, if so, uses the schema's encoding logic.
// Otherwise, it validates the value and updates the field using default encoding logic for the type of the field.
//
// For T_BUFFER fields with range errors (oversized data), the value is still written to allow
// truncation during encode/decode, but an error is returned to notify about potential data loss.
func (f *State) Set(fieldName string, newValue any) error {
	if originalName, ok := f.aliasMap[fieldName]; ok {
		fieldName = originalName
	}

	// Handle decoded fields using their specific encoder
	if df, ok := f.schema.decodedFields[fieldName]; ok {
		return df.Decoder.Encode(f, newValue)
	}

	// Find the field in the schema
	field, ok := f.schema.fieldsMap[fieldName]
	if !ok {
		return fmt.Errorf("field \"%s\" not found in schema", fieldName)
	}

	// Validate type and range before setting
	validationErr := field.Validate(newValue)

	if validationErr != nil {
		// Type errors are always fatal - cannot proceed
		if errors.Is(validationErr, ErrInvalidType) {
			return fmt.Errorf("field \"%s\": %v", fieldName, validationErr)
		}

		// Range errors for non-T_BUFFER types are fatal
		// For T_BUFFER, we allow the operation to continue (truncation will occur)
		if field.Type != T_BUFFER {
			return fmt.Errorf("field \"%s\": %v", fieldName, validationErr)
		}
	}

	// Convert fixed-point values to their internal representation
	switch field.Type {
	case T_FIXED:
		newValue = toSignedFixedPoint(newValue, field.fixedPointCachedFactor)
	case T_UFIXED:
		newValue = toUnsignedFixedPoint(newValue, field.fixedPointCachedFactor)
	}

	// Set the value (Frame.Set will handle truncation for T_BUFFER)
	setErr := f.Frame.Set(fieldName, newValue)
	if setErr != nil {
		return setErr
	}

	// Return validation error for T_BUFFER after successful set (with truncation)
	if validationErr != nil && field.Type == T_BUFFER {
		return fmt.Errorf("field \"%s\": %v", fieldName, validationErr)
	}

	return nil
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
// It includes both regular fields, decoded fields, and field aliases for backward compatibility.
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
	// Add aliases
	for alias, originalName := range e.aliasMap {
		if v, exists := data[originalName]; exists {
			data[alias] = v
		}
	}
	return data, nil
}
