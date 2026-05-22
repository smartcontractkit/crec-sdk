package queries

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
)

// CallContractABIResult is a raw call result plus ABI-decoded outputs.
type CallContractABIResult struct {
	*CallContractResult
	Outputs []any
}

// CallContractWithABI ABI-encodes function arguments, performs the query, and decodes return bytes.
func (c *Client) CallContractWithABI(
	ctx context.Context,
	channelID uuid.UUID,
	chainSelector string,
	contractAddress string,
	abiFragment string,
	functionName string,
	args []any,
	blockSelection BlockSelection,
	idempotencyKey string,
	options ...CallOption,
) (*CallContractABIResult, error) {
	contractABI, method, err := parseABIFragment(abiFragment, functionName)
	if err != nil {
		return nil, err
	}

	coercedArgs, err := coerceABIArgs(method.Inputs, args)
	if err != nil {
		return nil, err
	}

	callData, err := contractABI.Pack(method.Name, coercedArgs...)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrABIArgumentType, err)
	}

	rawResult, err := c.CallContract(
		ctx,
		channelID,
		chainSelector,
		contractAddress,
		callData,
		blockSelection,
		idempotencyKey,
		options...,
	)
	if err != nil {
		return nil, err
	}

	result := &CallContractABIResult{CallContractResult: rawResult}
	if rawResult.Error != nil || len(method.Outputs) == 0 {
		return result, nil
	}

	outputs, err := method.Outputs.Unpack(rawResult.RawReturnData)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDecodeABIOutput, err)
	}
	result.Outputs = outputs

	return result, nil
}

func parseABIFragment(fragment string, functionName string) (abi.ABI, abi.Method, error) {
	fragment = strings.TrimSpace(fragment)
	if fragment == "" {
		return abi.ABI{}, abi.Method{}, ErrABIRequired
	}

	normalized, err := normalizeABIFragment(fragment)
	if err != nil {
		return abi.ABI{}, abi.Method{}, err
	}

	parsedABI, err := abi.JSON(strings.NewReader(normalized))
	if err != nil {
		return abi.ABI{}, abi.Method{}, fmt.Errorf("%w: %w", ErrParseABI, err)
	}

	if functionName == "" {
		if len(parsedABI.Methods) != 1 {
			return abi.ABI{}, abi.Method{}, ErrABIFunctionNameRequired
		}
		for name := range parsedABI.Methods {
			functionName = name
			break
		}
	}

	method, ok := parsedABI.Methods[functionName]
	if !ok {
		return abi.ABI{}, abi.Method{}, fmt.Errorf("%w: %s", ErrABIFunctionNotFound, functionName)
	}

	return parsedABI, method, nil
}

func normalizeABIFragment(fragment string) (string, error) {
	if json.Valid([]byte(fragment)) {
		if strings.HasPrefix(fragment, "{") {
			return "[" + fragment + "]", nil
		}
		return fragment, nil
	}

	humanReadable, err := humanReadableFunctionToABIJSON(fragment)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrParseABI, err)
	}
	return humanReadable, nil
}

func humanReadableFunctionToABIJSON(fragment string) (string, error) {
	fragment = strings.TrimSpace(fragment)
	fragment = strings.TrimPrefix(fragment, "function ")
	fragment = strings.TrimSpace(fragment)

	open := strings.Index(fragment, "(")
	if open <= 0 {
		return "", fmt.Errorf("missing function name or inputs")
	}
	close := strings.Index(fragment[open:], ")")
	if close < 0 {
		return "", fmt.Errorf("missing closing parenthesis")
	}
	close += open

	name := strings.TrimSpace(fragment[:open])
	if name == "" {
		return "", fmt.Errorf("missing function name")
	}

	inputs, err := humanReadableABIArguments(fragment[open+1 : close])
	if err != nil {
		return "", err
	}

	outputs := []map[string]string{}
	rest := strings.TrimSpace(fragment[close+1:])
	if returnsIndex := strings.Index(rest, "returns"); returnsIndex >= 0 {
		returnsPart := strings.TrimSpace(rest[returnsIndex+len("returns"):])
		returnOpen := strings.Index(returnsPart, "(")
		returnClose := strings.LastIndex(returnsPart, ")")
		if returnOpen < 0 || returnClose <= returnOpen {
			return "", fmt.Errorf("invalid returns clause")
		}
		outputs, err = humanReadableABIArguments(returnsPart[returnOpen+1 : returnClose])
		if err != nil {
			return "", err
		}
	}

	item := map[string]any{
		"type":            "function",
		"name":            name,
		"inputs":          inputs,
		"outputs":         outputs,
		"stateMutability": "view",
	}
	b, err := json.Marshal([]map[string]any{item})
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func humanReadableABIArguments(list string) ([]map[string]string, error) {
	list = strings.TrimSpace(list)
	if list == "" {
		return []map[string]string{}, nil
	}

	parts := strings.Split(list, ",")
	args := make([]map[string]string, 0, len(parts))
	for i, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			return nil, fmt.Errorf("empty argument")
		}
		fields := strings.Fields(part)
		if len(fields) == 0 {
			return nil, fmt.Errorf("empty argument")
		}

		argType := fields[0]
		argName := fmt.Sprintf("arg%d", i)
		if len(fields) > 1 {
			argName = fields[1]
		}

		args = append(args, map[string]string{
			"type": argType,
			"name": argName,
		})
	}
	return args, nil
}

func coerceABIArgs(inputs abi.Arguments, args []any) ([]any, error) {
	if args == nil {
		args = []any{}
	}
	if len(inputs) != len(args) {
		return nil, fmt.Errorf("%w: expected %d got %d", ErrABIArgumentCount, len(inputs), len(args))
	}

	coerced := make([]any, len(args))
	for i, input := range inputs {
		v, err := coerceABIArg(input.Type, args[i])
		if err != nil {
			return nil, fmt.Errorf("%w: argument %d (%s): %w", ErrABIArgumentType, i, input.Name, err)
		}
		coerced[i] = v
	}
	return coerced, nil
}

func coerceABIArg(typ abi.Type, value any) (any, error) {
	switch typ.T {
	case abi.AddressTy:
		switch v := value.(type) {
		case common.Address:
			return v, nil
		case string:
			if !common.IsHexAddress(v) {
				return nil, fmt.Errorf("invalid address %q", v)
			}
			return common.HexToAddress(v), nil
		default:
			return nil, fmt.Errorf("expected address, got %T", value)
		}

	case abi.IntTy:
		return integerToBigInt(value, false, typ.Size)

	case abi.UintTy:
		return integerToBigInt(value, true, typ.Size)

	case abi.BoolTy:
		v, ok := value.(bool)
		if !ok {
			return nil, fmt.Errorf("expected bool, got %T", value)
		}
		return v, nil

	case abi.StringTy:
		v, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("expected string, got %T", value)
		}
		return v, nil

	case abi.BytesTy:
		return bytesFromABIValue(value)

	case abi.FixedBytesTy:
		b, err := bytesFromABIValue(value)
		if err != nil {
			return nil, err
		}
		if len(b) != typ.Size {
			return nil, fmt.Errorf("expected bytes%d, got %d bytes", typ.Size, len(b))
		}
		arrayType := reflect.ArrayOf(typ.Size, reflect.TypeOf(byte(0)))
		arrayValue := reflect.New(arrayType).Elem()
		reflect.Copy(arrayValue, reflect.ValueOf(b))
		return arrayValue.Interface(), nil

	default:
		return value, nil
	}
}

func integerToBigInt(value any, unsigned bool, bitSize int) (*big.Int, error) {
	var n *big.Int

	switch v := value.(type) {
	case *big.Int:
		if v == nil {
			return nil, fmt.Errorf("nil *big.Int")
		}
		n = new(big.Int).Set(v)
	case big.Int:
		n = new(big.Int).Set(&v)
	case string:
		parsed := new(big.Int)
		var ok bool
		if strings.HasPrefix(v, "0x") || strings.HasPrefix(v, "0X") {
			_, ok = parsed.SetString(v[2:], 16)
		} else {
			_, ok = parsed.SetString(v, 10)
		}
		if !ok {
			return nil, fmt.Errorf("invalid integer string %q", v)
		}
		n = parsed
	case int:
		n = big.NewInt(int64(v))
	case int8:
		n = big.NewInt(int64(v))
	case int16:
		n = big.NewInt(int64(v))
	case int32:
		n = big.NewInt(int64(v))
	case int64:
		n = big.NewInt(v)
	case uint:
		n = new(big.Int).SetUint64(uint64(v))
	case uint8:
		n = new(big.Int).SetUint64(uint64(v))
	case uint16:
		n = new(big.Int).SetUint64(uint64(v))
	case uint32:
		n = new(big.Int).SetUint64(uint64(v))
	case uint64:
		n = new(big.Int).SetUint64(v)
	case float64:
		if math.Trunc(v) != v {
			return nil, fmt.Errorf("float64 value is not an integer")
		}
		if v > 9007199254740991 || v < -9007199254740991 {
			return nil, fmt.Errorf("float64 value may have lost integer precision")
		}
		n = big.NewInt(int64(v))
	default:
		return nil, fmt.Errorf("expected integer, got %T", value)
	}

	if unsigned && n.Sign() < 0 {
		return nil, fmt.Errorf("negative value for unsigned integer")
	}

	if bitSize > 0 {
		if unsigned {
			max := new(big.Int).Lsh(big.NewInt(1), uint(bitSize))
			if n.Sign() < 0 || n.Cmp(max) >= 0 {
				return nil, fmt.Errorf("value exceeds uint%d", bitSize)
			}
		} else {
			max := new(big.Int).Lsh(big.NewInt(1), uint(bitSize-1))
			min := new(big.Int).Neg(max)
			max.Sub(max, big.NewInt(1))
			if n.Cmp(min) < 0 || n.Cmp(max) > 0 {
				return nil, fmt.Errorf("value exceeds int%d", bitSize)
			}
		}
	}

	return n, nil
}

func bytesFromABIValue(value any) ([]byte, error) {
	switch v := value.(type) {
	case []byte:
		return v, nil
	case string:
		if strings.HasPrefix(v, "0x") || strings.HasPrefix(v, "0X") {
			return hex.DecodeString(v[2:])
		}
		return []byte(v), nil
	case common.Hash:
		return v.Bytes(), nil
	case [32]byte:
		return v[:], nil
	default:
		return nil, fmt.Errorf("expected bytes, got %T", value)
	}
}
