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
	ocrReportPayloadOffset = 109
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
		Str("ocr_report", *event.OcrReport).
		Str("ocr_context", *event.OcrContext).
		Msg("Verifying event")

	ocrReport, err := common.ParseHexOrString(*event.OcrReport)
	if err != nil {
		v.logger.Error().Err(err).Msg("Failed to parse report")
		return false, err
	}
	ocrContext, err := common.ParseHexOrString(*event.OcrContext)
	if err != nil {
		v.logger.Error().Err(err).Msg("Failed to parse report context")
		return false, err
	}

	if len(ocrReport) < ocrReportPayloadOffset+32 { // 32 bytes for event hash
		v.logger.Error().Msg("Report is too short")
		return false, err
	}

	// compute the event hash from the event data
	eventHash := v.EventHash(event)

	// ensure locally computed event hash matches the one in the report
	eventHashValid := v.VerifyEventHash(ocrReport, eventHash)
	if !eventHashValid {
		v.logger.Error().Err(err).Msg("Failed to verify event hash")
		return false, err
	}

	// generate the report hash matching the DON signing format
	reportHash := crypto.Keccak256Hash(append(crypto.Keccak256(ocrReport), ocrContext...))

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

		// If we have enough valid signatures, we can stop
		if validSigCount >= v.minRequiredSignatures {
			break
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

func (v *EventVerifier) VerifyEventHash(ocrReport []byte, eventHash common.Hash) bool {

	// OCR report layout:
	// version                offset   0, size  1
	// workflow_execution_id  offset   1, size 32
	// timestamp              offset  33, size  4
	// don_id                 offset  37, size  4
	// don_config_version,    offset  41, size  4
	// workflow_cid           offset  45, size 32
	// workflow_name          offset  77, size 10
	// workflow_owner         offset  87, size 20
	// report_id              offset 107, size  2
	// payload			      offset 109, size  ... <- event hash

	reportEventHash := common.BytesToHash(ocrReport[ocrReportPayloadOffset:])
	v.logger.Debug().
		Str("local_event_hash", eventHash.String()).
		Str("report_event_hash", reportEventHash.String()).
		Msg("Comparing event hashes")

	return eventHash == reportEventHash
}
