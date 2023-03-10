package bstates

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"crypto/sha256"

	"github.com/jaracil/ei"
	hash "github.com/mitchellh/hashstructure/v2"
)

const (
	MOD_GZIP     = "z"
	MOD_BITTRANS = "t"
)

const (
	SCHEMA_VERSION_1_0 = "1.0"
	SCHEMA_VERSION_2_0 = "2.0"
)

type StateSchema struct {
	fields          []StateField
	decodedFields   map[string]DecodedStateField
	fieldsBitSize   int
	fieldsByteSize  int
	encoderPipeline []string
	decoderPipeline []string
	decoderIntMaps  map[string]map[int64]interface{}
}

type StateSchemaParams struct {
	Fields          []StateField
	DecodedFields   []DecodedStateField
	EncoderPipeline string
	DecoderIntMaps  map[string]map[int64]interface{}
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

	e.decodedFields = make(map[string]DecodedStateField)
	for _, decodedField := range params.DecodedFields {
		e.decodedFields[decodedField.Name] = decodedField
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
	decodedFieldsList := []DecodedStateField{}
	for _, f := range s.decodedFields {
		decodedFieldsList = append(decodedFieldsList, f)
	}
	data := map[string]interface{}{
		"version":         SCHEMA_VERSION_2_0,
		"encoderPipeline": strings.Join(s.encoderPipeline, ":"),
		"decoderIntMaps":  s.decoderIntMaps,
		"decodedFields":   decodedFieldsList,
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

	var version string
	version, err = ei.N(rawMap).M("version").String()
	if err != nil {
		version = SCHEMA_VERSION_1_0
	}

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
	s.decodedFields = map[string]DecodedStateField{}

	if version == SCHEMA_VERSION_2_0 {
		var rawFields []interface{}
		if rawFields, err = ei.N(rawMap).M("decodedFields").Slice(); err != nil {
			return err
		}
		for _, rawField := range rawFields {
			msi, ok := rawField.(map[string]interface{})
			if !ok {
				return fmt.Errorf("wrong type for decoded state field")
			}
			field := DecodedStateField{}
			err = field.FromMsi(msi)
			if err != nil {
				return err
			}
			s.decodedFields[field.Name] = field
		}
	} else if version == SCHEMA_VERSION_1_0 {
		rawMappedFields := ei.N(rawMap).M("mappedFields").MapStrZ()
		for name, data := range rawMappedFields {
			msi, ok := data.(map[string]interface{})
			if !ok {
				return fmt.Errorf("wrong type for state field")
			}
			field := DecodedStateField{}
			field.Name = name

			field.Decoder, err = NewDecoder(string(IntMapDecoderType), msi)
			if err != nil {
				return fmt.Errorf("field decoder error: %v", err)
			}
			s.decodedFields[name] = field
		}
		rawDecodedFields := ei.N(rawMap).M("decodedFields").MapStrZ()
		for name, data := range rawDecodedFields {
			msi, ok := data.(map[string]interface{})
			if !ok {
				return fmt.Errorf("wrong type for state field")
			}
			field := DecodedStateField{
				Name: name,
			}
			decoderParams := map[string]interface{}{}

			decoderParams["from"], err = ei.N(msi).M("from").String()
			if err != nil {
				return fmt.Errorf("no source field specified for decoded field: %v", err)
			}

			decoderStr, err := ei.N(msi).M("decoder").String()
			if err != nil {
				return fmt.Errorf("no decoder specified for decoded field: %v", err)
			}
			field.Decoder, err = NewDecoder(decoderStr, decoderParams)
			if err != nil {
				return fmt.Errorf("field decoder error: %v", err)
			}
			s.decodedFields[name] = field
		}
	}

	return nil
}

func (s *StateSchema) GetHashString() string {
	hash := s.GetSHA256()
	return base64.StdEncoding.EncodeToString(hash[:])
}

type SchemaHashInput struct {
	Fields          []StateField
	DecodedFields   map[string]DecodedStateField
	EncoderPipeline []string
	DecoderPipeline []string
	DecoderIntMaps  map[string]map[int64]interface{}
}

func (s *StateSchema) GetSHA256() [32]byte {
	hi := SchemaHashInput{
		Fields:          s.fields,
		DecodedFields:   s.decodedFields,
		EncoderPipeline: s.encoderPipeline,
		DecoderPipeline: s.decoderPipeline,
		DecoderIntMaps:  s.decoderIntMaps,
	}
	hashRaw, _ := hash.Hash(hi, hash.FormatV2, &hash.HashOptions{
		Hasher: NewSHA256Hasher(),
	})
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, hashRaw)
	sha256.New()
	return sha256.Sum256(b)
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

type DecodedStateField struct {
	Name    string
	Decoder Decoder
}

func (df *DecodedStateField) ToMsi() (map[string]interface{}, error) {
	m := map[string]interface{}{}
	m["name"] = df.Name
	m["decoder"] = df.Decoder.Name()
	m["params"] = df.Decoder.GetParams()
	return m, nil
}

func (df *DecodedStateField) FromMsi(m map[string]interface{}) error {
	var err error
	df.Name, err = ei.N(m).M("name").String()
	if err != nil {
		return fmt.Errorf("name field error: %v", err)
	}
	params, err := ei.N(m).M("params").MapStr()
	if err != nil {
		return fmt.Errorf("params field error: %v", err)
	}
	dtype, err := ei.N(m).M("decoder").String()
	if err != nil {
		return fmt.Errorf("decoder field error: %v", err)
	}
	df.Decoder, err = NewDecoder(dtype, params)
	if err != nil {
		return err
	}
	return nil
}

func (df *DecodedStateField) MarshalJSON() (res []byte, err error) {
	rawMap, err := df.ToMsi()
	if err != nil {
		return nil, fmt.Errorf("fix me: decoded field marshal error: %v", err)
	}
	res, err = json.Marshal(rawMap)
	return res, err
}

func (df *DecodedStateField) UnmarshalJSON(b []byte) error {
	var rawField map[string]interface{}
	err := json.Unmarshal(b, &rawField)
	if err != nil {
		return fmt.Errorf("decoded field unmarshal error: %v", err)
	}
	return df.FromMsi(rawField)
}
