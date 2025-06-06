package kafka

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/segmentio/kafka-go"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
	"github.com/smartcontractkit/chainlink-common/pkg/values/pb"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/cvn-sdk/events/types"
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
	beholderEvents, err := t.readBeholderEventsFromKafkaMessage(ctx)

	if len(beholderEvents) == 0 {
		return nil, fmt.Errorf("no beholder events found in the message")
	}

	var verifiableEvent types.VerifiableEvent
	err = json.Unmarshal([]byte(beholderEvents[0].RawData), &verifiableEvent)
	if err != nil {
		return nil, err
	}

	return &verifiableEvent, nil
}

func (t *Reader) readBeholderEventsFromKafkaMessage(ctx context.Context) ([]BeholderEvent, error) {
	var beholderEvents []BeholderEvent
	msg, err := t.EventReader.ReadMessage(ctx)
	if err != nil {
		return nil, fmt.Errorf("error while reading message: %w", err)
	}
	msgs, err := t.parseJsonAndExtractBytesValueBody(msg.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	for _, bodyBytesValue := range msgs {
		decodedBody, err := base64.StdEncoding.DecodeString(bodyBytesValue)
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64, %w", err)
		}
		pbMap := new(pb.Map)
		if err := proto.Unmarshal(decodedBody, pbMap); err != nil {
			return nil, fmt.Errorf("error unmarshalling body into pb.Map: %w", err)
		}
		value, err := values.FromMapValueProto(pbMap)
		if err != nil {
			return nil, fmt.Errorf("error creating values.Value object from pb.Value: %w", err)
		}
		var beholderEvent BeholderEvent
		err = value.UnwrapTo(&beholderEvent)
		if err != nil {
			return nil, fmt.Errorf(
				"error unwrapping values.Value into model.BeholderEvent (possibly not bff event): %w", err,
			)
		}
		beholderEvents = append(beholderEvents, beholderEvent)
	}
	return beholderEvents, nil
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
