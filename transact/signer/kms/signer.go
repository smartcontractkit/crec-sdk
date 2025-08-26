// Portions of this file are derived from github.com/matelang/go-ethereum-aws-kms-tx-signer
// Copyright (c) 2023 matelang, licensed under MIT License.
// See THIRD_PARTY_LICENSES.md for full license text.

package kms

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/asn1"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/pkg/errors"
	"github.com/smartcontractkit/cvn-sdk/transact/signer"
)

var _ signer.Signer = &Signer{}

// KMSClient is a narrow interface for AWS KMS operations needed by the signer
//
//go:generate mockery --name=KMSClient --output=mocks --outpkg=mocks
type KMSClient interface {
	GetPublicKey(ctx context.Context, params *kms.GetPublicKeyInput, optFns ...func(*kms.Options)) (*kms.GetPublicKeyOutput, error)
	Sign(ctx context.Context, params *kms.SignInput, optFns ...func(*kms.Options)) (*kms.SignOutput, error)
}

// Signer implements ECDSA signing using AWS KMS
type Signer struct {
	client KMSClient
	keyID  string
}

// NewSigner creates a new KMS signer with the specified key ID
// The keyID should be the ARN, key ID, or alias of the KMS key
// Loads AWS configuration from the AWS CLI environment (see https://docs.aws.amazon.com/cli/v1/userguide/cli-configure-envvars.html)
func NewSigner(ctx context.Context, keyID string) (*Signer, error) {
	if keyID == "" {
		return nil, fmt.Errorf("keyID cannot be empty")
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &Signer{
		client: kms.NewFromConfig(cfg),
		keyID:  keyID,
	}, nil
}

// NewSignerWithConfig creates a new KMS signer with custom AWS configuration
func NewSignerWithConfig(cfg aws.Config, keyID string) (*Signer, error) {
	if keyID == "" {
		return nil, fmt.Errorf("keyID cannot be empty")
	}

	return &Signer{
		client: kms.NewFromConfig(cfg),
		keyID:  keyID,
	}, nil
}

// NewSignerWithClient creates a new KMS signer with a custom KMS client (useful for testing)
func NewSignerWithClient(client KMSClient, keyID string) (*Signer, error) {
	if keyID == "" {
		return nil, fmt.Errorf("keyID cannot be empty")
	}

	return &Signer{
		client: client,
		keyID:  keyID,
	}, nil
}

func (s *Signer) Sign(ctx context.Context, hash []byte) ([]byte, error) {
	pubkey, err := GetPubKeyCtx(ctx, s.client, s.keyID)
	if err != nil {
		return nil, err
	}

	pubKeyBytes := secp256k1.S256().Marshal(pubkey.X, pubkey.Y)

	rBytes, sBytes, err := getSignatureFromKms(ctx, s.client, s.keyID, hash)
	if err != nil {
		return nil, err
	}

	// Adjust S value from signature according to Ethereum standard
	sBigInt := new(big.Int).SetBytes(sBytes)
	if sBigInt.Cmp(secp256k1HalfN) > 0 {
		sBytes = new(big.Int).Sub(secp256k1N, sBigInt).Bytes()
	}

	signature, err := getEthereumSignature(pubKeyBytes, hash, rBytes, sBytes)
	if err != nil {
		return nil, err
	}

	return signature, nil
}

// ------------------------------
// The below code has been ported over from https://github.com/matelang/go-ethereum-aws-kms-tx-signer
// Licensed under MIT License (see header above)
// ------------------------------
const awsKmsSignOperationMessageType = "DIGEST"
const awsKmsSignOperationSigningAlgorithm = "ECDSA_SHA_256"

var secp256k1N = crypto.S256().Params().N
var secp256k1HalfN = new(big.Int).Div(secp256k1N, big.NewInt(2))

type asn1EcPublicKey struct {
	EcPublicKeyInfo asn1EcPublicKeyInfo
	PublicKey       asn1.BitString
}

type asn1EcPublicKeyInfo struct {
	Algorithm  asn1.ObjectIdentifier
	Parameters asn1.ObjectIdentifier
}

type asn1EcSig struct {
	R asn1.RawValue
	S asn1.RawValue
}

func getPublicKeyDerBytesFromKMS(ctx context.Context, svc KMSClient, keyId string) ([]byte, error) {
	getPubKeyOutput, err := svc.GetPublicKey(ctx, &kms.GetPublicKeyInput{
		KeyId: aws.String(keyId),
	})
	if err != nil {
		return nil, errors.Wrapf(err, "can not get public key from KMS for KeyId=%s", keyId)
	}

	var asn1pubk asn1EcPublicKey
	_, err = asn1.Unmarshal(getPubKeyOutput.PublicKey, &asn1pubk)
	if err != nil {
		return nil, errors.Wrapf(err, "can not parse asn1 public key for KeyId=%s", keyId)
	}

	return asn1pubk.PublicKey.Bytes, nil
}

func getSignatureFromKms(
	ctx context.Context, svc KMSClient, keyId string, txHashBytes []byte,
) ([]byte, []byte, error) {
	signInput := &kms.SignInput{
		KeyId:            aws.String(keyId),
		SigningAlgorithm: awsKmsSignOperationSigningAlgorithm,
		MessageType:      awsKmsSignOperationMessageType,
		Message:          txHashBytes,
	}

	signOutput, err := svc.Sign(ctx, signInput)
	if err != nil {
		return nil, nil, err
	}

	var sigAsn1 asn1EcSig
	_, err = asn1.Unmarshal(signOutput.Signature, &sigAsn1)
	if err != nil {
		return nil, nil, err
	}

	return sigAsn1.R.Bytes, sigAsn1.S.Bytes, nil
}

func getEthereumSignature(expectedPublicKeyBytes []byte, txHash []byte, r []byte, s []byte) ([]byte, error) {
	rsSignature := append(adjustSignatureLength(r), adjustSignatureLength(s)...)
	signature := append(rsSignature, []byte{0}...)

	recoveredPublicKeyBytes, err := crypto.Ecrecover(txHash, signature)
	if err != nil {
		return nil, err
	}

	if hex.EncodeToString(recoveredPublicKeyBytes) != hex.EncodeToString(expectedPublicKeyBytes) {
		signature = append(rsSignature, []byte{1}...)
		recoveredPublicKeyBytes, err = crypto.Ecrecover(txHash, signature)
		if err != nil {
			return nil, err
		}

		if hex.EncodeToString(recoveredPublicKeyBytes) != hex.EncodeToString(expectedPublicKeyBytes) {
			return nil, errors.New("can not reconstruct public key from sig")
		}
	}

	return signature, nil
}

func GetPubKeyCtx(ctx context.Context, svc KMSClient, keyId string) (*ecdsa.PublicKey, error) {
	pubKeyBytes, err := getPublicKeyDerBytesFromKMS(ctx, svc, keyId)
	if err != nil {
		return nil, err
	}

	pubkey, err := crypto.UnmarshalPubkey(pubKeyBytes)
	if err != nil {
		return nil, errors.Wrap(err, "can not construct secp256k1 public key from key bytes")
	}
	return pubkey, nil
}

func adjustSignatureLength(buffer []byte) []byte {
	buffer = bytes.TrimLeft(buffer, "\x00")
	for len(buffer) < 32 {
		zeroBuf := []byte{0}
		buffer = append(zeroBuf, buffer...)
	}
	return buffer
}
