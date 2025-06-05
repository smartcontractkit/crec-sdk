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

const (
	ocrReportMetadataLen = 109
)

type EventVerifierOptions struct {
	ValidSigners          []string
	MinRequiredSignatures int
	Logger                *zerolog.Logger
}

type EventVerifier struct {
	logger  *zerolog.Logger
	options *EventVerifierOptions
}

func NewEventVerifier(opts *EventVerifierOptions) *EventVerifier {
	logger := opts.Logger
	if logger == nil {
		lgr := zerolog.New(os.Stdout).With().Timestamp().Logger()
		logger = &lgr
	}

	logger.Info().Msg("Creating CVN event verifier")

	// TODO: Validate options

	return &EventVerifier{
		logger:  logger,
		options: opts,
	}
}

func (v *EventVerifier) Verify(verifiableEvent *types.VerifiableEvent) (bool, error) {
	v.logger.Debug().Str("eventType", verifiableEvent.Type).Msg("Verifying verifiableEvent")

	eventHash := v.EventHash(verifiableEvent)

	report, err := common.ParseHexOrString(verifiableEvent.Report)
	if err != nil {
		v.logger.Error().Err(err).Msg("Failed to parse report")
		return false, err
	}
	reportContext, err := common.ParseHexOrString(verifiableEvent.Context)
	if err != nil {
		v.logger.Error().Err(err).Msg("Failed to parse report context")
		return false, err
	}

	if len(report) < ocrReportMetadataLen {
		v.logger.Error().Msg("Report is too short")
		return false, nil
	}

	// build new report with the locally calculated event hash
	verifiableReport := append(report[:ocrReportMetadataLen], eventHash.Bytes()...)

	// generate the report hash matching the DON signing format
	reportHash := crypto.Keccak256Hash(append(crypto.Keccak256(verifiableReport), reportContext...))

	v.logger.Debug().
		Str("eventHash", eventHash.String()).
		Str("reportHash", reportHash.String()).
		Msg("Verifying report hash")

	validSigCount := 0
	availableSigners := make(map[common.Address]bool)

	for _, signer := range v.options.ValidSigners {
		availableSigners[common.HexToAddress(signer)] = true
	}

	for _, sig := range verifiableEvent.Signatures {
		sigBytes, err := common.ParseHexOrString(sig)
		if err != nil {
			v.logger.Error().Err(err).Msg("Failed to parse signature")
			return false, err
		}
		if sigBytes[64] == 27 || sigBytes[64] == 28 {
			sigBytes[64] -= 27 // Adjust signature for Ethereum signatures
		}
		pubKey, err := crypto.SigToPub(reportHash.Bytes(), sigBytes)
		if err != nil {
			v.logger.Error().Err(err).Msg("Failed to recover public key from signature")
			return false, err
		}

		signer := crypto.PubkeyToAddress(*pubKey)
		v.logger.Debug().
			Str("signer", signer.Hex()).
			Str("signature", sig).
			Msg("Recovered signer from signature")

		if availableSigners[signer] {
			v.logger.Info().Str("signer", signer.Hex()).Msg("Signature verified successfully")
			validSigCount++
			availableSigners[signer] = false // Mark this signer as used
		}

		// If we have enough valid signatures, we can return early
		if validSigCount >= v.options.MinRequiredSignatures {
			return true, nil
		}
	}
	v.logger.Debug().Int("validSigCount", validSigCount).Msg("Total valid signatures")

	return validSigCount >= v.options.MinRequiredSignatures, nil
}

func (v *EventVerifier) Decode(verifiableEvent *types.VerifiableEvent, event any) error {
	jsonBytes, err := v.ToJson(verifiableEvent)
	if err != nil {
		v.logger.Error().Err(err).Msg("Failed to convert verifiable event to JSON")
		return err
	}
	return json.Unmarshal(jsonBytes, event)
}

func (v *EventVerifier) ToJson(verifiableEvent *types.VerifiableEvent) ([]byte, error) {
	decodedStr, err := base64.StdEncoding.DecodeString(verifiableEvent.Payload)
	if err != nil {
		v.logger.Error().Err(err).Msg("Failed to decode base64 payload")
		return []byte{}, err
	}
	return decodedStr, nil
}

func (v *EventVerifier) EventHash(verifiableEvent *types.VerifiableEvent) common.Hash {
	return crypto.Keccak256Hash([]byte(verifiableEvent.Type + verifiableEvent.Payload))
}
