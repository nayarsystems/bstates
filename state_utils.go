package bstates

import (
	"fmt"
	"reflect"
)

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
