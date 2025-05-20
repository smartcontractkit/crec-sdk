package events

import (
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"

	"github.com/smartcontractkit/cvm-sdk/internal/creclient/kafka"
)

type CVMEvent struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Data      string    `json:"data"`
}

type EventListenerOptions struct {
	Brokers []string
	Topic   string
	GroupID string
}

type EventListener struct {
	KafkaReader *kafka.Reader
	Log         zerolog.Logger
}

func NewEventListener(opts *EventListenerOptions) *EventListener {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	logger.Info().Msg("Creating CRE event listener")

	kafkaReader := kafka.NewKafkaEventListener(opts.Brokers, opts.Topic, opts.GroupID)

	return &EventListener{
		KafkaReader: kafkaReader,
		Log:         logger,
	}
}

func (l *EventListener) Read() (*CVMEvent, error) {
	l.Log.Debug().Msg("Reading event from broker")

	msg, err := l.KafkaReader.ReadMessage()
	if err != nil {
		return nil, fmt.Errorf("error while reading message: %w", err)
	}
	l.Log.Debug().Str("data", string(msg.Value)).Msg("Event received")

	return &CVMEvent{
		ID:        "1",
		Timestamp: msg.Time,
		Data:      string(msg.Value),
	}, nil
}

func (l *EventListener) Close() error {
	l.Log.Info().Msg("Closing event listener")
	return l.KafkaReader.EventReader.Close()
}
