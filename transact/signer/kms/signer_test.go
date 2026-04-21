package kms

import (
	"crypto/ecdsa"
	"encoding/asn1"
	"math/big"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/smartcontractkit/crec-sdk/transact/signer/kms/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testKeyID = "test-key-id"

// createTestPublicKeyDER is a helper function to create a valid secp256k1 public key in DER format
func createTestPublicKeyDER(t *testing.T) []byte {
	privateKey, _ := crypto.GenerateKey()
	pubKey := &privateKey.PublicKey
	return createPublicKeyDER(t, pubKey)
}

// createTestDERSignature is a helper function to create a valid DER signature
func createTestDERSignature(r, s []byte) []byte {
	sig := asn1EcSig{
		R: asn1.RawValue{Tag: 2, Bytes: r},
		S: asn1.RawValue{Tag: 2, Bytes: s},
	}
	derBytes, _ := asn1.Marshal(sig)
	return derBytes
}

// createPublicKeyDER creates a DER-encoded public key from an existing ecdsa.PublicKey
func createPublicKeyDER(t *testing.T, pubKey *ecdsa.PublicKey) []byte {
	pubKeyBytes := crypto.FromECDSAPub(pubKey)

	asn1PubKey := asn1EcPublicKey{
		EcPublicKeyInfo: asn1EcPublicKeyInfo{
			Algorithm:  asn1.ObjectIdentifier{1, 2, 840, 10045, 2, 1}, // ecPublicKey
			Parameters: asn1.ObjectIdentifier{1, 3, 132, 0, 10},       // secp256k1
		},
		PublicKey: asn1.BitString{
			Bytes:     pubKeyBytes,
			BitLength: len(pubKeyBytes) * 8,
		},
	}

	derBytes, err := asn1.Marshal(asn1PubKey)
	require.NoError(t, err)
	return derBytes
}

func mockGetPublicKeySuccess(t *testing.T, mockClient *mocks.KMSClient, keyID string, pubKeyDER []byte) {
	mockClient.On("GetPublicKey", t.Context(), &kms.GetPublicKeyInput{
		KeyId: aws.String(keyID),
	}).Return(&kms.GetPublicKeyOutput{
		PublicKey: pubKeyDER,
	}, nil)
}

func mockGetPublicKeyError(t *testing.T, mockClient *mocks.KMSClient, keyID string) {
	mockClient.On("GetPublicKey", t.Context(), &kms.GetPublicKeyInput{
		KeyId: aws.String(keyID),
	}).Return((*kms.GetPublicKeyOutput)(nil), assert.AnError)
}

func mockSignSuccess(t *testing.T, mockClient *mocks.KMSClient, keyID string, hash []byte, signature []byte) {
	mockClient.On("Sign", t.Context(), &kms.SignInput{
		KeyId:            aws.String(keyID),
		SigningAlgorithm: awsKmsSignOperationSigningAlgorithm,
		MessageType:      awsKmsSignOperationMessageType,
		Message:          hash,
	}).Return(&kms.SignOutput{
		Signature: signature,
	}, nil)
}

func mockSignError(t *testing.T, mockClient *mocks.KMSClient, keyID string, hash []byte) {
	mockClient.On("Sign", t.Context(), &kms.SignInput{
		KeyId:            aws.String(keyID),
		SigningAlgorithm: awsKmsSignOperationSigningAlgorithm,
		MessageType:      awsKmsSignOperationMessageType,
		Message:          hash,
	}).Return((*kms.SignOutput)(nil), assert.AnError)
}

func TestNewSigner(t *testing.T) {
	tests := []struct {
		name    string
		keyID   string
		wantErr bool
	}{
		{
			name:    "empty key ID should fail",
			keyID:   "",
			wantErr: true,
		},
		{
			name:    "valid key ID should create signer",
			keyID:   "arn:aws:kms:us-west-2:123456789012:key/12345678-1234-1234-1234-123456789012",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signer, err := NewSigner(t.Context(), tt.keyID)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, signer)
			} else {
				// Note: This will fail without valid AWS credentials, but tests the constructor logic
				if err != nil {
					t.Skip("Skipping test that requires AWS credentials")
				}
				assert.NotNil(t, signer)
				assert.Equal(t, tt.keyID, signer.keyID)
			}
		})
	}
}

func TestNewSignerWithConfig(t *testing.T) {
	cfg := aws.Config{
		Region: "us-west-2",
	}
	keyID := testKeyID

	signer, err := NewSignerWithConfig(cfg, keyID)
	require.NoError(t, err)
	assert.NotNil(t, signer)
	assert.Equal(t, keyID, signer.keyID)
}

func TestNewSignerWithConfig_EmptyKeyID(t *testing.T) {
	cfg := aws.Config{
		Region: "us-west-2",
	}

	signer, err := NewSignerWithConfig(cfg, "")
	assert.Error(t, err)
	assert.Nil(t, signer)
	assert.Contains(t, err.Error(), "keyID cannot be empty")
}

func TestAdjustSignatureLength(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected []byte
	}{
		{
			name:     "already 32 bytes",
			input:    make([]byte, 32),
			expected: make([]byte, 32),
		},
		{
			name:     "shorter than 32 bytes",
			input:    []byte{0x01, 0x02, 0x03},
			expected: append(make([]byte, 29), []byte{0x01, 0x02, 0x03}...),
		},
		{
			name:     "with leading zeros",
			input:    []byte{0x00, 0x00, 0x01, 0x02},
			expected: append(make([]byte, 30), []byte{0x01, 0x02}...),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := adjustSignatureLength(tt.input)
			assert.Equal(t, 32, len(result))
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewSignerWithClient(t *testing.T) {
	mockClient := mocks.NewKMSClient(t)
	keyID := testKeyID

	signer, err := NewSignerWithClient(mockClient, keyID)
	require.NoError(t, err)
	assert.NotNil(t, signer)
	assert.Equal(t, keyID, signer.keyID)
	assert.Equal(t, mockClient, signer.client)
}

func TestNewSignerWithClient_EmptyKeyID(t *testing.T) {
	mockClient := mocks.NewKMSClient(t)

	signer, err := NewSignerWithClient(mockClient, "")
	assert.Error(t, err)
	assert.Nil(t, signer)
	assert.Contains(t, err.Error(), "keyID cannot be empty")
}

func TestNewSignerWithConfig_WithClientOption(t *testing.T) {
	mockClient := mocks.NewKMSClient(t)
	cfg := aws.Config{
		Region: "us-west-2",
	}

	signer, err := NewSignerWithConfig(cfg, testKeyID, WithClient(mockClient))
	require.NoError(t, err)
	assert.NotNil(t, signer)
	assert.Equal(t, mockClient, signer.client)
}

func TestSigner_Sign_Success(t *testing.T) {
	privateKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	hash := crypto.Keccak256([]byte("test message"))

	realSig, err := crypto.Sign(hash, privateKey)
	require.NoError(t, err)

	r := realSig[:32]
	s := realSig[32:64]

	pubKeyDER := createPublicKeyDER(t, &privateKey.PublicKey)
	derSig := createTestDERSignature(r, s)

	mockClient := mocks.NewKMSClient(t)
	mockGetPublicKeySuccess(t, mockClient, testKeyID, pubKeyDER)
	mockSignSuccess(t, mockClient, testKeyID, hash, derSig)

	signer, err := NewSignerWithClient(mockClient, testKeyID)
	require.NoError(t, err)

	signature, err := signer.Sign(t.Context(), hash)
	require.NoError(t, err)
	assert.Len(t, signature, 65) // Ethereum signature is 65 bytes (r + s + v)
}

func TestSigner_Sign_GetPublicKeyError(t *testing.T) {
	hash := make([]byte, 32)

	mockClient := mocks.NewKMSClient(t)
	mockGetPublicKeyError(t, mockClient, testKeyID)

	signer, err := NewSignerWithClient(mockClient, testKeyID)
	require.NoError(t, err)

	signature, err := signer.Sign(t.Context(), hash)
	assert.Error(t, err)
	assert.Nil(t, signature)
}

func TestSigner_Sign_SignError(t *testing.T) {
	hash := make([]byte, 32)

	pubKeyDER := createTestPublicKeyDER(t)

	mockClient := mocks.NewKMSClient(t)
	mockGetPublicKeySuccess(t, mockClient, testKeyID, pubKeyDER)
	mockSignError(t, mockClient, testKeyID, hash)

	signer, err := NewSignerWithClient(mockClient, testKeyID)
	require.NoError(t, err)

	signature, err := signer.Sign(t.Context(), hash)
	assert.Error(t, err)
	assert.Nil(t, signature)
}

func TestSigner_Sign_InvalidDERSignature(t *testing.T) {
	hash := make([]byte, 32)
	pubKeyDER := createTestPublicKeyDER(t)

	mockClient := mocks.NewKMSClient(t)
	mockGetPublicKeySuccess(t, mockClient, testKeyID, pubKeyDER)
	mockSignSuccess(t, mockClient, testKeyID, hash, []byte("invalid der signature"))

	signer, err := NewSignerWithClient(mockClient, testKeyID)
	require.NoError(t, err)

	signature, err := signer.Sign(t.Context(), hash)
	assert.Error(t, err)
	assert.Nil(t, signature)
}

func TestSigner_Sign_InvalidPublicKey(t *testing.T) {
	hash := make([]byte, 32)

	mockClient := mocks.NewKMSClient(t)
	mockGetPublicKeySuccess(t, mockClient, testKeyID, []byte("invalid der public key"))

	signer, err := NewSignerWithClient(mockClient, testKeyID)
	require.NoError(t, err)

	signature, err := signer.Sign(t.Context(), hash)
	assert.Error(t, err)
	assert.Nil(t, signature)
}

func TestGetPubKeyCtx_Success(t *testing.T) {
	pubKeyDER := createTestPublicKeyDER(t)

	mockClient := mocks.NewKMSClient(t)
	mockGetPublicKeySuccess(t, mockClient, testKeyID, pubKeyDER)

	pubKey, err := GetPubKeyCtx(t.Context(), mockClient, testKeyID)
	assert.NoError(t, err)
	assert.NotNil(t, pubKey)
	assert.IsType(t, &ecdsa.PublicKey{}, pubKey)
}

func TestGetPubKeyCtx_GetPublicKeyError(t *testing.T) {
	mockClient := mocks.NewKMSClient(t)
	mockGetPublicKeyError(t, mockClient, testKeyID)

	pubKey, err := GetPubKeyCtx(t.Context(), mockClient, testKeyID)
	assert.Error(t, err)
	assert.Nil(t, pubKey)
}

func TestGetSignatureFromKms_Success(t *testing.T) {
	r := make([]byte, 32)
	s := make([]byte, 32)
	r[31] = 1 // Set last byte to 1
	s[31] = 2 // Set last byte to 2

	derSig := createTestDERSignature(r, s)

	mockClient := mocks.NewKMSClient(t)
	hash := make([]byte, 32)
	mockSignSuccess(t, mockClient, testKeyID, hash, derSig)
	rBytes, sBytes, err := getSignatureFromKms(t.Context(), mockClient, testKeyID, hash)
	assert.NoError(t, err)
	assert.Equal(t, r, rBytes)
	assert.Equal(t, s, sBytes)
}

func TestGetSignatureFromKms_SignError(t *testing.T) {
	mockClient := mocks.NewKMSClient(t)
	hash := make([]byte, 32)
	mockSignError(t, mockClient, testKeyID, hash)

	rBytes, sBytes, err := getSignatureFromKms(t.Context(), mockClient, testKeyID, hash)
	assert.Error(t, err)
	assert.Nil(t, rBytes)
	assert.Nil(t, sBytes)
}

func TestGetSignatureFromKms_InvalidRS(t *testing.T) {
	mockClient := mocks.NewKMSClient(t)
	hash := make([]byte, 32)

	secp256k1N, _ := new(big.Int).SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141", 16)

	tests := []struct {
		name string
		r    *big.Int
		s    *big.Int
	}{
		{"r is zero", big.NewInt(0), big.NewInt(1)},
		{"s is zero", big.NewInt(1), big.NewInt(0)},
		{"r is N", secp256k1N, big.NewInt(1)},
		{"s is N", big.NewInt(1), secp256k1N},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sigAsn1, err := asn1.Marshal(struct{ R, S *big.Int }{tt.r, tt.s})
			require.NoError(t, err)

			mockSignSuccess(t, mockClient, testKeyID, hash, sigAsn1)

			rBytes, sBytes, err := getSignatureFromKms(t.Context(), mockClient, testKeyID, hash)
			assert.Error(t, err)
			assert.Nil(t, rBytes)
			assert.Nil(t, sBytes)
			assert.Contains(t, err.Error(), "value out of range")
		})
	}
}

func TestGetSignatureFromKms_TrailingGarbage(t *testing.T) {
	mockClient := mocks.NewKMSClient(t)
	hash := make([]byte, 32)
	
	// Create a valid signature
	sigAsn1, err := asn1.Marshal(struct{ R, S *big.Int }{big.NewInt(1), big.NewInt(1)})
	require.NoError(t, err)
	
	// Append garbage bytes
	sigAsn1 = append(sigAsn1, []byte("garbage")...)
	
	mockSignSuccess(t, mockClient, testKeyID, hash, sigAsn1)
	
	rBytes, sBytes, err := getSignatureFromKms(t.Context(), mockClient, testKeyID, hash)
	assert.Error(t, err)
	assert.Nil(t, rBytes)
	assert.Nil(t, sBytes)
	assert.Contains(t, err.Error(), "trailing garbage bytes")
}
