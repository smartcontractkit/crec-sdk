package httptrigger

import (
	"encoding/json"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
)

func TestSign(t *testing.T) {
	privateKey, _ := crypto.HexToECDSA("ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80")
	message := &Message{
		Body: MessageBody{
			MessageId: "web-api-trigger@1.0.0",
			Method:    "web_api_trigger",
			DonId:     "workflow_don_1",
			Payload:   json.RawMessage(`{"key":"value"}`),
		},
	}
	err := message.Sign(privateKey)
	assert.NoError(t, err)
	assert.Equal(
		t,
		"0x0b5f5332eac0bd15ffe087722b38b4593c46c892d19904b56cf492b095470a7b443689ad358dbbda80a39f5593daced0a8f05d2c9b0de768d9b8c2f4053c379b01",
		message.Signature,
	)
	assert.Equal(t, "0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266", message.Body.Sender)
}
