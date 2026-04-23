// Package parsing provides common parsing utilities for event decoding.
// It handles scientific notation and decimal number parsing for blockchain event data.
package parsing

import (
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

// Sentinel errors for numeric parsing.
var (
	ErrEmptyString                 = errors.New("empty string")
	ErrEmptyMantissa               = errors.New("empty mantissa")
	ErrEmptyExponent               = errors.New("empty exponent")
	ErrInvalidExponent             = errors.New("invalid exponent")
	ErrExponentOutOfRange          = errors.New("exponent out of allowed range")
	ErrInvalidMantissaNoDigits     = errors.New("invalid mantissa: no digits")
	ErrInvalidMantissa            = errors.New("invalid mantissa")
	ErrInvalidNumberFormat         = errors.New("invalid number format")
	ErrInvalidDecimalNoDigits      = errors.New("invalid decimal: no digits")
	ErrNonZeroFractionalPart       = errors.New("non-zero fractional part cannot be converted to integer")
	ErrInvalidIntegerPart          = errors.New("invalid integer part")
	ErrUnableParse                 = errors.New("unable to parse")
	ErrNegativeUint64              = errors.New("negative value not allowed for uint64")
	ErrUint64TooLarge              = errors.New("value too large for uint64")
	ErrUint32OutOfRange            = errors.New("value out of range for uint32")
	ErrUint16OutOfRange            = errors.New("value out of range for uint16")
	ErrUint8OutOfRange             = errors.New("value out of range for uint8")
)

// ScientificNotationToBigInt converts scientific notation strings to big.Int.
// It handles formats like "1.2e+21", "1e18", "-5e10" that big.Int.SetString cannot parse directly,
// and decimal numbers like "600000000000000000000.000000".
// Returns an error if the string is empty or parsing fails.
func ScientificNotationToBigInt(value string) (*big.Int, error) {
	if value == "" {
		return nil, ErrEmptyString
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
		return nil, ErrEmptyMantissa
	}
	if exponentStr == "" {
		return nil, ErrEmptyExponent
	}

	// Parse exponent, removing optional '+' sign
	exponentStr = strings.TrimPrefix(exponentStr, "+")
	exponent, err := strconv.Atoi(exponentStr)
	if err != nil {
		return nil, fmt.Errorf("%w: parsing %q: %w", ErrInvalidExponent, exponentStr, err)
	}

	const maxExponent = 100000 // Limit to prevent DoS via massive allocation
	if exponent < -maxExponent || exponent > maxExponent {
		return nil, fmt.Errorf("%w: exponent %d not in [-%d, %d]", ErrExponentOutOfRange, exponent, maxExponent, maxExponent)
	}

	// Parse mantissa, handling decimal point
	var mantissaBig *big.Int
	if dotIndex := strings.Index(mantissaStr, "."); dotIndex >= 0 {
		intPart := mantissaStr[:dotIndex]
		fracPart := mantissaStr[dotIndex+1:]

		// Validate parts
		if len(intPart) == 0 && len(fracPart) == 0 {
			return nil, ErrInvalidMantissaNoDigits
		}

		// Handle case where integer part is just a sign or empty
		combinedStr := intPart + fracPart
		if combinedStr == "" || combinedStr == "-" || combinedStr == "+" {
			return nil, fmt.Errorf("%w: %q", ErrInvalidMantissa, mantissaStr)
		}

		var ok bool
		mantissaBig, ok = new(big.Int).SetString(combinedStr, 10)
		if !ok {
			return nil, fmt.Errorf("%w: %q", ErrInvalidMantissa, mantissaStr)
		}
		exponent -= len(fracPart)
	} else {
		var ok bool
		mantissaBig, ok = new(big.Int).SetString(mantissaStr, 10)
		if !ok {
			return nil, fmt.Errorf("%w: %q", ErrInvalidMantissa, mantissaStr)
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
		return nil, fmt.Errorf("%w: %q", ErrInvalidNumberFormat, value)
	}

	intPart := value[:dotIndex]
	fracPart := value[dotIndex+1:]

	// Validate we have something to parse
	if intPart == "" && fracPart == "" {
		return nil, ErrInvalidDecimalNoDigits
	}

	// Only allow conversion if fractional part is all zeros
	if strings.Trim(fracPart, "0") != "" {
		return nil, fmt.Errorf("%w: %q", ErrNonZeroFractionalPart, value)
	}

	// Handle empty integer part (e.g., ".000")
	if intPart == "" {
		return big.NewInt(0), nil
	}

	result, ok := new(big.Int).SetString(intPart, 10)
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrInvalidIntegerPart, intPart)
	}
	return result, nil
}

// ScientificNotationToUint64 converts scientific notation strings to uint64.
// It handles formats like "1.2e+21", "1e18" that strconv.ParseUint cannot parse directly.
// Returns an error if the value cannot be parsed, is negative, or exceeds uint64 range.
func ScientificNotationToUint64(value string) (uint64, error) {
	// Fast path: try direct parsing
	if result, err := strconv.ParseUint(value, 10, 64); err == nil {
		return result, nil
	}

	// Handle scientific notation using the big.Int parser and then convert
	bigIntResult, err := ScientificNotationToBigInt(value)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrUnableParse, err)
	}

	// Check for negative values
	if bigIntResult.Sign() < 0 {
		return 0, fmt.Errorf("%w: %s", ErrNegativeUint64, value)
	}

	// Check if the result fits in uint64
	if !bigIntResult.IsUint64() {
		return 0, fmt.Errorf("%w: %s", ErrUint64TooLarge, value)
	}

	return bigIntResult.Uint64(), nil
}

// ScientificNotationToUint32 converts scientific notation strings to uint32.
// It handles formats like "1e2", "2.5e+1" that strconv.ParseUint cannot parse directly.
// Returns an error if the value cannot be parsed or is outside the uint32 range (0 to 4294967295).
func ScientificNotationToUint32(value string) (uint32, error) {
	// Fast path: try direct parsing
	if result, err := strconv.ParseUint(value, 10, 32); err == nil {
		return uint32(result), nil
	}

	// Handle scientific notation using the big.Int parser and then convert
	bigIntResult, err := ScientificNotationToBigInt(value)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrUnableParse, err)
	}

	// Check if the result fits in uint32 (0 to 4294967295)
	if bigIntResult.Sign() < 0 || bigIntResult.Cmp(big.NewInt(4294967295)) > 0 {
		return 0, fmt.Errorf("%w: %s", ErrUint32OutOfRange, value)
	}

	return uint32(bigIntResult.Uint64()), nil
}

// ScientificNotationToUint16 converts scientific notation strings to uint16.
// It handles formats like "1e2", "2.5e+1" that strconv.ParseUint cannot parse directly.
// Returns an error if the value cannot be parsed or is outside the uint16 range (0 to 65535).
func ScientificNotationToUint16(value string) (uint16, error) {
	// Fast path: try direct parsing
	if result, err := strconv.ParseUint(value, 10, 16); err == nil {
		return uint16(result), nil
	}

	// Handle scientific notation using the big.Int parser and then convert
	bigIntResult, err := ScientificNotationToBigInt(value)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrUnableParse, err)
	}

	// Check if the result fits in uint16 (0 to 65535)
	if bigIntResult.Sign() < 0 || bigIntResult.Cmp(big.NewInt(65535)) > 0 {
		return 0, fmt.Errorf("%w: %s", ErrUint16OutOfRange, value)
	}

	return uint16(bigIntResult.Uint64()), nil
}

// ScientificNotationToUint8 converts scientific notation strings to uint8.
// It handles formats like "1e2", "2.5e+1" that strconv.ParseUint cannot parse directly.
// Returns an error if the value cannot be parsed or is outside the uint8 range (0 to 255).
func ScientificNotationToUint8(value string) (uint8, error) {
	// Fast path: try direct parsing
	if result, err := strconv.ParseUint(value, 10, 8); err == nil {
		return uint8(result), nil
	}

	// Handle scientific notation using the big.Int parser and then convert
	bigIntResult, err := ScientificNotationToBigInt(value)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrUnableParse, err)
	}

	// Check if the result fits in uint8 (0-255)
	if bigIntResult.Sign() < 0 || bigIntResult.Cmp(big.NewInt(255)) > 0 {
		return 0, fmt.Errorf("%w: %s", ErrUint8OutOfRange, value)
	}

	return uint8(bigIntResult.Uint64()), nil
}
