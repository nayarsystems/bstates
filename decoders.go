package bstates

import (
	"errors"
	"fmt"
	"time"

	"github.com/jaracil/ei"
)

// FieldDecoderType defines the type for different field decoder names.
type FieldDecoderType string

// Implemented decoders are: [BufferToStringDecoder], [NumberToUnixTsMsDecoder], [IntMapDecoder] and [FlagsDecoder].
const (
	BufferToStringDecoderType   FieldDecoderType = "BufferToString"
	NumberToUnixTsMsDecoderType FieldDecoderType = "NumberToUnixTsMs"
	IntMapDecoderType           FieldDecoderType = "IntMap"
	FlagsDecoderType            FieldDecoderType = "Flags"
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
	Name() FieldDecoderType       // Decoder type
	Decode(s *State) (any, error) // function called to get the decoded value from the state
	Encode(s *State, v any) error // function called to encode the value into the state
	GetParams() map[string]any    // returns a MSI
}

// NewDecoder creates a new [Decoder] instance based on the provided
// decoder type and parameters.
//
// dtype: Should be one of [FieldDecoderType].
func NewDecoder(dtype string, params map[string]any) (d Decoder, err error) {
	switch FieldDecoderType(dtype) {
	case BufferToStringDecoderType:
		d, err = NewBufferToStringDecoder(params)
	case IntMapDecoderType:
		d, err = NewIntMapDecoder(params)
	case NumberToUnixTsMsDecoderType:
		d, err = NewNumberToUnixTsMsDecoder(params)
	case FlagsDecoderType:
		d, err = NewFlagsDecoder(params)
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

func (d *BufferToStringDecoder) GetParams() map[string]any {
	m := map[string]any{}
	m["from"] = d.From
	return m
}

func NewBufferToStringDecoder(params map[string]any) (d *BufferToStringDecoder, err error) {
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

func (d *BufferToStringDecoder) Decode(s *State) (any, error) {
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

func (d *BufferToStringDecoder) Encode(s *State, v any) error {
	return s.Set(d.From, []byte(v.(string)))
}

// IntMapDecoder implements a [Decoder] which decodes an integer value into a string based on a mapping defined
// in the State object.
type IntMapDecoder struct {
	From  string // "from" parameter: name of the encoded field as defined in StateSchema.Fields
	MapId string // "mapId" parameter: name of the map as defined in the StateSchema.DecoderIntMaps
}

func (d *IntMapDecoder) GetParams() map[string]any {
	m := map[string]any{}
	m["from"] = d.From
	m["mapId"] = d.MapId
	return m
}

func NewIntMapDecoder(params map[string]any) (d *IntMapDecoder, err error) {
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

func (d *IntMapDecoder) Decode(s *State) (any, error) {
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

func (d *IntMapDecoder) Encode(s *State, v any) error {
	// This is a read-only decoder
	return errors.New("IntMapDecoder is a read-only decoder (can't encode)")
}

// NumberToUnixTsMsDecoder implements a [Decoder] which decodes a numeric value using the following formula:
//
// decodedValue = UnixMillis(year) + valueToDecode*factor
type NumberToUnixTsMsDecoder struct {
	From   string  // "from" parameter: name of the encoded field as defined in StateSchema.Fields
	Year   uint    // "year"
	Factor float64 // "factor"
}

func (d *NumberToUnixTsMsDecoder) GetParams() map[string]any {
	m := map[string]any{}
	m["from"] = d.From
	m["year"] = d.Year
	m["factor"] = d.Factor
	return m
}

func NewNumberToUnixTsMsDecoder(params map[string]any) (d *NumberToUnixTsMsDecoder, err error) {
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

func (d *NumberToUnixTsMsDecoder) Decode(s *State) (any, error) {
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

func (d *NumberToUnixTsMsDecoder) Encode(s *State, v any) error {
	unixTsMs, err := ei.N(v).Uint64()
	if err != nil {
		return err
	}

	offsetDate := time.Date(int(d.Year), time.January, 1, 0, 0, 0, 0, time.UTC)
	offsetDateUnixMs := offsetDate.UnixMilli()

	return s.Set(d.From, float64(int64(unixTsMs)-offsetDateUnixMs)/d.Factor)
}

type FlagsDecoder struct {
	From  string           // "from" parameter: name of the encoded field as defined in StateSchema.Fields
	Flags map[string]uint8 // "flags" parameter: map of flag name to bit position
}

func NewFlagsDecoder(params map[string]any) (d *FlagsDecoder, err error) {
	d = &FlagsDecoder{}
	d.From, err = ei.N(params).M("from").String()
	if err != nil {
		return nil, err
	}
	flagsM, err := ei.N(params).M("flags").MapStr()
	if err != nil {
		return nil, err
	}
	d.Flags = map[string]uint8{}
	for k, v := range flagsM {
		vv, err := ei.N(v).Uint8()
		if err != nil {
			return nil, fmt.Errorf("flag \"%s\" bit position error: %v", k, err)
		}
		// Note: Bit position validation against field size is performed at runtime
		// during Encode/Decode operations when field schema information is available
		d.Flags[k] = vv
	}
	return
}

func (d *FlagsDecoder) Name() FieldDecoderType {
	return FlagsDecoderType
}

func (d *FlagsDecoder) GetParams() map[string]any {
	m := map[string]any{}
	m["from"] = d.From
	m["flags"] = d.Flags
	return m
}

func (d *FlagsDecoder) Decode(s *State) (any, error) {
	// Get the field information to validate bit positions
	field, exists := s.schema.fieldsMap[d.From]
	if !exists {
		return nil, fmt.Errorf("field \"%s\" not found in schema", d.From)
	}

	fromValueI, err := s.Get(d.From)
	if err != nil {
		return nil, err
	}
	fromValue, err := ei.N(fromValueI).Uint64()
	if err != nil {
		return nil, err
	}
	flags := []string{}
	for fname, fbit := range d.Flags {
		// Validate that the bit position fits within the field size
		if int(fbit) >= field.Size {
			return nil, fmt.Errorf("flag \"%s\" bit position %d exceeds field size %d bits", fname, fbit, field.Size)
		}
		if (fromValue & (1 << fbit)) != 0 {
			flags = append(flags, fname)
		}
	}
	return flags, nil
}

func (d *FlagsDecoder) Encode(s *State, v any) error {
	// Get the field information to validate bit positions
	field, exists := s.schema.fieldsMap[d.From]
	if !exists {
		return fmt.Errorf("field \"%s\" not found in schema", d.From)
	}

	// Only accept []string or []any (with string values)
	flags, ok := v.([]string)
	if !ok {
		if flagsAny, ok := v.([]any); ok {
			// Convert []any to []string
			flags = make([]string, len(flagsAny))
			for i, fa := range flagsAny {
				fs, ok := fa.(string)
				if !ok {
					return fmt.Errorf("expected []string, got []any with non-string value at index %d", i)
				}
				flags[i] = fs
			}
		} else {
			return fmt.Errorf("expected []string or []any (with string values), got %T", v)
		}
	}

	var fromValue uint64 = 0
	for _, fname := range flags {
		fbit, ok := d.Flags[fname]
		if !ok {
			return fmt.Errorf("unknown flag \"%s\"", fname)
		}
		// Validate that the bit position fits within the field size
		if int(fbit) >= field.Size {
			return fmt.Errorf("flag \"%s\" bit position %d exceeds field size %d bits", fname, fbit, field.Size)
		}
		fromValue |= (1 << fbit)
	}
	return s.Set(d.From, fromValue)
}
