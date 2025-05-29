package verifier

import (
	"encoding/base64"
	"encoding/json"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rs/zerolog"

	"github.com/smartcontractkit/cvm-sdk/events/types"
)

type EventVerifierOptions struct {
	ValidSigners          []string
	MinRequiredSignatures int
}

type EventVerifier struct {
	Log                   zerolog.Logger
	ValidSigners          []string
	MinRequiredSignatures int
}

func NewEventVerifier(opts *EventVerifierOptions) *EventVerifier {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	logger.Info().Msg("Creating CVN event verifier")

	return &EventVerifier{
		Log:                   logger,
		ValidSigners:          opts.ValidSigners,
		MinRequiredSignatures: opts.MinRequiredSignatures,
	}
}

func (v *EventVerifier) Verify(verifiableEvent *types.VerifiableEvent) (bool, error) {
	v.Log.Debug().Str("eventType", verifiableEvent.Type).Msg("Verifying verifiableEvent")

	hash := v.EventHash(verifiableEvent)

	v.Log.Debug().Str("hash", hash.Hex()).Msg("Verifying event hash")

	validSigCount := 0
	availableSigners := make(map[common.Address]bool)

	for _, signer := range v.ValidSigners {
		availableSigners[common.HexToAddress(signer)] = true
	}

	for _, sig := range verifiableEvent.Signatures {
		sigBytes, err := common.ParseHexOrString(sig)
		if err != nil {
			v.Log.Error().Err(err).Msg("Failed to parse signature")
			return false, err
		}
		if sigBytes[64] == 27 || sigBytes[64] == 28 {
			sigBytes[64] -= 27 // Adjust signature format for Ethereum signatures
		}
		pubKey, err := crypto.SigToPub(hash.Bytes(), sigBytes)
		if err != nil {
			v.Log.Error().Err(err).Msg("Failed to recover public key from signature")
			return false, err
		}

		signer := crypto.PubkeyToAddress(*pubKey)
		v.Log.Debug().Str("signer", signer.Hex()).Msg("Recovered signer from signature")

		if availableSigners[signer] {
			v.Log.Info().Str("signer", signer.Hex()).Msg("Signature verified successfully")
			validSigCount++
			availableSigners[signer] = false // Mark this signer as used
		}

		// If we have enough valid signatures, we can return early
		if validSigCount >= v.MinRequiredSignatures {
			return true, nil
		}
	}
	v.Log.Debug().Int("validSigCount", validSigCount).Msg("Total valid signatures")

	return validSigCount >= v.MinRequiredSignatures, nil
}

func (v *EventVerifier) Decode(verifiableEvent *types.VerifiableEvent, event any) error {
	decodedStr, err := base64.StdEncoding.DecodeString(verifiableEvent.Payload)
	if err != nil {
		v.Log.Error().Err(err).Msg("Failed to decode base64 payload")
		return err
	}

	return json.Unmarshal(decodedStr, event)
}

func (v *EventVerifier) EventHash(verifiableEvent *types.VerifiableEvent) common.Hash {
	return crypto.Keccak256Hash([]byte(verifiableEvent.Type + verifiableEvent.Payload))
}
