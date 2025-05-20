package httptrigger

import (
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	MessageIdMaxLen     = 128
	MessageMethodMaxLen = 64
	MessageDonIdMaxLen  = 64
	MessageReceiverLen  = 2 + 2*20
)

type Message struct {
	Signature string      `json:"signature"`
	Body      MessageBody `json:"body"`
}

type MessageBody struct {
	MessageId string          `json:"message_id"`
	Method    string          `json:"method"`
	DonId     string          `json:"don_id"`
	Receiver  string          `json:"receiver"`
	Payload   json.RawMessage `json:"payload,omitempty"`
	Sender    string          `json:"-"`
}

func (m *Message) Sign(privateKey *ecdsa.PrivateKey) error {
	if m == nil {
		return errors.New("nil message")
	}
	rawData := getRawMessageBody(&m.Body)
	signature, err := signData(privateKey, rawData...)
	if err != nil {
		return err
	}
	m.Signature = "0x" + common.Bytes2Hex(signature)
	m.Body.Sender = strings.ToLower(crypto.PubkeyToAddress(privateKey.PublicKey).Hex())
	return nil
}

func getRawMessageBody(msgBody *MessageBody) [][]byte {
	alignedMessageId := make([]byte, MessageIdMaxLen)
	copy(alignedMessageId, msgBody.MessageId)
	alignedMethod := make([]byte, MessageMethodMaxLen)
	copy(alignedMethod, msgBody.Method)
	alignedDonId := make([]byte, MessageDonIdMaxLen)
	copy(alignedDonId, msgBody.DonId)
	alignedReceiver := make([]byte, MessageReceiverLen)
	copy(alignedReceiver, msgBody.Receiver)
	return [][]byte{alignedMessageId, alignedMethod, alignedDonId, alignedReceiver, msgBody.Payload}
}
