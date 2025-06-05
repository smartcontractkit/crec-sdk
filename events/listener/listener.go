package listener

import (
	"context"
	"fmt"
	"os"

	"github.com/rs/zerolog"

	"github.com/smartcontractkit/cvm-sdk/events/types"
	"github.com/smartcontractkit/cvm-sdk/internal/creclient/kafka"
)

type EventListenerOptions struct {
	Endpoints []string
}

type EventListener struct {
	kafkaReader *kafka.Reader
	logger      zerolog.Logger
}

func NewEventListener(opts *EventListenerOptions) *EventListener {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	logger.Info().Msg("Creating CVN event listener")

	// TODO: The topic is hard coded since using beholder otlp logs is a temporary solution.
	kafkaReader := kafka.NewKafkaEventListener(opts.Endpoints, "beholder_otlp_logs", "t1")

	return &EventListener{
		kafkaReader: kafkaReader,
		logger:      logger,
	}
}

func (l *EventListener) Read(cxt context.Context) (*types.VerifiableEvent, error) {
	l.logger.Debug().Msg("Reading event from broker")
	verifiableEvent, err := l.kafkaReader.ReadMessage(cxt)
	if err != nil {
		return nil, fmt.Errorf("error while reading message: %w", err)
	}
	return verifiableEvent, nil
}

func (l *EventListener) Close() error {
	l.logger.Info().Msg("Closing event listener")
	return l.kafkaReader.EventReader.Close()
}
