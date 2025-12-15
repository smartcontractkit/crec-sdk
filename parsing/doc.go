// Package parsing provides common utilities for parsing blockchain event data.
//
// This package handles the conversion of string representations of numbers
// (including scientific notation) to Go numeric types. It's particularly useful
// for parsing event data from blockchain watchers where numbers may be
// represented in scientific notation (e.g., "1.2e+21") or as decimal strings
// with trailing zeros (e.g., "600000000000000000000.000000").
//
// # Primary Functions
//
//   - ScientificNotationToBigInt: Converts strings to *big.Int
//   - ScientificNotationToUint64: Converts strings to uint64
//   - ScientificNotationToUint32: Converts strings to uint32
//   - ScientificNotationToUint16: Converts strings to uint16
//   - ScientificNotationToUint8: Converts strings to uint8
//
// All functions use arbitrary-precision arithmetic internally to avoid
// float64 precision loss issues that would occur with naive string parsing.
//
// # Example Usage
//
//	value, err := parsing.ScientificNotationToBigInt("1.5e+18")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	// value is 1500000000000000000
package parsing

