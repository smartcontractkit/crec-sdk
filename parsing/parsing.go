// Package parsing provides common parsing utilities for event decoding.
// It handles scientific notation and decimal number parsing for blockchain event data.
package parsing

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

// ScientificNotationToBigInt converts scientific notation strings to big.Int.
// Handles formats like "1.2e+21", "1e18", "-5e10", etc. that big.Int.SetString cannot parse directly.
// Also handles decimal numbers like "600000000000000000000.000000".
// Returns an error if parsing fails.
func ScientificNotationToBigInt(value string) (*big.Int, error) {
	if value == "" {
		return nil, fmt.Errorf("empty string")
	}

	// Fast path: try direct integer parsing
	if result, ok := new(big.Int).SetString(value, 10); ok {
		return result, nil
	}

	// Check for scientific notation (e or E) without allocating a lowercase copy
	eIndex := strings.IndexAny(value, "eE")
	if eIndex >= 0 {
		return parseScientificNotation(value[:eIndex], value[eIndex+1:])
	}

	// Handle plain decimal numbers (e.g., "123.000000")
	return parseDecimalNumber(value)
}

// parseScientificNotation handles the "mantissa e exponent" format using pure big.Int arithmetic.
func parseScientificNotation(mantissaStr, exponentStr string) (*big.Int, error) {
	if mantissaStr == "" {
		return nil, fmt.Errorf("empty mantissa")
	}
	if exponentStr == "" {
		return nil, fmt.Errorf("empty exponent")
	}

	// Parse exponent, removing optional '+' sign
	exponentStr = strings.TrimPrefix(exponentStr, "+")
	exponent, err := strconv.Atoi(exponentStr)
	if err != nil {
		return nil, fmt.Errorf("invalid exponent %q: %w", exponentStr, err)
	}

	// Parse mantissa, handling decimal point
	var mantissaBig *big.Int
	if dotIndex := strings.Index(mantissaStr, "."); dotIndex >= 0 {
		intPart := mantissaStr[:dotIndex]
		fracPart := mantissaStr[dotIndex+1:]

		// Validate parts
		if len(intPart) == 0 && len(fracPart) == 0 {
			return nil, fmt.Errorf("invalid mantissa: no digits")
		}

		// Handle case where integer part is just a sign or empty
		combinedStr := intPart + fracPart
		if combinedStr == "" || combinedStr == "-" || combinedStr == "+" {
			return nil, fmt.Errorf("invalid mantissa %q", mantissaStr)
		}

		var ok bool
		mantissaBig, ok = new(big.Int).SetString(combinedStr, 10)
		if !ok {
			return nil, fmt.Errorf("invalid mantissa %q", mantissaStr)
		}
		exponent -= len(fracPart)
	} else {
		var ok bool
		mantissaBig, ok = new(big.Int).SetString(mantissaStr, 10)
		if !ok {
			return nil, fmt.Errorf("invalid mantissa %q", mantissaStr)
		}
	}

	// Apply exponent using big.Int arithmetic (no float64 precision loss)
	if exponent > 0 {
		multiplier := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(exponent)), nil)
		return mantissaBig.Mul(mantissaBig, multiplier), nil
	} else if exponent < 0 {
		// Divide and truncate toward zero
		divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(-exponent)), nil)
		return mantissaBig.Div(mantissaBig, divisor), nil
	}
	return mantissaBig, nil
}

// parseDecimalNumber handles plain decimal numbers like "123.000000".
// Only succeeds if the fractional part is all zeros (can be safely truncated).
func parseDecimalNumber(value string) (*big.Int, error) {
	dotIndex := strings.Index(value, ".")
	if dotIndex < 0 {
		return nil, fmt.Errorf("invalid number format: %q", value)
	}

	intPart := value[:dotIndex]
	fracPart := value[dotIndex+1:]

	// Validate we have something to parse
	if intPart == "" && fracPart == "" {
		return nil, fmt.Errorf("invalid decimal: no digits")
	}

	// Only allow conversion if fractional part is all zeros
	if strings.Trim(fracPart, "0") != "" {
		return nil, fmt.Errorf("non-zero fractional part cannot be converted to integer: %q", value)
	}

	// Handle empty integer part (e.g., ".000")
	if intPart == "" {
		return big.NewInt(0), nil
	}

	result, ok := new(big.Int).SetString(intPart, 10)
	if !ok {
		return nil, fmt.Errorf("invalid integer part: %q", intPart)
	}
	return result, nil
}

// ScientificNotationToUint64 converts scientific notation strings to uint64.
// Handles formats like "1.2e+21", "1e18", etc. that strconv.ParseUint cannot parse directly.
// Returns an error if the value cannot be parsed or is too large for uint64.
func ScientificNotationToUint64(value string) (uint64, error) {
	// Fast path: try direct parsing
	if result, err := strconv.ParseUint(value, 10, 64); err == nil {
		return result, nil
	}

	// Handle scientific notation using the big.Int parser and then convert
	bigIntResult, err := ScientificNotationToBigInt(value)
	if err != nil {
		return 0, fmt.Errorf("unable to parse: %w", err)
	}

	// Check for negative values
	if bigIntResult.Sign() < 0 {
		return 0, fmt.Errorf("negative value not allowed for uint64: %s", value)
	}

	// Check if the result fits in uint64
	if !bigIntResult.IsUint64() {
		return 0, fmt.Errorf("value too large for uint64: %s", value)
	}

	return bigIntResult.Uint64(), nil
}

// ScientificNotationToUint32 converts scientific notation strings to uint32.
// Handles formats like "1e2", "2.5e+1", etc. that strconv.ParseUint cannot parse directly.
// Returns an error if the value cannot be parsed or is out of range for uint32.
func ScientificNotationToUint32(value string) (uint32, error) {
	// Fast path: try direct parsing
	if result, err := strconv.ParseUint(value, 10, 32); err == nil {
		return uint32(result), nil
	}

	// Handle scientific notation using the big.Int parser and then convert
	bigIntResult, err := ScientificNotationToBigInt(value)
	if err != nil {
		return 0, fmt.Errorf("unable to parse: %w", err)
	}

	// Check if the result fits in uint32 (0 to 4294967295)
	if bigIntResult.Sign() < 0 || bigIntResult.Cmp(big.NewInt(4294967295)) > 0 {
		return 0, fmt.Errorf("value out of range for uint32: %s", value)
	}

	return uint32(bigIntResult.Uint64()), nil
}

// ScientificNotationToUint16 converts scientific notation strings to uint16.
// Handles formats like "1e2", "2.5e+1", etc. that strconv.ParseUint cannot parse directly.
// Returns an error if the value cannot be parsed or is out of range for uint16.
func ScientificNotationToUint16(value string) (uint16, error) {
	// Fast path: try direct parsing
	if result, err := strconv.ParseUint(value, 10, 16); err == nil {
		return uint16(result), nil
	}

	// Handle scientific notation using the big.Int parser and then convert
	bigIntResult, err := ScientificNotationToBigInt(value)
	if err != nil {
		return 0, fmt.Errorf("unable to parse: %w", err)
	}

	// Check if the result fits in uint16 (0 to 65535)
	if bigIntResult.Sign() < 0 || bigIntResult.Cmp(big.NewInt(65535)) > 0 {
		return 0, fmt.Errorf("value out of range for uint16: %s", value)
	}

	return uint16(bigIntResult.Uint64()), nil
}

// ScientificNotationToUint8 converts scientific notation strings to uint8.
// Handles formats like "1e2", "2.5e+1", etc. that strconv.ParseUint cannot parse directly.
// Returns an error if the value cannot be parsed or is out of range for uint8.
func ScientificNotationToUint8(value string) (uint8, error) {
	// Fast path: try direct parsing
	if result, err := strconv.ParseUint(value, 10, 8); err == nil {
		return uint8(result), nil
	}

	// Handle scientific notation using the big.Int parser and then convert
	bigIntResult, err := ScientificNotationToBigInt(value)
	if err != nil {
		return 0, fmt.Errorf("unable to parse: %w", err)
	}

	// Check if the result fits in uint8 (0-255)
	if bigIntResult.Sign() < 0 || bigIntResult.Cmp(big.NewInt(255)) > 0 {
		return 0, fmt.Errorf("value out of range for uint8: %s", value)
	}

	return uint8(bigIntResult.Uint64()), nil
}

