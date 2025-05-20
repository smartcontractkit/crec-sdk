package kafka

import (
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"
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
				MinBytes: 1,    // 10KB
				MaxBytes: 10e6, // 10MB
			},
		),
	}
}

func (t *Reader) Close() error {
	return t.EventReader.Close()
}

func (t *Reader) ReadMessage() (*kafka.Message, error) {
	msg, err := t.EventReader.ReadMessage(context.Background())
	if err != nil {
		return nil, err
	}
	return &msg, nil
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

// func (t *KafkaReader) ReadBeholderEventsFromKafkaMessage() ([]CREEvent, error) {
// 	var beholderEvents []CREEvent
// 	msg, err := t.EventReader.ReadMessage(context.Background())
// 	if err != nil {
// 		return nil, fmt.Errorf("error while reading message: %w", err)
// 	}
// 	msgs, err := t.parseJsonAndExtractBytesValueBody(msg.Value)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to parse JSON: %w", err)
// 	}
//
// 	for _, bodyBytesValue := range msgs {
// 		decodedBody, err := base64.StdEncoding.DecodeString(bodyBytesValue)
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to decode base64, %w", err)
// 		}
// 		pbMap := new(pb.Map)
// 		if err := proto.Unmarshal(decodedBody, pbMap); err != nil {
// 			return nil, fmt.Errorf("error unmarshalling body into pb.Map: %w", err)
// 		}
// 		value, err := values.FromMapValueProto(pbMap)
// 		if err != nil {
// 			return nil, fmt.Errorf("error creating values.Value object from pb.Value: %w", err)
// 		}
// 		var creEvent CREEvent
// 		err = value.UnwrapTo(&creEvent)
// 		if err != nil {
// 			return nil, fmt.Errorf(
// 				"error unwrapping values.Value into model.BeholderEvent (possibly not bff event): %w", err,
// 			)
// 		}
// 		beholderEvents = append(beholderEvents, creEvent)
// 	}
// 	return beholderEvents, nil
// }
