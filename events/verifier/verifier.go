package verifier

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rs/zerolog"

	"github.com/smartcontractkit/cvn-sdk/client"
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
	logger                *zerolog.Logger
	minRequiredSignatures int
	validSigners          []string
}

func NewEventVerifier(opts *EventVerifierOptions) (*EventVerifier, error) {
	if opts == nil {
		return nil, errors.New("EventVerifier requires a valid options struct")
	}
	if len(opts.ValidSigners) < opts.MinRequiredSignatures {
		return nil, errors.New("EventVerifier requires at least as many valid signer as the minimum required signatures")
	}

	logger := opts.Logger
	if logger == nil {
		lgr := zerolog.New(os.Stdout).With().Timestamp().Logger()
		logger = &lgr
	}

	logger.Info().Msg("Creating CVN event verifier")

	validSigners := make([]string, len(opts.ValidSigners))
	copy(validSigners, opts.ValidSigners)

	return &EventVerifier{
		logger:                logger,
		minRequiredSignatures: opts.MinRequiredSignatures,
		validSigners:          validSigners,
	}, nil
}

func (v *EventVerifier) Verify(event *client.Event) (bool, error) {
	v.logger.Debug().
		Str("event_service", *event.Service).
		Str("event_name", *event.Name).
		Msg("Verifying event")

	eventHash := v.EventHash(event)

	report, err := common.ParseHexOrString(*event.OcrReport)
	if err != nil {
		v.logger.Error().Err(err).Msg("Failed to parse report")
		return false, err
	}
	reportContext, err := common.ParseHexOrString(*event.OcrContext)
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
		Str("event_hash", eventHash.String()).
		Str("report_hash", reportHash.String()).
		Msg("Verifying report hash")

	validSigCount := 0
	availableSigners := make(map[common.Address]bool)

	for _, signer := range v.validSigners {
		availableSigners[common.HexToAddress(signer)] = true
	}

	for _, sig := range *event.Signatures {
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
		if validSigCount >= v.minRequiredSignatures {
			return true, nil
		}
	}
	v.logger.Debug().Int("valid_signatures", validSigCount).Msg("Finished signature checking")

	return validSigCount >= v.minRequiredSignatures, nil
}

func (v *EventVerifier) Decode(event *client.Event, payload any) error {
	jsonBytes, err := v.ToJson(event)
	if err != nil {
		v.logger.Error().Err(err).Msg("Failed to convert verifiable event to JSON")
		return err
	}
	return json.Unmarshal(jsonBytes, payload)
}

func (v *EventVerifier) ToJson(event *client.Event) ([]byte, error) {
	decodedStr, err := base64.StdEncoding.DecodeString(*event.VerifiableEvent)
	if err != nil {
		v.logger.Error().Err(err).Msg("Failed to decode base64 payload")
		return []byte{}, err
	}
	return decodedStr, nil
}

func (v *EventVerifier) EventHash(event *client.Event) common.Hash {
	return crypto.Keccak256Hash([]byte(*event.Service + "." + *event.Name + "." + *event.VerifiableEvent))
}
