package workflows

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"testing"

	httpcap "github.com/smartcontractkit/cre-sdk-go/capabilities/networking/http"
	httpmock "github.com/smartcontractkit/cre-sdk-go/capabilities/networking/http/mock"
	"github.com/smartcontractkit/cre-sdk-go/cre/testutils"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

const testABIForCommon = `[
  {"type":"event","name":"Sender","inputs":[{"name":"sender","type":"address","indexed":true,"internalType":"address"}],"anonymous":false}
]`

func TestParseCursor(t *testing.T) {
	ci, err := ParseCursor("100-2-0xabc")
	require.NoError(t, err)
	require.Equal(t, uint64(100), ci.BlockNumber)
	require.Equal(t, uint64(2), ci.LogIndex)
	require.Equal(t, "0xabc", ci.TxHash)

	_, err = ParseCursor("bad-cursor")
	require.Error(t, err)
}

func TestBuildAndSignEventEnvelope_IncludesParametersInEventAndTopLevel(t *testing.T) {
	// prepare parameters: sender address encoded as bytes -> sanitised to hex
	addr := common.HexToAddress("0x1234567890123456789012345678901234567890")
	raw := map[string]any{
		"Sender": addr.Bytes(),
	}
	params := SanitiseJSON(raw).(map[string]any)

	res, err := BuildAndHashEventEnvelope(
		"test_service",
		"Sender",
		"0xContract",
		testABIForCommon,
		"1",
		100, // blockNumber
		2,   // logIndex
		"0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		1_700_000_000,
		params,
		map[string]any{"extra": "meta"},
	)
	require.NoError(t, err)
	require.NotEmpty(t, res.Base64Event)
	require.Equal(t, "test_service.Sender", res.Type)
	require.NotEqual(t, common.Hash{}, res.EventHash)

	// decode and validate shape
	decoded, err := base64.StdEncoding.DecodeString(res.Base64Event)
	require.NoError(t, err)
	var obj map[string]any
	require.NoError(t, json.Unmarshal(decoded, &obj))

	ev := obj["event"].(map[string]any)
	topParams := obj["parameters"].(map[string]any)
	evParams := ev["parameters"].(map[string]any)

	// both places must contain the same keys/value (additive exposure)
	require.Equal(t, topParams["sender"], "0x1234567890123456789012345678901234567890")
	require.Equal(t, evParams["sender"], "0x1234567890123456789012345678901234567890")

	// block/log info exposed
	require.Equal(t, float64(2), ev["log_index"])      // JSON unmarshals numbers to float64
	require.Equal(t, float64(100), ev["block_number"]) // JSON unmarshals numbers to float64

	// transaction fields contain hash and block_number
	tx := obj["transaction"].(map[string]any)
	require.Equal(t, "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", tx["hash"])
	require.Equal(t, float64(100), tx["block_number"])

	// metadata carried through
	meta := obj["metadata"].(map[string]any)
	require.Equal(t, "meta", meta["extra"])
}

func TestSanitiseJSON_Conversions(t *testing.T) {
	// prepare a structure with varied types
	b20 := make([]byte, 20) // 20 bytes -> address-like
	b32 := make([]byte, 32) // 32 bytes -> bytes32-like
	b64Addr := base64.StdEncoding.EncodeToString(b20)
	b64Bytes := base64.StdEncoding.EncodeToString(b32)

	raw := map[string]any{
		"Address":     b20,
		"Bytes32":     b32,
		"Base64Addr":  b64Addr,
		"Base64Bytes": b64Bytes,
		"Big":         big.NewInt(1234567890),
		"Bool":        true,
		"SmallInt":    42,
		// Also simulate JSON-decoded numeric arrays for 20/32 bytes
		"Arr20": func() []any {
			a := make([]any, 20)
			for i := range a {
				a[i] = float64(i)
			}
			return a
		}(),
		"Arr32": func() []any {
			a := make([]any, 32)
			for i := range a {
				a[i] = float64(i)
			}
			return a
		}(),
	}

	out := SanitiseJSON(raw).(map[string]any)

	// []byte -> hex
	require.Regexp(t, `^0x[0-9a-f]{40}$`, out["address"].(string))
	require.Regexp(t, `^0x[0-9a-f]{64}$`, out["bytes32"].(string))

	// base64 -> hex (20 or 32 bytes only)
	require.Regexp(t, `^0x[0-9a-f]{40}$`, out["base64_addr"].(string))
	require.Regexp(t, `^0x[0-9a-f]{64}$`, out["base64_bytes"].(string))

	// numeric []interface{} -> hex (for 20/32 lengths)
	require.Regexp(t, `^0x[0-9a-f]{40}$`, out["arr20"].(string))
	require.Regexp(t, `^0x[0-9a-f]{64}$`, out["arr32"].(string))

	// big.Int -> string
	require.Equal(t, "1234567890", out["big"])

	// bool preserved
	require.Equal(t, true, out["bool"])

	// ints preserved as numbers
	require.Equal(t, 42, out["small_int"])
}

func TestPostSignedEvent_HTTPPayloadStructure(t *testing.T) {
	// Provide a runtime with a preloaded secret for the API key.
	// Note: secrets are stored under empty namespace when GetSecret is called without one.
	rt := testutils.NewRuntime(t, testutils.Secrets{
		"": map[testutils.ID]string{
			"courier": "API-KEY",
		},
	})
	// HTTP capability mock to capture POST payload
	httpCap, err := httpmock.NewClientCapability(t)
	require.NoError(t, err)
	httpCap.SendRequest = func(_ context.Context, req *httpcap.Request) (*httpcap.Response, error) {
		// headers
		require.Equal(t, "application/json", req.Headers["Content-Type"])
		require.Equal(t, "API-KEY", req.Headers["Api-Key"])
		require.Equal(t, "http://example.com/system/onchain-watcher-events", req.Url)
		require.Equal(t, "POST", req.Method)

		// JSON body
		var body map[string]any
		require.NoError(t, json.Unmarshal(req.Body, &body))

		// required fields
		require.NotEmpty(t, body["event_id"])
		require.Equal(t, "test", body["domain"])
		require.Equal(t, "Sender", body["name"])
		// Courier protocol expects chain_selector as number
		require.Equal(t, float64(11155111), body["chain_selector"], "chain_selector should be a number")
		require.Equal(t, "0xABCDEF", body["address"])

		// ocr report/context hex encoded
		ocrCtx := body["ocr_context"].(string)
		ocrRpt := body["ocr_report"].(string)
		require.Equal(t, "0x", ocrCtx[:2])
		require.Equal(t, "0x", ocrRpt[:2])
		_, err := hex.DecodeString(ocrCtx[2:])
		require.NoError(t, err)
		_, err = hex.DecodeString(ocrRpt[2:])
		require.NoError(t, err)

		// signatures present
		sigs := body["signatures"].([]any)
		require.GreaterOrEqual(t, len(sigs), 1)
		for _, s := range sigs {
			str := s.(string)
			require.Greater(t, len(str), 2)
			require.Equal(t, "0x", str[:2])
		}

		// verifiable_event matches provided one (checked below)
		require.NotEmpty(t, body["verifiable_event"])
		return &httpcap.Response{StatusCode: 200}, nil
	}

	// Workflow config for POST (use secret id, not inline key)
	cfg := &Config{
		Network:       "evm",
		ChainID:       "1",
		ChainSelector: "11155111", // Provide explicit selector
		CourierURL:    "http://example.com",
		Service:       "test",
		ApiKeySecret:  "courier",
		DetectEventTriggerConfig: DetectEventTriggerConfig{
			ContractName: "TestConsumer",
		},
	}

	// Build a pre-consensus verifiable event
	params := SanitiseJSON(map[string]any{
		"Sender": common.HexToAddress("0x1111111111111111111111111111111111111111").Bytes(),
	}).(map[string]any)
	pre, err := BuildAndHashEventEnvelope(
		"test",
		"Sender",
		"0xABCDEF",
		testABIForCommon,
		"1",
		100,
		2,
		"0xaaaa",
		4321,
		params,
		map[string]any{},
	)
	require.NoError(t, err)
	require.NotEmpty(t, pre.Base64Event)

	// Post to courier
	out, err := PostSignedEvent(cfg, rt, "Sender", "0xABCDEF", pre)
	require.NoError(t, err)
	require.Equal(t, pre.Base64Event, out)

	// sanity: verify the embedded verifiable_event decodes
	decoded, err := base64.StdEncoding.DecodeString(out)
	require.NoError(t, err)
	var obj map[string]any
	require.NoError(t, json.Unmarshal(decoded, &obj))
	ev := obj["event"].(map[string]any)
	require.Equal(t, "test", ev["service"])
	require.Equal(t, "Sender", ev["name"])
}
