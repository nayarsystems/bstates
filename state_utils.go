package bstates

import (
	"fmt"
	"reflect"
)

// StatesToMsiStates converts a slice of State pointers into a slice of
// maps containing the corresponding MSI states.
//
// Each State is transformed into a map using the ToMsi() method.
func StatesToMsiStates(states []*State) (out []map[string]interface{}, err error) {
	for _, e := range states {
		var de map[string]interface{}
		if de, err = e.ToMsi(); err != nil {
			return
		}
		out = append(out, de)
	}
	return
}

// GetDeltaMsiState compares two State objects and returns a map
// containing the values that have changed between them.
func GetDeltaMsiState(from *State, to *State) (map[string]interface{}, error) {
	data := map[string]interface{}{}
	fields := to.GetFieldsDesc()
	fieldNames := []string{}
	for _, f := range fields {
		fieldNames = append(fieldNames, f.Name)
	}
	for name := range to.schema.decodedFields {
		fieldNames = append(fieldNames, name)
	}
	for _, name := range fieldNames {
		fromValue, err := from.Get(name)
		if err != nil {
			return nil, fmt.Errorf("field \"%s\" not found in source state", name)
		}
		toValue, err := to.Get(name)
		if err != nil {
			return nil, fmt.Errorf("field \"%s\" not found in final state", name)
		}
		if !reflect.DeepEqual(fromValue, toValue) {
			data[name] = toValue
		}
	}
	return data, nil
}

// GetDeltaMsiStates returns a slice of maps that represent the changes
// between each successive pair of State objects in the provided slice.
//
// The first State is converted to its MSI representation, and subsequent
// States are compared to the last seen State using GetDeltaMsiState().
func GetDeltaMsiStates(states []*State) ([]map[string]interface{}, error) {
	out := []map[string]interface{}{}
	if len(states) > 0 {
		evIni := states[0]
		evIniMsi, err := evIni.ToMsi()
		if err != nil {
			return nil, err
		}
		out = append(out, evIniMsi)
		lastEv := evIni
		for i := 1; i < len(states); i++ {
			ev := states[i]
			evMsi, err := GetDeltaMsiState(lastEv, ev)
			if err != nil {
				return nil, err
			}
			out = append(out, evMsi)
			lastEv = ev
		}
	}
	return out, nil
}
