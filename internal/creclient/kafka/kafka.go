package kafka

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"

	"github.com/segmentio/kafka-go"

	"github.com/smartcontractkit/cvm-sdk/events/types"
)

type Reader struct {
	EventReader *kafka.Reader
}

func NewKafkaEventListener(brokers []string, topic string, groupId string) *Reader {
	return &Reader{
		EventReader: kafka.NewReader(
			kafka.ReaderConfig{
				Brokers:  brokers,
				Topic:    topic,
				GroupID:  groupId,
				MinBytes: 1,    // do not block for small messages
				MaxBytes: 10e6, // 10MB
			},
		),
	}
}

func (t *Reader) Close() error {
	return t.EventReader.Close()
}

func (t *Reader) ReadMessage(ctx context.Context) (*types.VerifiableEvent, error) {
	msg, err := t.EventReader.ReadMessage(ctx)
	if err != nil {
		return nil, err
	}
	msgs, err := t.parseJsonAndExtractBytesValueBody(msg.Value)
	if err != nil {
		return nil, err
	}
	if len(msgs) == 0 {
		return nil, errors.New("cannot process empty message body")
	}

	var verifiableEvent types.VerifiableEvent

	decoded, err := base64.StdEncoding.DecodeString(msgs[0])
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(decoded, &verifiableEvent)
	if err != nil {
		return nil, err
	}

	return &verifiableEvent, nil
}

func (t *Reader) parseJsonAndExtractBytesValueBody(jsonData []byte) ([]string, error) {
	var body Body
	if err := json.Unmarshal(jsonData, &body); err != nil {
		return nil, err
	}
	msgs := make([]string, 0)
	resourceLogs := body.ResourceLogs
	for _, resourceLog := range resourceLogs {
		for _, scopeLog := range resourceLog.ScopeLogs {
			for _, record := range scopeLog.LogRecords {
				msgs = append(msgs, record.Body.BytesValue)
			}
		}
	}
	return msgs, nil
}
