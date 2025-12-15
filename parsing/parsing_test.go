package parsing_test

import (
	"math/big"
	"testing"

	"github.com/smartcontractkit/crec-sdk/parsing"
	"github.com/stretchr/testify/require"
)

func TestParsing_ScientificNotationToBigInt(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string // string representation of expected big.Int
		expectError bool
	}{
		// Regular decimal numbers
		{
			name:        "simple integer",
			input:       "123456789",
			expected:    "123456789",
			expectError: false,
		},
		{
			name:        "zero",
			input:       "0",
			expected:    "0",
			expectError: false,
		},
		{
			name:        "large integer",
			input:       "123456789012345678901234567890",
			expected:    "123456789012345678901234567890",
			expectError: false,
		},

		// Negative numbers
		{
			name:        "negative integer",
			input:       "-123456789",
			expected:    "-123456789",
			expectError: false,
		},
		{
			name:        "negative scientific notation",
			input:       "-1.5e+10",
			expected:    "-15000000000",
			expectError: false,
		},
		{
			name:        "negative mantissa with positive exponent",
			input:       "-5e18",
			expected:    "-5000000000000000000",
			expectError: false,
		},

		// Scientific notation - positive exponents
		{
			name:        "simple scientific notation",
			input:       "1e18",
			expected:    "1000000000000000000",
			expectError: false,
		},
		{
			name:        "scientific notation with decimal",
			input:       "1.2e+21",
			expected:    "1200000000000000000000",
			expectError: false,
		},
		{
			name:        "scientific notation uppercase E",
			input:       "5E20",
			expected:    "500000000000000000000",
			expectError: false,
		},
		{
			name:        "scientific notation with plus sign",
			input:       "3.14e+5",
			expected:    "314000",
			expectError: false,
		},
		{
			name:        "large scientific notation",
			input:       "1.23456789e+30",
			expected:    "1234567890000000000000000000000",
			expectError: false,
		},
		{
			name:        "exponent zero",
			input:       "123e0",
			expected:    "123",
			expectError: false,
		},
		{
			name:        "exponent zero with decimal",
			input:       "1.5e0",
			expected:    "1", // truncated
			expectError: false,
		},

		// Scientific notation - negative exponents (should truncate to integer)
		{
			name:        "scientific notation negative exponent small",
			input:       "1.5e-1",
			expected:    "0", // truncated to integer
			expectError: false,
		},
		{
			name:        "scientific notation negative exponent zero result",
			input:       "5e-10",
			expected:    "0", // truncated to integer
			expectError: false,
		},
		{
			name:        "scientific notation that becomes integer",
			input:       "1.23e2",
			expected:    "123",
			expectError: false,
		},
		{
			name:        "negative exponent with large mantissa preserves precision",
			input:       "123456789012345678901234567890e-10",
			expected:    "12345678901234567890",
			expectError: false,
		},
		{
			name:        "large mantissa with decimal and negative effective exponent",
			input:       "1.23456789012345678901234567890e+20",
			expected:    "123456789012345678901",
			expectError: false,
		},

		// Edge cases - very large numbers
		{
			name:        "very large scientific notation",
			input:       "1e100",
			expected:    "10000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
			expectError: false,
		},
		{
			name:        "scientific notation with many decimal places",
			input:       "1.23456789123456789e+20",
			expected:    "123456789123456789000",
			expectError: false,
		},
		{
			name:        "extremely precise mantissa",
			input:       "9.99999999999999999999999999e+26",
			expected:    "999999999999999999999999999",
			expectError: false,
		},

		// Decimal numbers without scientific notation
		{
			name:        "decimal number with zeros",
			input:       "600000000000000000000.000000",
			expected:    "600000000000000000000",
			expectError: false,
		},
		{
			name:        "simple decimal number with non-zero fractional part",
			input:       "123.456",
			expected:    "",
			expectError: true, // Should fail because fractional part is not all zeros
		},
		{
			name:        "decimal number with integer zero but non-zero fractional part",
			input:       "0.123456",
			expected:    "",
			expectError: true, // Should fail because fractional part is not all zeros
		},
		{
			name:        "large decimal number with non-zero fractional part",
			input:       "999999999999999999999.999999999",
			expected:    "",
			expectError: true, // Should fail because fractional part is not all zeros
		},
		{
			name:        "decimal with single digit",
			input:       "5.0",
			expected:    "5",
			expectError: false,
		},
		{
			name:        "decimal with many trailing zeros",
			input:       "42.00000000000000000000",
			expected:    "42",
			expectError: false,
		},
		{
			name:        "decimal starting with dot and zeros",
			input:       ".000",
			expected:    "0",
			expectError: false,
		},

		// Invalid inputs
		{
			name:        "invalid format",
			input:       "not-a-number",
			expected:    "",
			expectError: true,
		},
		{
			name:        "empty string",
			input:       "",
			expected:    "",
			expectError: true,
		},
		{
			name:        "invalid scientific notation missing exponent value",
			input:       "1e",
			expected:    "",
			expectError: true,
		},
		{
			name:        "multiple decimal points",
			input:       "1.2.3e10",
			expected:    "",
			expectError: true,
		},
		{
			name:        "invalid exponent",
			input:       "1eabc",
			expected:    "",
			expectError: true,
		},
		{
			name:        "just e",
			input:       "e10",
			expected:    "",
			expectError: true,
		},
		{
			name:        "just decimal point",
			input:       ".",
			expected:    "",
			expectError: true,
		},
		{
			name:        "multiple e characters",
			input:       "1e2e3",
			expected:    "",
			expectError: true,
		},
		{
			name:        "exponent with decimal",
			input:       "1e2.5",
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parsing.ScientificNotationToBigInt(tt.input)

			if tt.expectError {
				require.Error(t, err, "ScientificNotationToBigInt(%q) should return error", tt.input)
				return
			}

			require.NoError(t, err, "ScientificNotationToBigInt(%q) unexpected error: %v", tt.input, err)

			expected := new(big.Int)
			expected.SetString(tt.expected, 10)
			require.Equal(t, 0, result.Cmp(expected), "ScientificNotationToBigInt(%q) = %v, want %v", tt.input, result, expected)
		})
	}
}

func TestParsing_ScientificNotationToUint64(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    uint64
		expectError bool
	}{
		// Regular decimal numbers
		{
			name:        "simple integer",
			input:       "123456",
			expected:    123456,
			expectError: false,
		},
		{
			name:        "zero",
			input:       "0",
			expected:    0,
			expectError: false,
		},
		{
			name:        "max uint64",
			input:       "18446744073709551615",
			expected:    18446744073709551615,
			expectError: false,
		},

		// Scientific notation - positive exponents
		{
			name:        "simple scientific notation",
			input:       "1e6",
			expected:    1000000,
			expectError: false,
		},
		{
			name:        "scientific notation with decimal",
			input:       "1.5e+3",
			expected:    1500,
			expectError: false,
		},
		{
			name:        "scientific notation uppercase E",
			input:       "5E4",
			expected:    50000,
			expectError: false,
		},
		{
			name:        "large scientific notation",
			input:       "1.23e+15",
			expected:    1230000000000000,
			expectError: false,
		},
		{
			name:        "near max uint64 scientific",
			input:       "1.8e+19",
			expected:    18000000000000000000,
			expectError: false,
		},

		// Scientific notation - negative exponents (truncated to integer)
		{
			name:        "scientific notation negative exponent small",
			input:       "1.5e-1",
			expected:    0, // truncated to integer
			expectError: false,
		},
		{
			name:        "scientific notation that becomes integer",
			input:       "1.23e2",
			expected:    123,
			expectError: false,
		},

		// Decimal numbers
		{
			name:        "decimal with zero fraction",
			input:       "12345.000",
			expected:    12345,
			expectError: false,
		},

		// Error cases
		{
			name:        "invalid format",
			input:       "not-a-number",
			expected:    0,
			expectError: true,
		},
		{
			name:        "empty string",
			input:       "",
			expected:    0,
			expectError: true,
		},
		{
			name:        "negative number",
			input:       "-123",
			expected:    0,
			expectError: true,
		},
		{
			name:        "negative scientific notation",
			input:       "-1e5",
			expected:    0,
			expectError: true,
		},
		{
			name:        "value too large for uint64",
			input:       "1e100",
			expected:    0,
			expectError: true,
		},
		{
			name:        "just over max uint64",
			input:       "18446744073709551616",
			expected:    0,
			expectError: true,
		},
		{
			name:        "decimal with non-zero fraction",
			input:       "123.456",
			expected:    0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parsing.ScientificNotationToUint64(tt.input)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestParsing_ScientificNotationToUint8(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    uint8
		expectError bool
	}{
		// Regular decimal numbers
		{
			name:        "simple integer",
			input:       "123",
			expected:    123,
			expectError: false,
		},
		{
			name:        "zero",
			input:       "0",
			expected:    0,
			expectError: false,
		},
		{
			name:        "max uint8",
			input:       "255",
			expected:    255,
			expectError: false,
		},

		// Scientific notation - positive exponents
		{
			name:        "simple scientific notation",
			input:       "1e2",
			expected:    100,
			expectError: false,
		},
		{
			name:        "scientific notation with decimal",
			input:       "2.5e+1",
			expected:    25,
			expectError: false,
		},
		{
			name:        "scientific notation uppercase E",
			input:       "1E1",
			expected:    10,
			expectError: false,
		},
		{
			name:        "max uint8 via scientific",
			input:       "2.55e2",
			expected:    255,
			expectError: false,
		},

		// Scientific notation - negative exponents (truncated to integer)
		{
			name:        "scientific notation negative exponent small",
			input:       "1.5e-1",
			expected:    0, // truncated to integer
			expectError: false,
		},
		{
			name:        "scientific notation that becomes integer",
			input:       "1.23e1",
			expected:    12, // truncated
			expectError: false,
		},

		// Decimal numbers
		{
			name:        "decimal with zero fraction",
			input:       "100.0",
			expected:    100,
			expectError: false,
		},

		// Error cases
		{
			name:        "invalid format",
			input:       "not-a-number",
			expected:    0,
			expectError: true,
		},
		{
			name:        "empty string",
			input:       "",
			expected:    0,
			expectError: true,
		},
		{
			name:        "negative number",
			input:       "-1",
			expected:    0,
			expectError: true,
		},
		{
			name:        "value too large for uint8",
			input:       "256",
			expected:    0,
			expectError: true,
		},
		{
			name:        "scientific notation too large",
			input:       "1e3",
			expected:    0,
			expectError: true,
		},
		{
			name:        "just over max uint8 via scientific",
			input:       "2.56e2",
			expected:    0,
			expectError: true,
		},
		{
			name:        "negative scientific notation",
			input:       "-1e1",
			expected:    0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parsing.ScientificNotationToUint8(tt.input)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestParsing_ScientificNotationToBigInt_Precision verifies that large numbers
// are handled without float64 precision loss.
func TestParsing_ScientificNotationToBigInt_Precision(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "18 digit mantissa preserved",
			input:    "123456789012345678e5",
			expected: "12345678901234567800000",
		},
		{
			name:     "large mantissa with negative exponent",
			input:    "999999999999999999999999999999e-15",
			expected: "999999999999999",
		},
		{
			name:     "precise blockchain amount",
			input:    "1.234567890123456789e+18",
			expected: "1234567890123456789",
		},
		{
			name:     "wei to ether scale",
			input:    "1000000000000000000e-18",
			expected: "1",
		},
		{
			name:     "very large number with many decimal places",
			input:    "9.99999999999999999999999999e+26",
			expected: "999999999999999999999999999",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parsing.ScientificNotationToBigInt(tt.input)
			require.NoError(t, err, "parsing should succeed for %q", tt.input)

			expected, _ := new(big.Int).SetString(tt.expected, 10)
			require.Equal(t, 0, result.Cmp(expected), "precision mismatch for %q: got %v, want %v", tt.input, result, expected)
		})
	}
}

// TestParsing_ScientificNotationToBigInt_ErrorMessages verifies error messages are informative.
func TestParsing_ScientificNotationToBigInt_ErrorMessages(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		errorSubstring string
	}{
		{
			name:           "empty string error",
			input:          "",
			errorSubstring: "empty string",
		},
		{
			name:           "empty mantissa error",
			input:          "e10",
			errorSubstring: "empty mantissa",
		},
		{
			name:           "empty exponent error",
			input:          "1e",
			errorSubstring: "empty exponent",
		},
		{
			name:           "invalid exponent error",
			input:          "1eabc",
			errorSubstring: "invalid exponent",
		},
		{
			name:           "non-zero fractional part error",
			input:          "123.456",
			errorSubstring: "non-zero fractional part",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parsing.ScientificNotationToBigInt(tt.input)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.errorSubstring, "error should mention: %s", tt.errorSubstring)
		})
	}
}

