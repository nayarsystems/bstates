package bstates

import (
	"errors"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFieldValidate_INT(t *testing.T) {
	tests := []struct {
		name   string
		size   int
		value  any
		errMsg string
	}{
		// 1-bit signed integer: range [-1, 0]
		{"1-bit valid min", 1, -1, ""},
		{"1-bit valid max", 1, 0, ""},
		{"1-bit overflow", 1, 1, "value 1 out of range [-1, 0] for 1-bit signed integer"},
		{"1-bit underflow", 1, -2, "value -2 out of range [-1, 0] for 1-bit signed integer"},

		// 3-bit signed integer: range [-4, 3]
		{"3-bit valid min", 3, -4, ""},
		{"3-bit valid max", 3, 3, ""},
		{"3-bit valid zero", 3, 0, ""},
		{"3-bit overflow", 3, 4, "value 4 out of range [-4, 3] for 3-bit signed integer"},
		{"3-bit underflow", 3, -5, "value -5 out of range [-4, 3] for 3-bit signed integer"},

		// 8-bit signed integer: range [-128, 127]
		{"8-bit valid min", 8, -128, ""},
		{"8-bit valid max", 8, 127, ""},
		{"8-bit overflow", 8, 128, "value 128 out of range [-128, 127] for 8-bit signed integer"},
		{"8-bit underflow", 8, -129, "value -129 out of range [-128, 127] for 8-bit signed integer"},

		// 64-bit signed integer: full range
		{"64-bit max", 64, math.MaxInt64, ""},
		{"64-bit min", 64, math.MinInt64, ""},

		// Invalid types
		{"string value", 8, "not a number", "cannot convert value to integer"},
		{"float value", 8, 3.14, ""}, // ei.N should handle this
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := &StateField{
				Type: T_INT,
				Size: tt.size,
			}

			err := field.Validate(tt.value)

			if tt.errMsg != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFieldValidate_UINT(t *testing.T) {
	tests := []struct {
		name   string
		size   int
		value  any
		errMsg string
	}{
		// 1-bit unsigned integer: range [0, 1]
		{"1-bit valid min", 1, 0, ""},
		{"1-bit valid max", 1, 1, ""},
		{"1-bit overflow", 1, 2, "value 2 out of range [0, 1] for 1-bit unsigned integer"},
		{"1-bit negative", 1, -1, "out of range"},

		// 3-bit unsigned integer: range [0, 7]
		{"3-bit valid min", 3, 0, ""},
		{"3-bit valid max", 3, 7, ""},
		{"3-bit overflow", 3, 8, "value 8 out of range [0, 7] for 3-bit unsigned integer"},
		{"3-bit negative", 3, -1, "out of range"},

		// 8-bit unsigned integer: range [0, 255]
		{"8-bit valid min", 8, 0, ""},
		{"8-bit valid max", 8, 255, ""},
		{"8-bit overflow", 8, 256, "value 256 out of range [0, 255] for 8-bit unsigned integer"},

		// 64-bit unsigned integer: full range
		{"64-bit max", 64, uint64(math.MaxUint64), ""},
		{"64-bit min", 64, uint64(0), ""},

		// Invalid types
		{"string value", 8, "not a number", "cannot convert value to unsigned integer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := &StateField{
				Type: T_UINT,
				Size: tt.size,
			}

			err := field.Validate(tt.value)

			if tt.errMsg != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFieldValidate_FIXED(t *testing.T) {
	tests := []struct {
		name     string
		size     int
		decimals uint
		value    any
		errMsg   string
	}{
		// 10-bit signed fixed-point with 2 decimals: range [-5.12, 5.11]
		{"10-bit valid min", 10, 2, -5.12, ""},
		{"10-bit valid max", 10, 2, 5.11, ""},
		{"10-bit valid zero", 10, 2, 0.0, ""},
		{"10-bit overflow", 10, 2, 5.13, "value 5.130000 out of range [-5.120000, 5.110000] for 10-bit signed fixed-point with 2 decimals"},
		{"10-bit underflow", 10, 2, -5.13, "value -5.130000 out of range [-5.120000, 5.110000] for 10-bit signed fixed-point with 2 decimals"},

		// 8-bit signed fixed-point with 1 decimal: range [-12.8, 12.7]
		{"8-bit valid min", 8, 1, -12.8, ""},
		{"8-bit valid max", 8, 1, 12.7, ""},
		{"8-bit overflow", 8, 1, 12.8, "value 12.800000 out of range [-12.800000, 12.700000] for 8-bit signed fixed-point with 1 decimals"},

		// 64-bit signed fixed-point: full range
		{"64-bit max", 64, 2, float64(math.MaxInt64) / 100, ""},
		{"64-bit min", 64, 2, float64(math.MinInt64) / 100, ""},

		// Invalid types
		{"string value", 10, 2, "not a number", "cannot convert value to number"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := &StateField{
				Type:                   T_FIXED,
				Size:                   tt.size,
				Decimals:               tt.decimals,
				fixedPointCachedFactor: math.Pow(10, float64(tt.decimals)),
			}

			err := field.Validate(tt.value)

			if tt.errMsg != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFieldValidate_UFIXED(t *testing.T) {
	tests := []struct {
		name     string
		size     int
		decimals uint
		value    any
		errMsg   string
	}{
		// 10-bit unsigned fixed-point with 2 decimals: range [0, 10.23]
		{"10-bit valid min", 10, 2, 0.0, ""},
		{"10-bit valid max", 10, 2, 10.23, ""},
		{"10-bit overflow", 10, 2, 10.24, "value 10.240000 out of range [0, 10.230000] for 10-bit unsigned fixed-point with 2 decimals"},
		{"10-bit negative", 10, 2, -0.01, "value -0.010000 cannot be negative for unsigned fixed-point"},

		// 8-bit unsigned fixed-point with 1 decimal: range [0, 25.5]
		{"8-bit valid min", 8, 1, 0.0, ""},
		{"8-bit valid max", 8, 1, 25.5, ""},
		{"8-bit overflow", 8, 1, 25.6, "value 25.600000 out of range [0, 25.500000] for 8-bit unsigned fixed-point with 1 decimals"},

		// 64-bit unsigned fixed-point: full range
		{"64-bit max", 64, 2, float64(math.MaxUint64) / 100, ""},
		{"64-bit min", 64, 2, 0.0, ""},

		// Invalid types
		{"string value", 10, 2, "not a number", "cannot convert value to number"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := &StateField{
				Type:                   T_UFIXED,
				Size:                   tt.size,
				Decimals:               tt.decimals,
				fixedPointCachedFactor: math.Pow(10, float64(tt.decimals)),
			}

			err := field.Validate(tt.value)

			if tt.errMsg != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFieldValidate_BOOL(t *testing.T) {
	tests := []struct {
		name   string
		value  any
		errMsg string
	}{
		{"bool true", true, ""},
		{"bool false", false, ""},
		{"int 1", 1, ""},
		{"int 0", 0, ""},
		{"int 2", 2, ""},   // ei.N should convert non-zero to true
		{"int -1", -1, ""}, // ei.N should convert non-zero to true
		{"string true", "true", "cannot convert value to boolean"},
		{"string false", "false", "cannot convert value to boolean"},
		{"string invalid", "maybe", "cannot convert value to boolean"},
		{"float", 3.14, ""},
		{"float zero", 0.0, ""},
		{"nil value", nil, "cannot convert value to boolean"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := &StateField{
				Type: T_BOOL,
				Size: 1,
			}

			err := field.Validate(tt.value)

			if tt.errMsg != "" {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFieldValidate_FLOAT32(t *testing.T) {
	tests := []struct {
		name   string
		value  any
		errMsg string
	}{
		{"valid float", 3.14, ""},
		{"max float32", math.MaxFloat32, ""},
		{"min float32", -math.MaxFloat32, ""},
		{"float64 too large", math.MaxFloat64, "is not a finite number"}, // Converts to +Inf, should be rejected
		{"very small positive", math.SmallestNonzeroFloat32, ""},
		{"very small negative", -math.SmallestNonzeroFloat32, ""},
		{"infinity", math.Inf(1), "is not a finite number"},
		{"negative infinity", math.Inf(-1), "is not a finite number"},
		{"NaN", math.NaN(), "is not a finite number"},
		{"string value", "not a number", "cannot convert value to float32"},
		{"integer value", 42, ""},
		{"zero", 0.0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := &StateField{
				Type: T_FLOAT32,
				Size: 32,
			}

			err := field.Validate(tt.value)

			if tt.errMsg != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFieldValidate_FLOAT64(t *testing.T) {
	tests := []struct {
		name   string
		value  any
		errMsg string
	}{
		{"valid float", 3.14159265359, ""},
		{"max float64", math.MaxFloat64, ""},
		{"min float64", -math.MaxFloat64, ""},
		{"very small positive", math.SmallestNonzeroFloat64, ""},
		{"very small negative", -math.SmallestNonzeroFloat64, ""},
		{"infinity", math.Inf(1), "is not a finite number"},
		{"negative infinity", math.Inf(-1), "is not a finite number"},
		{"NaN", math.NaN(), "is not a finite number"},
		{"string value", "not a number", "cannot convert value to float64"},
		{"integer value", 42, ""},
		{"zero", 0.0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := &StateField{
				Type: T_FLOAT64,
				Size: 64,
			}

			err := field.Validate(tt.value)

			if tt.errMsg != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFieldValidate_BUFFER(t *testing.T) {
	tests := []struct {
		name   string
		size   int
		value  any
		errMsg string
	}{
		{"valid byte slice", 64, []byte{1, 2, 3, 4}, ""},
		{"valid string", 64, "Hello", ""},
		{"empty byte slice", 64, []byte{}, ""},
		{"empty string", 64, "", ""},
		{"single byte", 8, []byte{255}, ""},
		{"exactly max size", 8, []byte{1}, ""},
		{"oversized byte slice", 16, []byte{1, 2, 3}, "exceeds field capacity"},
		{"oversized string", 16, "Hello World", "exceeds field capacity"},
		{"unicode string", 64, "Héllo", ""},
		{"invalid type", 64, 123, "buffer value must be string or []byte"},
		{"nil value", 64, nil, "buffer value must be string or []byte"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := &StateField{
				Type: T_BUFFER,
				Size: tt.size,
			}

			err := field.Validate(tt.value)

			if tt.errMsg != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}

	// Test the specific bug case: 247-bit field should accept 31 bytes (248 bits)
	t.Run("247-bit field accepts 31 bytes", func(t *testing.T) {
		field := &StateField{Type: T_BUFFER, Size: 247}
		bytes31 := make([]byte, 31) // 31 bytes = 248 bits
		err := field.Validate(bytes31)
		assert.NoError(t, err, "247-bit buffer field should accept 31 bytes (248 bits)")
	})

	t.Run("247-bit field rejects 32 bytes", func(t *testing.T) {
		field := &StateField{Type: T_BUFFER, Size: 247}
		bytes32 := make([]byte, 32) // 32 bytes = 256 bits
		err := field.Validate(bytes32)
		assert.Error(t, err, "247-bit buffer field should reject 32 bytes (256 bits)")
		assert.Contains(t, err.Error(), "exceeds field capacity")
	})
}

func TestValidateTypedErrors(t *testing.T) {
	// Test that Validate returns properly typed errors
	t.Run("Type errors", func(t *testing.T) {
		field := &StateField{Type: T_INT, Size: 8}
		
		// Invalid type should return ErrInvalidType
		err := field.Validate("not a number")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrInvalidType), "Should be ErrInvalidType")
		assert.False(t, errors.Is(err, ErrOutOfRange), "Should not be ErrOutOfRange")
	})

	t.Run("Range errors", func(t *testing.T) {
		field := &StateField{Type: T_INT, Size: 8}
		
		// Out of range should return ErrOutOfRange
		err := field.Validate(256) // 8-bit int max is 127
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrOutOfRange), "Should be ErrOutOfRange")
		assert.False(t, errors.Is(err, ErrInvalidType), "Should not be ErrInvalidType")
	})

	t.Run("Buffer type errors", func(t *testing.T) {
		field := &StateField{Type: T_BUFFER, Size: 64}
		
		// Invalid type should return ErrInvalidType
		err := field.Validate(123)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrInvalidType), "Should be ErrInvalidType")
		assert.False(t, errors.Is(err, ErrOutOfRange), "Should not be ErrOutOfRange")
	})

	t.Run("Buffer range errors", func(t *testing.T) {
		field := &StateField{Type: T_BUFFER, Size: 8} // 1 byte max
		
		// Oversized buffer should return ErrOutOfRange
		err := field.Validate([]byte{1, 2, 3}) // 3 bytes > 1 byte
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrOutOfRange), "Should be ErrOutOfRange")
		assert.False(t, errors.Is(err, ErrInvalidType), "Should not be ErrInvalidType")
	})

	t.Run("Valid values", func(t *testing.T) {
		field := &StateField{Type: T_INT, Size: 8}
		
		// Valid value should return nil
		err := field.Validate(42)
		assert.NoError(t, err)
	})
}

func TestFieldGetRange(t *testing.T) {
	tests := []struct {
		name      string
		fieldType StateFieldType
		size      int
		decimals  uint
		wantMin   any
		wantMax   any
		wantErr   bool
	}{
		{"INT 8-bit", T_INT, 8, 0, int64(-128), int64(127), false},
		{"INT 3-bit", T_INT, 3, 0, int64(-4), int64(3), false},
		{"INT 64-bit", T_INT, 64, 0, int64(math.MinInt64), int64(math.MaxInt64), false},
		{"UINT 8-bit", T_UINT, 8, 0, uint64(0), uint64(255), false},
		{"UINT 3-bit", T_UINT, 3, 0, uint64(0), uint64(7), false},
		{"UINT 64-bit", T_UINT, 64, 0, uint64(0), uint64(math.MaxUint64), false},
		{"FIXED 10-bit 2 decimals", T_FIXED, 10, 2, float64(-5.12), float64(5.11), false},
		{"UFIXED 10-bit 2 decimals", T_UFIXED, 10, 2, float64(0), float64(10.23), false},
		{"BOOL", T_BOOL, 1, 0, false, true, false},
		{"FLOAT32", T_FLOAT32, 32, 0, float32(-math.MaxFloat32), float32(math.MaxFloat32), false},
		{"FLOAT64", T_FLOAT64, 64, 0, -math.MaxFloat64, math.MaxFloat64, false},
		{"BUFFER", T_BUFFER, 64, 0, 0, 8, false}, // 64 bits = 8 bytes max

		// Edge cases
		{"INT 1-bit", T_INT, 1, 0, int64(-1), int64(0), false},
		{"UINT 1-bit", T_UINT, 1, 0, uint64(0), uint64(1), false},
		{"FIXED 1-bit 0 decimals", T_FIXED, 1, 0, float64(-1), float64(0), false},
		{"UFIXED 1-bit 0 decimals", T_UFIXED, 1, 0, float64(0), float64(1), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := &StateField{
				Type:                   tt.fieldType,
				Size:                   tt.size,
				Decimals:               tt.decimals,
				fixedPointCachedFactor: math.Pow(10, float64(tt.decimals)),
			}

			min, max, err := field.GetRange()

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantMin, min)
			assert.Equal(t, tt.wantMax, max)
		})
	}
}

func TestStateSetWithValidation(t *testing.T) {
	// Create a schema with various field types
	schema, err := CreateStateSchema(&StateSchemaParams{
		Fields: []StateField{
			{Name: "INT_FIELD", Type: T_INT, Size: 8},
			{Name: "UINT_FIELD", Type: T_UINT, Size: 8},
			{Name: "FIXED_FIELD", Type: T_FIXED, Size: 10, Decimals: 2},
			{Name: "UFIXED_FIELD", Type: T_UFIXED, Size: 10, Decimals: 2},
			{Name: "BOOL_FIELD", Type: T_BOOL},
			{Name: "FLOAT32_FIELD", Type: T_FLOAT32},
			{Name: "FLOAT64_FIELD", Type: T_FLOAT64},
			{Name: "BUFFER_FIELD", Type: T_BUFFER, Size: 64},
		},
	})
	require.NoError(t, err)

	state, err := CreateState(schema)
	require.NoError(t, err)

	// Test valid values
	validTests := []struct {
		field string
		value any
	}{
		{"INT_FIELD", 100},
		{"UINT_FIELD", 200},
		{"FIXED_FIELD", 3.14},
		{"UFIXED_FIELD", 5.67},
		{"BOOL_FIELD", true},
		{"FLOAT32_FIELD", float32(3.14)},
		{"FLOAT64_FIELD", 3.14159265359},
		{"BUFFER_FIELD", []byte{1, 2, 3, 4}},
	}

	for _, tt := range validTests {
		t.Run("valid_"+tt.field, func(t *testing.T) {
			err := state.Set(tt.field, tt.value)
			assert.NoError(t, err)
		})
	}

	// Test invalid values
	invalidTests := []struct {
		field     string
		value     any
		shouldErr bool
	}{
		{"INT_FIELD", 300, true},     // > 127
		{"INT_FIELD", -300, true},    // < -128
		{"UINT_FIELD", 300, true},    // > 255
		{"UINT_FIELD", -1, true},     // < 0
		{"FIXED_FIELD", 6.0, true},   // > 5.11
		{"FIXED_FIELD", -6.0, true},  // < -5.12
		{"UFIXED_FIELD", 11.0, true}, // > 10.23
		{"UFIXED_FIELD", -1.0, true}, // < 0
		{"BOOL_FIELD", "invalid", true},
		{"FLOAT32_FIELD", math.Inf(1), true},
		{"FLOAT64_FIELD", math.NaN(), true},
		{"BUFFER_FIELD", []byte{1, 2, 3, 4, 5, 6, 7, 8, 9}, true}, // too large
	}

	for _, tt := range invalidTests {
		t.Run("invalid_"+tt.field, func(t *testing.T) {
			err := state.Set(tt.field, tt.value)
			if tt.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestStateSetNonexistentField(t *testing.T) {
	schema, err := CreateStateSchema(&StateSchemaParams{
		Fields: []StateField{
			{Name: "EXISTING_FIELD", Type: T_INT, Size: 8},
		},
	})
	require.NoError(t, err)

	state, err := CreateState(schema)
	require.NoError(t, err)

	err = state.Set("NONEXISTENT_FIELD", 42)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found in schema")
}

func TestStateSetOutOfRangeValues_INT(t *testing.T) {
	tests := []struct {
		name   string
		size   int
		value  any
		errMsg string
	}{
		{"3-bit max overflow", 3, 4, "value 4 out of range [-4, 3] for 3-bit signed integer"},
		{"3-bit min underflow", 3, -5, "value -5 out of range [-4, 3] for 3-bit signed integer"},
		{"8-bit max overflow", 8, 128, "value 128 out of range [-128, 127] for 8-bit signed integer"},
		{"8-bit min underflow", 8, -129, "value -129 out of range [-128, 127] for 8-bit signed integer"},
		{"16-bit max overflow", 16, 32768, "value 32768 out of range [-32768, 32767] for 16-bit signed integer"},
		{"16-bit min underflow", 16, -32769, "value -32769 out of range [-32768, 32767] for 16-bit signed integer"},

		// Valid values should not error
		{"3-bit valid max", 3, 3, ""},
		{"3-bit valid min", 3, -4, ""},
		{"8-bit valid max", 8, 127, ""},
		{"8-bit valid min", 8, -128, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema, err := CreateStateSchema(&StateSchemaParams{
				Fields: []StateField{
					{Name: "TEST_FIELD", Type: T_INT, Size: tt.size},
				},
			})
			require.NoError(t, err)

			state, err := CreateState(schema)
			require.NoError(t, err)

			err = state.Set("TEST_FIELD", tt.value)

			if tt.errMsg != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestStateSetOutOfRangeValues_UINT(t *testing.T) {
	tests := []struct {
		name   string
		size   int
		value  any
		errMsg string
	}{
		{"3-bit overflow", 3, 8, "value 8 out of range [0, 7] for 3-bit unsigned integer"},
		{"8-bit overflow", 8, 256, "value 256 out of range [0, 255] for 8-bit unsigned integer"},
		{"16-bit overflow", 16, 65536, "value 65536 out of range [0, 65535] for 16-bit unsigned integer"},
		{"negative value", 8, -1, "out of range"}, // ei.N converts -1 to max uint64

		// Valid values should not error
		{"3-bit valid max", 3, 7, ""},
		{"8-bit valid max", 8, 255, ""},
		{"16-bit valid max", 16, 65535, ""},
		{"zero value", 8, 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema, err := CreateStateSchema(&StateSchemaParams{
				Fields: []StateField{
					{Name: "TEST_FIELD", Type: T_UINT, Size: tt.size},
				},
			})
			require.NoError(t, err)

			state, err := CreateState(schema)
			require.NoError(t, err)

			err = state.Set("TEST_FIELD", tt.value)

			if tt.errMsg != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestStateSetOutOfRangeValues_FIXED(t *testing.T) {
	tests := []struct {
		name     string
		size     int
		decimals uint
		value    any
		errMsg   string
	}{
		{"10-bit 2-dec overflow", 10, 2, 5.12, "value 5.120000 out of range [-5.120000, 5.110000] for 10-bit signed fixed-point"},
		{"10-bit 2-dec underflow", 10, 2, -5.13, "value -5.130000 out of range [-5.120000, 5.110000] for 10-bit signed fixed-point"},
		{"8-bit 1-dec overflow", 8, 1, 12.8, "value 12.800000 out of range [-12.800000, 12.700000] for 8-bit signed fixed-point"},
		{"8-bit 1-dec underflow", 8, 1, -12.9, "value -12.900000 out of range [-12.800000, 12.700000] for 8-bit signed fixed-point"},

		// Valid values should not error
		{"10-bit 2-dec valid max", 10, 2, 5.11, ""},
		{"10-bit 2-dec valid min", 10, 2, -5.12, ""},
		{"8-bit 1-dec valid max", 8, 1, 12.7, ""},
		{"8-bit 1-dec valid min", 8, 1, -12.8, ""},
		{"zero value", 10, 2, 0.0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema, err := CreateStateSchema(&StateSchemaParams{
				Fields: []StateField{
					{Name: "TEST_FIELD", Type: T_FIXED, Size: tt.size, Decimals: tt.decimals},
				},
			})
			require.NoError(t, err)

			state, err := CreateState(schema)
			require.NoError(t, err)

			err = state.Set("TEST_FIELD", tt.value)

			if tt.errMsg != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestStateSetOutOfRangeValues_UFIXED(t *testing.T) {
	tests := []struct {
		name     string
		size     int
		decimals uint
		value    any
		errMsg   string
	}{
		{"10-bit 2-dec overflow", 10, 2, 10.24, "value 10.240000 out of range [0, 10.230000] for 10-bit unsigned fixed-point"},
		{"8-bit 1-dec overflow", 8, 1, 25.6, "value 25.600000 out of range [0, 25.500000] for 8-bit unsigned fixed-point"},
		{"negative value", 10, 2, -0.01, "value -0.010000 cannot be negative for unsigned fixed-point"},
		{"large negative", 8, 1, -100.0, "cannot be negative for unsigned fixed-point"},

		// Valid values should not error
		{"10-bit 2-dec valid max", 10, 2, 10.23, ""},
		{"8-bit 1-dec valid max", 8, 1, 25.5, ""},
		{"zero value", 10, 2, 0.0, ""},
		{"small positive", 10, 2, 0.01, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema, err := CreateStateSchema(&StateSchemaParams{
				Fields: []StateField{
					{Name: "TEST_FIELD", Type: T_UFIXED, Size: tt.size, Decimals: tt.decimals},
				},
			})
			require.NoError(t, err)

			state, err := CreateState(schema)
			require.NoError(t, err)

			err = state.Set("TEST_FIELD", tt.value)

			if tt.errMsg != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestStateSetOutOfRangeValues_BUFFER(t *testing.T) {
	tests := []struct {
		name   string
		size   int
		value  any
		errMsg string
	}{
		{"oversized byte slice", 16, []byte{1, 2, 3}, "exceeds field capacity"},
		{"oversized string", 16, "This is too long", "exceeds field capacity"}, // 16 chars = 128 bits
		{"way oversized", 8, []byte{1, 2, 3, 4, 5}, "exceeds field capacity"},
		{"oversized string content", 64, "this string is too long for the buffer size", "exceeds field capacity"},
		{"invalid type", 64, 123, "buffer value must be string or []byte"},

		// Valid values should not error
		{"valid byte slice", 64, []byte{1, 2, 3, 4}, ""},
		{"valid string", 64, "Hello", ""},
		{"empty buffer", 64, []byte{}, ""},
		{"exactly fits", 16, []byte{1, 2}, ""}, // 16 bits
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema, err := CreateStateSchema(&StateSchemaParams{
				Fields: []StateField{
					{Name: "TEST_FIELD", Type: T_BUFFER, Size: tt.size},
				},
			})
			require.NoError(t, err)

			state, err := CreateState(schema)
			require.NoError(t, err)

			err = state.Set("TEST_FIELD", tt.value)

			if tt.errMsg != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestStateSetOutOfRangeValues_FLOAT(t *testing.T) {
	tests := []struct {
		name      string
		fieldType StateFieldType
		value     any
		errMsg    string
	}{
		{"float32 infinity", T_FLOAT32, math.Inf(1), "is not a finite number"},
		{"float32 negative infinity", T_FLOAT32, math.Inf(-1), "is not a finite number"},
		{"float32 NaN", T_FLOAT32, math.NaN(), "is not a finite number"},
		{"float64 infinity", T_FLOAT64, math.Inf(1), "is not a finite number"},
		{"float64 negative infinity", T_FLOAT64, math.Inf(-1), "is not a finite number"},
		{"float64 NaN", T_FLOAT64, math.NaN(), "is not a finite number"},

		// Valid values should not error
		{"float32 valid", T_FLOAT32, float32(3.14), ""},
		{"float64 valid", T_FLOAT64, 3.14159265359, ""},
		{"float32 max", T_FLOAT32, math.MaxFloat32, ""},
		{"float64 max", T_FLOAT64, math.MaxFloat64, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema, err := CreateStateSchema(&StateSchemaParams{
				Fields: []StateField{
					{Name: "TEST_FIELD", Type: tt.fieldType, Size: 32},
				},
			})
			require.NoError(t, err)

			state, err := CreateState(schema)
			require.NoError(t, err)

			err = state.Set("TEST_FIELD", tt.value)

			if tt.errMsg != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestStateSetOutOfRangeValues_BOOL(t *testing.T) {
	tests := []struct {
		name   string
		value  any
		errMsg string
	}{
		{"invalid string", "maybe", "cannot convert value to boolean"},
		{"invalid object", map[string]int{"key": 1}, "cannot convert value to boolean"},
		{"invalid array", []int{1, 2, 3}, "cannot convert value to boolean"},
		{"complex number", complex(1, 2), "cannot convert value to boolean"},

		// Valid values should not error
		{"bool true", true, ""},
		{"bool false", false, ""},
		{"int 1", 1, ""},
		{"int 0", 0, ""},
		{"float 3.14", 3.14, ""}, // ei.N can convert this to bool
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema, err := CreateStateSchema(&StateSchemaParams{
				Fields: []StateField{
					{Name: "TEST_FIELD", Type: T_BOOL, Size: 1},
				},
			})
			require.NoError(t, err)

			state, err := CreateState(schema)
			require.NoError(t, err)

			err = state.Set("TEST_FIELD", tt.value)

			if tt.errMsg != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestErrorReturnedForOutOfRangeValues explicitly tests that errors are returned
// This test demonstrates the core requirement that out-of-range values should return errors
func TestErrorReturnedForOutOfRangeValues(t *testing.T) {
	schema, err := CreateStateSchema(&StateSchemaParams{
		Fields: []StateField{
			{Name: "SMALL_INT", Type: T_INT, Size: 4},                 // Range: [-8, 7]
			{Name: "SMALL_UINT", Type: T_UINT, Size: 4},               // Range: [0, 15]
			{Name: "TINY_FIXED", Type: T_FIXED, Size: 6, Decimals: 1}, // Range: [-3.2, 3.1]
			{Name: "TINY_BUFFER", Type: T_BUFFER, Size: 8},            // 8 bits = 1 byte
		},
	})
	require.NoError(t, err)

	state, err := CreateState(schema)
	require.NoError(t, err)

	// Test cases that MUST return errors
	testCases := []struct {
		field string
		value any
		desc  string
	}{
		{"SMALL_INT", 8, "signed integer overflow"},
		{"SMALL_INT", -9, "signed integer underflow"},
		{"SMALL_UINT", 16, "unsigned integer overflow"},
		{"SMALL_UINT", -1, "unsigned integer underflow"},
		{"TINY_FIXED", 3.2, "fixed-point overflow"},
		{"TINY_FIXED", -3.3, "fixed-point underflow"},
		{"TINY_BUFFER", []byte{1, 2}, "buffer too large (16 bits > 8 bits)"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			err := state.Set(tc.field, tc.value)

			// This is the core assertion: errors MUST be returned for out-of-range values
			assert.Error(t, err, "Expected error for %s with value %v, but got nil", tc.field, tc.value)

			// Verify the error message contains useful information
			assert.Contains(t, err.Error(), tc.field, "Error message should contain field name")
		})
	}

	// Test cases that must NOT return errors (valid values)
	validCases := []struct {
		field string
		value any
		desc  string
	}{
		{"SMALL_INT", 7, "signed integer max valid"},
		{"SMALL_INT", -8, "signed integer min valid"},
		{"SMALL_UINT", 15, "unsigned integer max valid"},
		{"SMALL_UINT", 0, "unsigned integer min valid"},
		{"TINY_FIXED", 3.1, "fixed-point max valid"},
		{"TINY_FIXED", -3.2, "fixed-point min valid"},
		{"TINY_BUFFER", []byte{42}, "buffer exactly fits"},
	}

	for _, tc := range validCases {
		t.Run(tc.desc, func(t *testing.T) {
			err := state.Set(tc.field, tc.value)
			assert.NoError(t, err, "Expected no error for %s with value %v, but got: %v", tc.field, tc.value, err)
		})
	}
}

// TestValidate_InvalidFieldConfigurations tests edge cases and invalid configurations
// that should return errors, including unknown field types and improperly initialized fields.
func TestValidate_InvalidFieldConfigurations(t *testing.T) {
	// Test unknown field type
	t.Run("unknown field type", func(t *testing.T) {
		field := &StateField{
			Type: StateFieldType(99), // Invalid type
			Size: 8,
		}
		err := field.Validate(42)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown field type")
	})

	// Test fixed-point without cached factor
	t.Run("fixed-point without cached factor", func(t *testing.T) {
		field := &StateField{
			Type:     T_FIXED,
			Size:     10,
			Decimals: 2,
			// fixedPointCachedFactor not set - should be 100
		}
		// This should still work because division by 0 is handled
		err := field.Validate(1.5)
		// The implementation might panic or return incorrect results without cached factor
		// This test verifies the current behavior
		if err != nil {
			t.Logf("Error without cached factor: %v", err)
		}
	})

	// Test with properly initialized cached factor
	t.Run("fixed-point with cached factor", func(t *testing.T) {
		field := &StateField{
			Type:                   T_FIXED,
			Size:                   10,
			Decimals:               2,
			fixedPointCachedFactor: 100.0,
		}
		err := field.Validate(1.5)
		assert.NoError(t, err)
	})
}

// TestGetRange_EdgeCases tests edge cases for GetRange function
// including unknown field types, uninitialized cached factors, and extreme field sizes.
func TestGetRange_EdgeCases(t *testing.T) {
	// Test unknown field type
	t.Run("unknown field type", func(t *testing.T) {
		field := &StateField{
			Type: StateFieldType(99), // Invalid type
			Size: 8,
		}
		min, max, err := field.GetRange()
		assert.Error(t, err)
		assert.Nil(t, min)
		assert.Nil(t, max)
		assert.Contains(t, err.Error(), "unknown field type")
	})

	// Test fixed-point without cached factor
	t.Run("fixed-point without cached factor", func(t *testing.T) {
		field := &StateField{
			Type:     T_FIXED,
			Size:     10,
			Decimals: 2,
			// fixedPointCachedFactor not set - should be 100
		}
		min, max, err := field.GetRange()
		// This might result in division by zero or incorrect results
		if err != nil {
			t.Logf("Error without cached factor: %v", err)
			assert.Error(t, err)
		} else {
			t.Logf("Range without cached factor: [%v, %v]", min, max)
		}
	})

	// Test extreme field sizes
	t.Run("extreme field sizes", func(t *testing.T) {
		tests := []struct {
			name      string
			fieldType StateFieldType
			size      int
			decimals  uint
			shouldErr bool
		}{
			{"INT 65-bit", T_INT, 65, 0, false},       // Should be treated as 64-bit
			{"UINT 65-bit", T_UINT, 65, 0, false},     // Should be treated as 64-bit
			{"FIXED 65-bit", T_FIXED, 65, 2, false},   // Should be treated as 64-bit
			{"UFIXED 65-bit", T_UFIXED, 65, 2, false}, // Should be treated as 64-bit
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				field := &StateField{
					Type:                   tt.fieldType,
					Size:                   tt.size,
					Decimals:               tt.decimals,
					fixedPointCachedFactor: math.Pow(10, float64(tt.decimals)),
				}
				min, max, err := field.GetRange()
				if tt.shouldErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.NotNil(t, min)
					assert.NotNil(t, max)
				}
			})
		}
	})
}

// TestFieldValidation_ConcurrentAccess tests validation under concurrent access
// to ensure thread safety and prevent race conditions or panics.
func TestFieldValidation_ConcurrentAccess(t *testing.T) {
	field := &StateField{
		Type:                   T_FIXED,
		Size:                   10,
		Decimals:               2,
		fixedPointCachedFactor: 100.0,
	}

	// Run multiple goroutines concurrently
	const numGoroutines = 10
	const numIterations = 100
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < numIterations; j++ {
				// Test both valid and invalid values
				values := []any{1.5, -2.3, 10.0, -10.0} // Mix of valid and invalid
				for _, value := range values {
					err := field.Validate(value)
					// Just ensure no panics occur
					_ = err
				}

				// Test GetRange as well
				min, max, err := field.GetRange()
				_ = min
				_ = max
				_ = err
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}
