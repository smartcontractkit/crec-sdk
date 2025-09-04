package dta

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// ParseAttributes parses ve.Metadata.WorkflowEvent.Attributes into the concrete struct, and returns it as interface{}.
func (v VerifiableEvent) ParseAttributes() (any, error) {
	attrs := v.Metadata.WorkflowEvent.Attributes
	if attrs == nil {
		return nil, errors.New("no attributes")
	}

	rt, ok := eventTypeRegistry[v.EventName()]
	if !ok {
		return nil, fmt.Errorf("unknown event type: %s", v.EventName())
	}

	// Create a new instance of the struct (pointer), then fill it.
	ptr := reflect.New(rt) // *T
	if err := fillFromAttributes(ptr.Elem(), attrs); err != nil {
		return nil, fmt.Errorf("decode %s: %w", v.EventName(), err)
	}
	return ptr.Elem().Interface(), nil
}

// fillFromAttributes sets fields in v (a struct value) from attrs using each field's json tag.
func fillFromAttributes(v reflect.Value, attrs map[string]Attribute) error {
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if sf.PkgPath != "" { // unexported
			continue
		}
		// If the field is an embedded struct (no json tag), recurse
		if sf.Anonymous && sf.Type.Kind() == reflect.Struct {
			if err := fillFromAttributes(v.Field(i), attrs); err != nil {
				return err
			}
			continue
		}

		tag := sf.Tag.Get("json")
		key := strings.Split(tag, ",")[0]
		if key == "" || key == "-" {
			continue
		}
		attr, ok := attrs[key]
		if !ok {
			// Not present: skip (leave zero value)
			continue
		}

		dst := v.Field(i)
		if !dst.CanSet() {
			continue
		}
		val := strings.TrimSpace(attr.Value)

		if err := setValue(dst, val); err != nil {
			return fmt.Errorf("field %s (%s): %w", sf.Name, key, err)
		}
	}
	return nil
}

// setValue converts the string "val" into the type of dst and assigns it.
func setValue(dst reflect.Value, val string) error {
	// Handle pointers specially (e.g., *big.Int)
	if dst.Kind() == reflect.Ptr {
		elemT := dst.Type().Elem()
		switch elemT {
		case reflect.TypeOf(big.Int{}):
			bi, err := parseBigInt(val)
			if err != nil {
				return err
			}
			dst.Set(reflect.ValueOf(bi))
			return nil
		default:
			// For other pointer-to-T, try to allocate and set recursively
			ptr := reflect.New(elemT)
			if err := setValue(ptr.Elem(), val); err != nil {
				return err
			}
			dst.Set(ptr)
			return nil
		}
	}

	// Non-pointers
	switch dst.Type() {
	case reflect.TypeOf(common.Address{}):
		dst.Set(reflect.ValueOf(common.HexToAddress(val)))
		return nil
	case reflect.TypeOf([32]byte{}):
		b32, err := parseBytes32(val)
		if err != nil {
			return err
		}
		dst.Set(reflect.ValueOf(b32))
		return nil
	case reflect.TypeOf([]byte{}):
		bs, err := parseBytes(val)
		if err != nil {
			return err
		}
		dst.SetBytes(bs)
		return nil
	case reflect.TypeOf(big.Int{}):
		bi, err := parseBigInt(val)
		if err != nil {
			return err
		}
		dst.Set(reflect.ValueOf(*bi))
		return nil
	}

	switch dst.Kind() {
	case reflect.String:
		dst.SetString(val)
		return nil
	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		u, err := parseUint(val, dst.Kind())
		if err != nil {
			return err
		}
		dst.SetUint(u)
		return nil
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		// Not used in event structs, but keep for completeness
		i, err := parseInt(val, dst.Kind())
		if err != nil {
			return err
		}
		dst.SetInt(i)
		return nil
	case reflect.Bool:
		b, err := parseBool(val)
		if err != nil {
			return err
		}
		dst.SetBool(b)
		return nil
	case reflect.Struct:
		// Special-case time.Time if you ever map it directly from attributes.
		if dst.Type() == reflect.TypeOf(time.Time{}) {
			// Attempt to parse RFC3339, fallback to unix secs
			if ts, err := time.Parse(time.RFC3339, val); err == nil {
				dst.Set(reflect.ValueOf(ts))
				return nil
			}
			if u, err := strconv.ParseInt(val, 10, 64); err == nil {
				dst.Set(reflect.ValueOf(time.Unix(u, 0).UTC()))
				return nil
			}
			return fmt.Errorf("cannot parse time: %q", val)
		}
	}

	return fmt.Errorf("unsupported destination type: %s", dst.Type().String())
}

// -------------------- parsers --------------------

func parseBigInt(s string) (*big.Int, error) {
	base := 10
	raw := strings.TrimSpace(s)
	if raw == "" {
		return big.NewInt(0), nil
	}
	if strings.HasPrefix(raw, "0x") || strings.HasPrefix(raw, "0X") {
		base = 16
		raw = raw[2:]
	}
	bi := new(big.Int)
	_, ok := bi.SetString(raw, base)
	if !ok {
		return nil, fmt.Errorf("invalid big.Int: %q", s)
	}
	return bi, nil
}

func parseUint(s string, kind reflect.Kind) (uint64, error) {
	base := 10
	raw := strings.TrimSpace(s)
	if raw == "" {
		return 0, nil
	}
	if strings.HasPrefix(raw, "0x") || strings.HasPrefix(raw, "0X") {
		base = 16
		raw = raw[2:]
	}
	u, err := strconv.ParseUint(raw, base, bitSizeForKind(kind))
	if err != nil {
		return 0, fmt.Errorf("invalid uint: %q: %w", s, err)
	}
	return u, nil
}

func parseInt(s string, kind reflect.Kind) (int64, error) {
	base := 10
	raw := strings.TrimSpace(s)
	if raw == "" {
		return 0, nil
	}
	neg := false
	if strings.HasPrefix(raw, "-") {
		neg = true
		raw = raw[1:]
	}
	if strings.HasPrefix(raw, "0x") || strings.HasPrefix(raw, "0X") {
		base = 16
	}
	i, err := strconv.ParseInt(raw, base, bitSizeForKind(kind))
	if err != nil {
		return 0, fmt.Errorf("invalid int: %q: %w", s, err)
	}
	if neg {
		i = -i
	}
	return i, nil
}

func parseBool(s string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "true", "t", "yes", "y":
		return true, nil
	case "0", "false", "f", "no", "n":
		return false, nil
	default:
		return false, fmt.Errorf("invalid bool: %q", s)
	}
}

func parseBytes32(s string) ([32]byte, error) {
	var out [32]byte
	b, err := parseBytes(s)
	if err != nil {
		return out, err
	}
	if len(b) != 32 {
		return out, fmt.Errorf("expected 32 bytes, got %d", len(b))
	}
	copy(out[:], b)
	return out, nil
}

func parseBytes(s string) ([]byte, error) {
	raw := strings.TrimSpace(s)
	if raw == "" {
		return []byte{}, nil
	}
	if strings.HasPrefix(raw, "0x") || strings.HasPrefix(raw, "0X") {
		raw = raw[2:]
	}
	if len(raw)%2 != 0 {
		// pad if needed (some emit odd-length hex)
		raw = "0" + raw
	}
	b, err := hex.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid hex: %w", err)
	}
	return b, nil
}

func bitSizeForKind(k reflect.Kind) int {
	switch k {
	case reflect.Uint8, reflect.Int8:
		return 8
	case reflect.Uint16, reflect.Int16:
		return 16
	case reflect.Uint32, reflect.Int32:
		return 32
	case reflect.Uint64, reflect.Int64:
		return 64
	default:
		return 0 // let ParseInt/Uint decide
	}
}
