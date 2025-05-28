package listener

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"

	"github.com/smartcontractkit/cvm-sdk/events/types"
	"github.com/smartcontractkit/cvm-sdk/internal/creclient/kafka"
)

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

func (l *EventListener) Read() (*types.VerifiableEvent, error) {
	l.Log.Debug().Msg("Reading event from broker")
	verifiableEvent, err := l.KafkaReader.ReadMessage()
	if err != nil {
		return nil, fmt.Errorf("error while reading message: %w", err)
	}
	return verifiableEvent, nil
}

func (l *EventListener) Close() error {
	l.Log.Info().Msg("Closing event listener")
	return l.KafkaReader.EventReader.Close()
}
