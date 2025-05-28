package verifier

import (
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

func (v *EventVerifier) Verify(event *types.VerifiableEvent) (bool, error) {
	v.Log.Debug().Str("eventType", event.Type).Msg("Verifying event")

	hash := crypto.Keccak256Hash([]byte(event.Payload))

	validSigCount := 0
	availableSigners := make(map[common.Address]bool)

	for _, signer := range v.ValidSigners {
		availableSigners[common.HexToAddress(signer)] = true
	}

	for _, sig := range event.Signatures {
		pubKey, err := crypto.SigToPub(hash.Bytes(), common.Hex2Bytes(sig))
		if err != nil {
			v.Log.Error().Err(err).Msg("Failed to recover public key from signature")
			return false, err
		}

		signer := crypto.PubkeyToAddress(*pubKey)

		if availableSigners[signer] {
			v.Log.Debug().Str("signer", signer.Hex()).Msg("Signature verified successfully")
			validSigCount++
			availableSigners[signer] = false // Mark this signer as used
		}
	}
	v.Log.Debug().Int("validSigCount", validSigCount).Msg("Total valid signatures")

	return validSigCount >= v.MinRequiredSignatures, nil
}
