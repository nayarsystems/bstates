package bstates

import (
	"fmt"
	"time"

	"github.com/jaracil/ei"
)

// FieldDecoderType defines the type for different field decoder names.
type FieldDecoderType string

// Implemented decoders are: [BufferToStringDecoder], [NumberToUnixTsMsDecoder] and [IntMapDecoder].
const (
	BufferToStringDecoderType   FieldDecoderType = "BufferToString"
	NumberToUnixTsMsDecoderType FieldDecoderType = "NumberToUnixTsMs"
	IntMapDecoderType           FieldDecoderType = "IntMap"
)

// Decoder is an interface that defines how to decode or transform state information. They
// provide a "virtual" field.
//
// While decoders usually use the "from" parameter to specify the name of a [StateField] to decode, it's not mandatory.
//
// Decoders can:
//   - Use multiple input fields (similar to having multiple "from" parameters), allowing for operations with
//     several operands.
//   - Access the entire state, enabling more complex decoding logic that isn't limited to specific fields.
//   - Define constant fields that do not depend on any state field, although this might have limited practical use.
type Decoder interface {
	Name() FieldDecoderType               // Decoder type
	Decode(s *State) (interface{}, error) // function called
	GetParams() map[string]interface{}    // returns a MSI
}

// NewDecoder creates a new [Decoder] instance based on the provided
// decoder type and parameters.
//
// dtype: Should be one of [FieldDecoderType].
func NewDecoder(dtype string, params map[string]interface{}) (d Decoder, err error) {
	switch FieldDecoderType(dtype) {
	case BufferToStringDecoderType:
		d, err = NewBufferToStringDecoder(params)
	case IntMapDecoderType:
		d, err = NewIntMapDecoder(params)
	case NumberToUnixTsMsDecoderType:
		d, err = NewNumberToUnixTsMsDecoder(params)
	default:
		err = fmt.Errorf("unknown decoder \"%s\"", dtype)
	}
	return
}

// BufferToString implements a [Decoder] which returns a string from a buffer.
// This means that the original buffer will be returned as a [string] object stopping at the first null character.
type BufferToStringDecoder struct {
	From string // "from" parameter: name of the encoded field as defined in StateSchema.Fields
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

// IntMapDecoder implements a [Decoder] which decodes an integer value into a string based on a mapping defined
// in the State object.
type IntMapDecoder struct {
	From  string // "from" parameter: name of the encoded field as defined in StateSchema.Fields
	MapId string // "mapId" parameter: name of the map as defined in the StateSchema.DecoderIntMaps
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

// NumberToUnixTsMsDecoder implements a [Decoder] which decodes a numeric value using the following formula:
//
// decodedValue = UnixMillis(year) + valueToDecode*factor
type NumberToUnixTsMsDecoder struct {
	From   string  // "from" parameter: name of the encoded field as defined in StateSchema.Fields
	Year   uint    // "year"
	Factor float64 // "factor"
}

func (d *NumberToUnixTsMsDecoder) GetParams() map[string]interface{} {
	m := map[string]interface{}{}
	m["from"] = d.From
	m["year"] = d.Year
	m["factor"] = d.Factor
	return m
}

func NewNumberToUnixTsMsDecoder(params map[string]interface{}) (d *NumberToUnixTsMsDecoder, err error) {
	d = &NumberToUnixTsMsDecoder{}
	d.From, err = ei.N(params).M("from").String()
	if err != nil {
		return nil, fmt.Errorf("\"from\" field error: %v", err)
	}
	d.Year, err = ei.N(params).M("year").Uint()
	if err != nil {
		return nil, fmt.Errorf("\"year\" field error: %v", err)
	}
	d.Factor, err = ei.N(params).M("factor").Float64()
	if err != nil {
		return nil, fmt.Errorf("\"factor\" field error: %v", err)
	}
	if d.Factor <= 0 {
		return nil, fmt.Errorf("\"factor\" must be > 0")
	}
	if d.Year < 1970 {
		return nil, fmt.Errorf("\"year\" must be >= 1970")
	}
	return
}

func (d *NumberToUnixTsMsDecoder) Name() FieldDecoderType {
	return NumberToUnixTsMsDecoderType
}

func (d *NumberToUnixTsMsDecoder) Decode(s *State) (interface{}, error) {
	fromValueI, err := s.Get(d.From)
	if err != nil {
		return nil, err
	}
	fromValue, err := ei.N(fromValueI).Float64()
	if err != nil {
		return nil, err
	}
	offsetDate := time.Date(int(d.Year), time.January, 1, 0, 0, 0, 0, time.UTC)
	offsetDateUnixMs := offsetDate.UnixMilli()
	// convert to millis using given factor
	unixTsMs := uint64(offsetDateUnixMs + int64(fromValue*d.Factor))
	return unixTsMs, nil
}
