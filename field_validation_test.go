package bstates

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFieldValidateRange_INT(t *testing.T) {
	tests := []struct {
		name        string
		size        int
		value       interface{}
		shouldError bool
		errorMsg    string
	}{
		// 3-bit signed integer: range [-4, 3]
		{"3-bit valid min", 3, -4, false, ""},
		{"3-bit valid max", 3, 3, false, ""},
		{"3-bit valid zero", 3, 0, false, ""},
		{"3-bit overflow", 3, 4, true, "value 4 out of range [-4, 3] for 3-bit signed integer"},
		{"3-bit underflow", 3, -5, true, "value -5 out of range [-4, 3] for 3-bit signed integer"},
		
		// 8-bit signed integer: range [-128, 127]
		{"8-bit valid min", 8, -128, false, ""},
		{"8-bit valid max", 8, 127, false, ""},
		{"8-bit overflow", 8, 128, true, "value 128 out of range [-128, 127] for 8-bit signed integer"},
		{"8-bit underflow", 8, -129, true, "value -129 out of range [-128, 127] for 8-bit signed integer"},
		
		// 64-bit signed integer: full range
		{"64-bit max", 64, math.MaxInt64, false, ""},
		{"64-bit min", 64, math.MinInt64, false, ""},
		
		// Invalid types
		{"string value", 8, "not a number", true, "value is not a valid integer"},
		{"float value", 8, 3.14, false, ""}, // ei.N should handle this
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := &StateField{
				Type: T_INT,
				Size: tt.size,
			}
			
			err := field.ValidateRange(tt.value)
			
			if tt.shouldError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFieldValidateRange_UINT(t *testing.T) {
	tests := []struct {
		name        string
		size        int
		value       interface{}
		shouldError bool
		errorMsg    string
	}{
		// 3-bit unsigned integer: range [0, 7]
		{"3-bit valid min", 3, 0, false, ""},
		{"3-bit valid max", 3, 7, false, ""},
		{"3-bit overflow", 3, 8, true, "value 8 out of range [0, 7] for 3-bit unsigned integer"},
		{"3-bit negative", 3, -1, true, "out of range"},
		
		// 8-bit unsigned integer: range [0, 255]
		{"8-bit valid min", 8, 0, false, ""},
		{"8-bit valid max", 8, 255, false, ""},
		{"8-bit overflow", 8, 256, true, "value 256 out of range [0, 255] for 8-bit unsigned integer"},
		
		// 64-bit unsigned integer: full range
		{"64-bit max", 64, uint64(math.MaxUint64), false, ""},
		{"64-bit min", 64, uint64(0), false, ""},
		
		// Invalid types
		{"string value", 8, "not a number", true, "value is not a valid unsigned integer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := &StateField{
				Type: T_UINT,
				Size: tt.size,
			}
			
			err := field.ValidateRange(tt.value)
			
			if tt.shouldError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFieldValidateRange_FIXED(t *testing.T) {
	tests := []struct {
		name        string
		size        int
		decimals    uint
		value       interface{}
		shouldError bool
		errorMsg    string
	}{
		// 10-bit signed fixed-point with 2 decimals: range [-5.12, 5.11]
		{"10-bit valid min", 10, 2, -5.12, false, ""},
		{"10-bit valid max", 10, 2, 5.11, false, ""},
		{"10-bit valid zero", 10, 2, 0.0, false, ""},
		{"10-bit overflow", 10, 2, 5.13, true, "value 5.130000 out of range [-5.120000, 5.110000] for 10-bit signed fixed-point with 2 decimals"},
		{"10-bit underflow", 10, 2, -5.13, true, "value -5.130000 out of range [-5.120000, 5.110000] for 10-bit signed fixed-point with 2 decimals"},
		
		// 8-bit signed fixed-point with 1 decimal: range [-12.8, 12.7]
		{"8-bit valid min", 8, 1, -12.8, false, ""},
		{"8-bit valid max", 8, 1, 12.7, false, ""},
		{"8-bit overflow", 8, 1, 12.8, true, "value 12.800000 out of range [-12.800000, 12.700000] for 8-bit signed fixed-point with 1 decimals"},
		
		// 64-bit signed fixed-point: full range
		{"64-bit max", 64, 2, float64(math.MaxInt64)/100, false, ""},
		{"64-bit min", 64, 2, float64(math.MinInt64)/100, false, ""},
		
		// Invalid types
		{"string value", 10, 2, "not a number", true, "value is not a valid number"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := &StateField{
				Type:                   T_FIXED,
				Size:                   tt.size,
				Decimals:               tt.decimals,
				fixedPointCachedFactor: math.Pow(10, float64(tt.decimals)),
			}
			
			err := field.ValidateRange(tt.value)
			
			if tt.shouldError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFieldValidateRange_UFIXED(t *testing.T) {
	tests := []struct {
		name        string
		size        int
		decimals    uint
		value       interface{}
		shouldError bool
		errorMsg    string
	}{
		// 10-bit unsigned fixed-point with 2 decimals: range [0, 10.23]
		{"10-bit valid min", 10, 2, 0.0, false, ""},
		{"10-bit valid max", 10, 2, 10.23, false, ""},
		{"10-bit overflow", 10, 2, 10.24, true, "value 10.240000 out of range [0, 10.230000] for 10-bit unsigned fixed-point with 2 decimals"},
		{"10-bit negative", 10, 2, -0.01, true, "value -0.010000 cannot be negative for unsigned fixed-point"},
		
		// 8-bit unsigned fixed-point with 1 decimal: range [0, 25.5]
		{"8-bit valid min", 8, 1, 0.0, false, ""},
		{"8-bit valid max", 8, 1, 25.5, false, ""},
		{"8-bit overflow", 8, 1, 25.6, true, "value 25.600000 out of range [0, 25.500000] for 8-bit unsigned fixed-point with 1 decimals"},
		
		// 64-bit unsigned fixed-point: full range
		{"64-bit max", 64, 2, float64(math.MaxUint64)/100, false, ""},
		{"64-bit min", 64, 2, 0.0, false, ""},
		
		// Invalid types
		{"string value", 10, 2, "not a number", true, "value is not a valid number"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := &StateField{
				Type:                   T_UFIXED,
				Size:                   tt.size,
				Decimals:               tt.decimals,
				fixedPointCachedFactor: math.Pow(10, float64(tt.decimals)),
			}
			
			err := field.ValidateRange(tt.value)
			
			if tt.shouldError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFieldValidateRange_BOOL(t *testing.T) {
	tests := []struct {
		name        string
		value       interface{}
		shouldError bool
	}{
		{"bool true", true, false},
		{"bool false", false, false},
		{"int 1", 1, false},
		{"int 0", 0, false},
		{"string true", "true", true},
		{"string false", "false", true},
		{"string invalid", "maybe", true},
		{"float", 3.14, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := &StateField{
				Type: T_BOOL,
				Size: 1,
			}
			
			err := field.ValidateRange(tt.value)
			
			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFieldValidateRange_FLOAT32(t *testing.T) {
	tests := []struct {
		name        string
		value       interface{}
		shouldError bool
		errorMsg    string
	}{
		{"valid float", 3.14, false, ""},
		{"max float32", math.MaxFloat32, false, ""},
		{"min float32", -math.MaxFloat32, false, ""},
		{"infinity", math.Inf(1), true, "is not a finite number"},
		{"negative infinity", math.Inf(-1), true, "is not a finite number"},
		{"NaN", math.NaN(), true, "is not a finite number"},
		{"string value", "not a number", true, "value is not a valid float32"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := &StateField{
				Type: T_FLOAT32,
				Size: 32,
			}
			
			err := field.ValidateRange(tt.value)
			
			if tt.shouldError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFieldValidateRange_FLOAT64(t *testing.T) {
	tests := []struct {
		name        string
		value       interface{}
		shouldError bool
		errorMsg    string
	}{
		{"valid float", 3.14159265359, false, ""},
		{"max float64", math.MaxFloat64, false, ""},
		{"min float64", -math.MaxFloat64, false, ""},
		{"infinity", math.Inf(1), true, "is not a finite number"},
		{"negative infinity", math.Inf(-1), true, "is not a finite number"},
		{"NaN", math.NaN(), true, "is not a finite number"},
		{"string value", "not a number", true, "value is not a valid float64"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := &StateField{
				Type: T_FLOAT64,
				Size: 64,
			}
			
			err := field.ValidateRange(tt.value)
			
			if tt.shouldError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFieldValidateRange_BUFFER(t *testing.T) {
	tests := []struct {
		name        string
		size        int
		value       interface{}
		shouldError bool
		errorMsg    string
	}{
		{"valid byte slice", 64, []byte{1, 2, 3, 4}, false, ""},
		{"valid base64 string", 64, "SGVsbG8=", false, ""}, // "Hello" in base64 (40 bits)
		{"empty byte slice", 64, []byte{}, false, ""},
		{"empty base64 string", 64, "", false, ""},
		{"oversized byte slice", 16, []byte{1, 2, 3}, true, "buffer size 24 bits exceeds field size 16 bits"},
		{"oversized base64", 16, "SGVsbG8gV29ybGQ=", true, "buffer size 88 bits exceeds field size 16 bits"},
		{"invalid base64", 64, "not base64!", true, "buffer value is not valid base64"},
		{"invalid type", 64, 123, true, "buffer value must be string (base64) or []byte"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := &StateField{
				Type: T_BUFFER,
				Size: tt.size,
			}
			
			err := field.ValidateRange(tt.value)
			
			if tt.shouldError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFieldGetRange(t *testing.T) {
	tests := []struct {
		name     string
		fieldType StateFieldType
		size     int
		decimals uint
		wantMin  interface{}
		wantMax  interface{}
		wantErr  bool
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
		{"BUFFER", T_BUFFER, 64, 0, 0, 64, false},
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
		value interface{}
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
		value     interface{}
		shouldErr bool
	}{
		{"INT_FIELD", 300, true},    // > 127
		{"INT_FIELD", -300, true},   // < -128
		{"UINT_FIELD", 300, true},   // > 255
		{"UINT_FIELD", -1, true},    // < 0
		{"FIXED_FIELD", 6.0, true},  // > 5.11
		{"FIXED_FIELD", -6.0, true}, // < -5.12
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
		name      string
		size      int
		value     interface{}
		expectErr bool
		errMsg    string
	}{
		{"3-bit max overflow", 3, 4, true, "value 4 out of range [-4, 3] for 3-bit signed integer"},
		{"3-bit min underflow", 3, -5, true, "value -5 out of range [-4, 3] for 3-bit signed integer"},
		{"8-bit max overflow", 8, 128, true, "value 128 out of range [-128, 127] for 8-bit signed integer"},
		{"8-bit min underflow", 8, -129, true, "value -129 out of range [-128, 127] for 8-bit signed integer"},
		{"16-bit max overflow", 16, 32768, true, "value 32768 out of range [-32768, 32767] for 16-bit signed integer"},
		{"16-bit min underflow", 16, -32769, true, "value -32769 out of range [-32768, 32767] for 16-bit signed integer"},
		
		// Valid values should not error
		{"3-bit valid max", 3, 3, false, ""},
		{"3-bit valid min", 3, -4, false, ""},
		{"8-bit valid max", 8, 127, false, ""},
		{"8-bit valid min", 8, -128, false, ""},
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
			
			if tt.expectErr {
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
		name      string
		size      int
		value     interface{}
		expectErr bool
		errMsg    string
	}{
		{"3-bit overflow", 3, 8, true, "value 8 out of range [0, 7] for 3-bit unsigned integer"},
		{"8-bit overflow", 8, 256, true, "value 256 out of range [0, 255] for 8-bit unsigned integer"},
		{"16-bit overflow", 16, 65536, true, "value 65536 out of range [0, 65535] for 16-bit unsigned integer"},
		{"negative value", 8, -1, true, "out of range"}, // ei.N converts -1 to max uint64
		
		// Valid values should not error
		{"3-bit valid max", 3, 7, false, ""},
		{"8-bit valid max", 8, 255, false, ""},
		{"16-bit valid max", 16, 65535, false, ""},
		{"zero value", 8, 0, false, ""},
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
			
			if tt.expectErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestStateSetOutOfRangeValues_FIXED(t *testing.T) {
	tests := []struct {
		name      string
		size      int
		decimals  uint
		value     interface{}
		expectErr bool
		errMsg    string
	}{
		{"10-bit 2-dec overflow", 10, 2, 5.12, true, "value 5.120000 out of range [-5.120000, 5.110000] for 10-bit signed fixed-point"},
		{"10-bit 2-dec underflow", 10, 2, -5.13, true, "value -5.130000 out of range [-5.120000, 5.110000] for 10-bit signed fixed-point"},
		{"8-bit 1-dec overflow", 8, 1, 12.8, true, "value 12.800000 out of range [-12.800000, 12.700000] for 8-bit signed fixed-point"},
		{"8-bit 1-dec underflow", 8, 1, -12.9, true, "value -12.900000 out of range [-12.800000, 12.700000] for 8-bit signed fixed-point"},
		
		// Valid values should not error
		{"10-bit 2-dec valid max", 10, 2, 5.11, false, ""},
		{"10-bit 2-dec valid min", 10, 2, -5.12, false, ""},
		{"8-bit 1-dec valid max", 8, 1, 12.7, false, ""},
		{"8-bit 1-dec valid min", 8, 1, -12.8, false, ""},
		{"zero value", 10, 2, 0.0, false, ""},
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
			
			if tt.expectErr {
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
		name      string
		size      int
		decimals  uint
		value     interface{}
		expectErr bool
		errMsg    string
	}{
		{"10-bit 2-dec overflow", 10, 2, 10.24, true, "value 10.240000 out of range [0, 10.230000] for 10-bit unsigned fixed-point"},
		{"8-bit 1-dec overflow", 8, 1, 25.6, true, "value 25.600000 out of range [0, 25.500000] for 8-bit unsigned fixed-point"},
		{"negative value", 10, 2, -0.01, true, "value -0.010000 cannot be negative for unsigned fixed-point"},
		{"large negative", 8, 1, -100.0, true, "cannot be negative for unsigned fixed-point"},
		
		// Valid values should not error
		{"10-bit 2-dec valid max", 10, 2, 10.23, false, ""},
		{"8-bit 1-dec valid max", 8, 1, 25.5, false, ""},
		{"zero value", 10, 2, 0.0, false, ""},
		{"small positive", 10, 2, 0.01, false, ""},
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
			
			if tt.expectErr {
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
		name      string
		size      int
		value     interface{}
		expectErr bool
		errMsg    string
	}{
		{"oversized byte slice", 16, []byte{1, 2, 3}, true, "buffer size 24 bits exceeds field size 16 bits"},
		{"oversized base64", 16, "SGVsbG8gV29ybGQ=", true, "buffer size 88 bits exceeds field size 16 bits"}, // "Hello World"
		{"way oversized", 8, []byte{1, 2, 3, 4, 5}, true, "buffer size 40 bits exceeds field size 8 bits"},
		{"invalid base64", 64, "not base64!", true, "buffer value is not valid base64"},
		{"invalid type", 64, 123, true, "buffer value must be string (base64) or []byte"},
		
		// Valid values should not error
		{"valid byte slice", 64, []byte{1, 2, 3, 4}, false, ""},
		{"valid base64", 64, "SGVsbG8=", false, ""}, // "Hello"
		{"empty buffer", 64, []byte{}, false, ""},
		{"exactly fits", 16, []byte{1, 2}, false, ""}, // 16 bits
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
			
			if tt.expectErr {
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
		value     interface{}
		expectErr bool
		errMsg    string
	}{
		{"float32 infinity", T_FLOAT32, math.Inf(1), true, "is not a finite number"},
		{"float32 negative infinity", T_FLOAT32, math.Inf(-1), true, "is not a finite number"},
		{"float32 NaN", T_FLOAT32, math.NaN(), true, "is not a finite number"},
		{"float64 infinity", T_FLOAT64, math.Inf(1), true, "is not a finite number"},
		{"float64 negative infinity", T_FLOAT64, math.Inf(-1), true, "is not a finite number"},
		{"float64 NaN", T_FLOAT64, math.NaN(), true, "is not a finite number"},
		
		// Valid values should not error
		{"float32 valid", T_FLOAT32, float32(3.14), false, ""},
		{"float64 valid", T_FLOAT64, 3.14159265359, false, ""},
		{"float32 max", T_FLOAT32, math.MaxFloat32, false, ""},
		{"float64 max", T_FLOAT64, math.MaxFloat64, false, ""},
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
			
			if tt.expectErr {
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
		name      string
		value     interface{}
		expectErr bool
		errMsg    string
	}{
		{"invalid string", "maybe", true, "value is not a valid boolean"},
		{"invalid object", map[string]int{"key": 1}, true, "value is not a valid boolean"},
		{"invalid array", []int{1, 2, 3}, true, "value is not a valid boolean"},
		
		// Valid values should not error
		{"bool true", true, false, ""},
		{"bool false", false, false, ""},
		{"int 1", 1, false, ""},
		{"int 0", 0, false, ""},
		{"float 3.14", 3.14, false, ""}, // ei.N can convert this to bool
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
			
			if tt.expectErr {
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
			{Name: "SMALL_INT", Type: T_INT, Size: 4},     // Range: [-8, 7]
			{Name: "SMALL_UINT", Type: T_UINT, Size: 4},   // Range: [0, 15]
			{Name: "TINY_FIXED", Type: T_FIXED, Size: 6, Decimals: 1}, // Range: [-3.2, 3.1]
			{Name: "TINY_BUFFER", Type: T_BUFFER, Size: 8}, // 8 bits = 1 byte
		},
	})
	require.NoError(t, err)

	state, err := CreateState(schema)
	require.NoError(t, err)

	// Test cases that MUST return errors
	testCases := []struct {
		field string
		value interface{}
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
		value interface{}
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