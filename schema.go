package bstates

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"crypto/sha256"

	"github.com/jaracil/ei"
)

const (
	MOD_GZIP     = "z"
	MOD_BITTRANS = "t"
)

type StateSchema struct {
	fields          []StateField
	mappedFields    map[string]MappedStateField
	decodedFields   map[string]DecodedStateFields
	fieldsBitSize   int
	fieldsByteSize  int
	encoderPipeline []string
	decoderPipeline []string
	decoderIntMaps  map[string]map[int64]interface{}
}

type StateSchemaParams struct {
	Fields          []StateField
	MappedFields    map[string]MappedStateField
	DecodedFields   map[string]DecodedStateFields
	EncoderPipeline string
	DecoderIntMaps  map[string]map[int64]interface{}
}

type MappedStateField struct {
	From  string `json:"from"`
	MapId string `json:"mapId"`
}

type FieldDecoderType string

const (
	BufferToString FieldDecoderType = "BufferToString"
)

type DecodedStateFields struct {
	From         string           `json:"from"`
	FieldDecoder FieldDecoderType `json:"decoder"`
}

func CreateStateSchema(params *StateSchemaParams) (e *StateSchema, err error) {
	e = &StateSchema{}
	if err = e.setPipelines(params.EncoderPipeline); err != nil {
		return nil, err
	}
	for _, field := range params.Fields {
		err := field.normalize()
		if err != nil {
			return nil, err
		}
		e.fieldsBitSize += field.Size
		e.fields = append(e.fields, field)
	}
	e.updateByteSize()
	e.mappedFields = make(map[string]MappedStateField)
	for name, mappedField := range params.MappedFields {
		e.mappedFields[name] = mappedField
	}

	e.decodedFields = make(map[string]DecodedStateFields)
	for name, decodedField := range params.DecodedFields {
		e.decodedFields[name] = decodedField
	}

	e.decoderIntMaps = map[string]map[int64]interface{}{}
	for mid, m := range params.DecoderIntMaps {
		nm := make(map[int64]interface{})
		e.decoderIntMaps[mid] = nm
		for i, v := range m {
			nm[i] = v
		}
	}
	return
}

func (s *StateSchema) GetFields() []*StateField {
	fieldsCopy := make([]*StateField, 0, len(s.fields))
	for _, field := range s.fields {
		fieldCopy := field
		fieldsCopy = append(fieldsCopy, &fieldCopy)
	}
	return fieldsCopy
}

func (s *StateSchema) GetBitSize() int {
	return s.fieldsBitSize
}

func (s *StateSchema) GetByteSize() int {
	return s.fieldsByteSize
}

func (s *StateSchema) CreateState() (*State, error) {
	return CreateState(s)
}

func (s *StateSchema) ToMsi() map[string]interface{} {
	data := map[string]interface{}{
		"version":         "1.0",
		"encoderPipeline": strings.Join(s.encoderPipeline, ":"),
		"decoderIntMaps":  s.decoderIntMaps,
		"mappedFields":    s.mappedFields,
		"decodedFields":   s.decodedFields,
		"fields":          s.fields,
	}
	return data
}

func (s *StateSchema) MarshalJSON() (res []byte, err error) {
	data := s.ToMsi()
	return json.Marshal(data)
}

func (s *StateSchema) UnmarshalJSON(b []byte) error {
	var rawMap map[string]interface{}
	var err error
	if err = json.Unmarshal(b, &rawMap); err != nil {
		return err
	}

	//// Uncomment code below to version checking
	// var version string
	// version, err = ei.N(rawMap).M("version").String()
	// if err != nil {
	// 	version = "1.0"
	// }

	if err = s.setPipelines(ei.N(rawMap).M("encoderPipeline").StringZ()); err != nil {
		return err
	}

	var rawFields []interface{}
	if rawFields, err = ei.N(rawMap).M("fields").Slice(); err != nil {
		return err
	}
	s.fields = []StateField{}
	for _, rawField := range rawFields {
		msi, ok := rawField.(map[string]interface{})
		if !ok {
			return fmt.Errorf("wrong type for state field")
		}
		field := StateField{}
		err = field.FromMsi(msi)
		if err != nil {
			return err
		}
		s.fieldsBitSize += field.Size
		if err != nil {
			return err
		}
		s.fields = append(s.fields, field)

	}
	s.updateByteSize()

	decoderIntMapsRaw := ei.N(rawMap).M("decoderIntMaps").MapStrZ()

	s.decoderIntMaps = map[string]map[int64]interface{}{}

	for mapId, mapDataRaw := range decoderIntMapsRaw {
		mapData, err := ei.N(mapDataRaw).MapStr()
		if err != nil {
			return fmt.Errorf("can't parse map \"%s\": %v", mapId, err)
		}
		newMap := map[int64]interface{}{}
		for fromStr, toValue := range mapData {
			fromInt, err := strconv.ParseInt(fromStr, 10, 64)
			if err != nil {
				return fmt.Errorf("can't parse \"%s\" as int key (map \"%s\"): %v", fromStr, mapId, err)
			}
			newMap[fromInt] = toValue
		}
		s.decoderIntMaps[mapId] = newMap
	}

	rawMappedFields := ei.N(rawMap).M("mappedFields").MapStrZ()
	s.mappedFields = map[string]MappedStateField{}
	for name, data := range rawMappedFields {
		msi, ok := data.(map[string]interface{})
		if !ok {
			return fmt.Errorf("wrong type for state field")
		}
		field := MappedStateField{}
		field.From, err = ei.N(msi).M("from").String()
		if err != nil {
			return fmt.Errorf("no source field specified for mapped field: %v", err)
		}
		field.MapId, err = ei.N(msi).M("mapId").String()
		if err != nil {
			return fmt.Errorf("no map id specified for mapped field: %v", err)
		}
		s.mappedFields[name] = field
	}

	rawDecodedFields := ei.N(rawMap).M("decodedFields").MapStrZ()
	s.decodedFields = map[string]DecodedStateFields{}
	for name, data := range rawDecodedFields {
		msi, ok := data.(map[string]interface{})
		if !ok {
			return fmt.Errorf("wrong type for state field")
		}
		field := DecodedStateFields{}
		field.From, err = ei.N(msi).M("from").String()
		if err != nil {
			return fmt.Errorf("no source field specified for decoded field: %v", err)
		}
		decoderStr, err := ei.N(msi).M("decoder").String()
		if err != nil {
			return fmt.Errorf("no decoder specified for decoded field: %v", err)
		}
		switch decoderStr {
		case string(BufferToString):
			field.FieldDecoder = BufferToString
		default:
			return fmt.Errorf("unknown decoder \"%s\"", decoderStr)
		}
		s.decodedFields[name] = field
	}
	return nil
}

func (s *StateSchema) GetHashString() string {
	hash := s.GetSHA256()
	return base64.StdEncoding.EncodeToString(hash[:])
}

func (s *StateSchema) GetSHA256() [32]byte {
	raw, _ := json.Marshal(s)
	return sha256.Sum256(raw)
}

func (s *StateSchema) GetEncoderPipeline() []string {
	return s.encoderPipeline
}

func (s *StateSchema) GetDecoderPipeline() []string {
	return s.decoderPipeline
}

func (s *StateSchema) updateByteSize() {
	s.fieldsByteSize = s.fieldsBitSize / 8
	if s.fieldsBitSize%8 != 0 {
		s.fieldsByteSize += 1
	}
}

func (e *StateSchema) setPipelines(pipelineRaw string) error {
	pipelineRegex, err := regexp.Compile(`^([^:]+)(:[^:]+)*$`)
	if err != nil {
		return err
	}
	indexes := pipelineRegex.FindStringSubmatchIndex(pipelineRaw)
	modifiers := []string{}
	for indexes != nil {
		modifier := string(pipelineRaw[indexes[2]:indexes[3]])
		modifiers = append(modifiers, modifier)
		fromIdx := indexes[3]
		if indexes[3] < len(pipelineRaw) {
			fromIdx++
		}
		pipelineRaw = pipelineRaw[fromIdx:]
		indexes = pipelineRegex.FindStringSubmatchIndex(pipelineRaw)
	}
	for _, mod := range modifiers {
		switch mod {
		case MOD_GZIP, MOD_BITTRANS:
		default:
			return fmt.Errorf("\"%s\" is not a modifier", mod)
		}
	}
	e.encoderPipeline = modifiers
	e.decoderPipeline = make([]string, len(e.encoderPipeline))
	for mi, m := range e.encoderPipeline {
		e.decoderPipeline[len(e.encoderPipeline)-1-mi] = m
	}
	return nil
}

type StateFieldType int

const (
	T_INT StateFieldType = iota
	T_UINT
	T_FLOAT32
	T_FLOAT64
	T_BOOL
	T_BUFFER
)

type StateField struct {
	Name            string
	Size            int
	DefaultValue    interface{}
	Type            StateFieldType
	LossyDebouncing time.Duration
	LossyThrottle   time.Duration
}

func (e *StateField) MarshalJSON() (res []byte, err error) {
	rawMap, err := e.ToMsi()
	if err != nil {
		return nil, err
	}
	res, err = json.Marshal(rawMap)
	return res, err
}

func (e *StateField) UnmarshalJSON(b []byte) error {
	var rawField map[string]interface{}
	err := json.Unmarshal(b, &rawField)
	if err != nil {
		return err
	}
	return e.FromMsi(rawField)
}

func (e *StateField) ToMsi() (msiData map[string]interface{}, err error) {
	rawMap := map[string]interface{}{}
	rawMap["name"] = e.Name
	rawMap["lossyDebouncingMs"] = e.LossyDebouncing.Milliseconds()
	rawMap["lossyThrottleMs"] = e.LossyThrottle.Milliseconds()
	var fieldTypeStr string
	switch e.Type {
	case T_INT:
		fieldTypeStr = "int"
	case T_UINT:
		fieldTypeStr = "uint"
	case T_BOOL:
		fieldTypeStr = "bool"
	case T_FLOAT32:
		fieldTypeStr = "float32"
	case T_FLOAT64:
		fieldTypeStr = "float64"
	case T_BUFFER:
		fieldTypeStr = "buffer"
	}
	rawMap["type"] = fieldTypeStr
	rawMap["size"] = e.Size
	if e.DefaultValue != nil {
		rawMap["defaultValue"] = e.DefaultValue
	}
	return rawMap, nil
}

func (e *StateField) FromMsi(rawField map[string]interface{}) (err error) {
	if e.Name = ei.N(rawField).M("name").StringZ(); e.Name == "" {
		return fmt.Errorf("field name not found")
	}
	var typeStr string
	if typeStr, err = ei.N(rawField).M("type").String(); err != nil {
		return err
	}
	var lossyDebouncingMs uint
	if lossyDebouncingMs, err = ei.N(rawField).M("lossyDebouncingMs").Uint(); err != nil {
		lossyDebouncingMs = 0
	}
	e.LossyDebouncing = time.Millisecond * time.Duration(lossyDebouncingMs)

	var lossyThrottleMs uint
	if lossyThrottleMs, err = ei.N(rawField).M("lossyThrottleMs").Uint(); err != nil {
		lossyThrottleMs = 0
	}
	e.LossyThrottle = time.Millisecond * time.Duration(lossyThrottleMs)

	e.Size = ei.N(rawField).M("size").IntZ()
	e.DefaultValue = ei.N(rawField).M("defaultValue").RawZ()
	switch {
	case typeStr == "int":
		e.Type = T_INT
	case typeStr == "uint":
		e.Type = T_UINT
	case typeStr == "bool":
		e.Type = T_BOOL
	case typeStr == "float32":
		e.Type = T_FLOAT32
	case typeStr == "float64":
		e.Type = T_FLOAT64
	case typeStr == "buffer":
		e.Type = T_BUFFER
	default:
		return fmt.Errorf("unkown field type '%s'", typeStr)
	}
	err = e.normalize()
	return
}

func (e *StateField) normalize() error {
	var defaultValue interface{}
	switch e.Type {
	case T_INT:
		if e.Size > 64 || e.Size <= 0 {
			return fmt.Errorf("invalid field size for int type (must be: 0 < size <= 64)")
		}
		defaultValue = int64(0)
	case T_UINT:
		if e.Size > 64 || e.Size <= 0 {
			return fmt.Errorf("invalid field size for int type (must be: 0 < size <= 64)")
		}
		defaultValue = uint64(0)
	case T_BOOL:
		e.Size = 1
		defaultValue = false
	case T_FLOAT32:
		e.Size = 32
		defaultValue = float32(0)
	case T_FLOAT64:
		e.Size = 64
		defaultValue = float64(0)
	case T_BUFFER:
		if e.Size <= 0 {
			return fmt.Errorf("invalid field size for buffer type (must be: 0 < size)")
		}
		byteSize := e.Size / 8
		if e.Size%8 != 0 {
			byteSize += 1
		}
		defaultValue = make([]byte, byteSize)
	}
	if e.DefaultValue == nil {
		e.DefaultValue = defaultValue
	} else {
		var err error
		switch e.Type {
		case T_INT:
			switch v := e.DefaultValue.(type) {
			case int8, int16, int32, int, int64:
				e.DefaultValue = v
			default:
				e.DefaultValue, err = ei.N(e.DefaultValue).Int64()
			}
		case T_UINT:
			switch v := e.DefaultValue.(type) {
			case uint8, uint16, uint32, uint, uint64:
				e.DefaultValue = v
			default:
				e.DefaultValue, err = ei.N(e.DefaultValue).Uint64()
			}
		case T_BOOL:
			e.DefaultValue, err = ei.N(e.DefaultValue).Bool()
		case T_FLOAT32:
			e.DefaultValue, err = ei.N(e.DefaultValue).Float32()
		case T_FLOAT64:
			e.DefaultValue, err = ei.N(e.DefaultValue).Float64()
		case T_BUFFER:
			switch v := e.DefaultValue.(type) {
			case string:
				e.DefaultValue, err = base64.StdEncoding.DecodeString(v)
			case []byte:
				e.DefaultValue = v
			default:
				err = fmt.Errorf("not an array of bytes")
			}
		}
		if err != nil {
			return fmt.Errorf("default value does not match field type: %v", err)
		}
	}
	return nil
}
