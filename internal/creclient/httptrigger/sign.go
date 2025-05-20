package httptrigger

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const ethSignedMessagePrefix = "\x19Ethereum Signed Message:\n"

func flatten(data ...[]byte) []byte {
	var result []byte
	for _, d := range data {
		result = append(result, d...)
	}
	return result
}

func generateEthPrefixedMsgHash(msg []byte) (hash common.Hash) {
	prefixedMsg := fmt.Sprintf("%s%d%s", ethSignedMessagePrefix, len(msg), msg)
	return crypto.Keccak256Hash([]byte(prefixedMsg))
}

func generateEthSignature(privateKey *ecdsa.PrivateKey, msg []byte) (signature []byte, err error) {
	hash := generateEthPrefixedMsgHash(msg)
	return crypto.Sign(hash[:], privateKey)
}

func signData(privateKey *ecdsa.PrivateKey, data ...[]byte) ([]byte, error) {
	return generateEthSignature(privateKey, flatten(data...))
}
