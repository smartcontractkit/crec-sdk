package reader

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/smartcontractkit/cvn-sdk/client"
)

type EventReaderOptions struct {
	Logger      *zerolog.Logger
	EventsAfter int64
}

type EventReader struct {
	cvnClient         *client.ClientWithResponses
	logger            *zerolog.Logger
	eventsAfter       int64
	lastReadTimestamp int64
	lastReadEventId   uuid.UUID
}

func NewEventReader(cvnClient *client.ClientWithResponses, opts *EventReaderOptions) (*EventReader, error) {
	if cvnClient == nil {
		return nil, errors.New("EventReader requires a valid CVN client")
	}
	if opts == nil {
		return nil, errors.New("EventReader requires a valid options struct")
	}

	logger := opts.Logger
	if logger == nil {
		lgr := zerolog.New(os.Stdout).With().Timestamp().Logger()
		logger = &lgr
	}

	logger.Info().Msg("Creating CVN event reader")

	eventsAfter := opts.EventsAfter
	if eventsAfter == 0 {
		eventsAfter = time.Now().Unix()
	}

	return &EventReader{
		cvnClient:         cvnClient,
		logger:            logger,
		eventsAfter:       eventsAfter,
		lastReadTimestamp: 0,
		lastReadEventId:   uuid.Nil,
	}, nil
}

func (l *EventReader) Read(ctx context.Context) (*[]client.Event, error) {
	l.logger.Debug().Msg("Reading events from CVN")

	eventsAfter := l.lastReadTimestamp
	if eventsAfter == 0 {
		eventsAfter = l.eventsAfter
	}

	params := &client.GetEventsParams{
		CreatedGt: &eventsAfter,
	}
	if l.lastReadEventId != uuid.Nil {
		params.StartingAfter = &l.lastReadEventId
	}

	if l.logger.GetLevel() <= zerolog.DebugLevel {
		reqParams, err := json.Marshal(params)
		if err != nil {
			l.logger.Error().Err(err).Msg("Failed to marshal request parameters")
			return nil, err
		}
		l.logger.Debug().
			RawJSON("params", reqParams).
			Msg("Reading events from CVN")
	}

	resp, err := l.cvnClient.GetEventsWithResponse(ctx, params)

	if err != nil {
		l.logger.Error().Err(err).Msg("Failed to read events from CVN")
		return nil, err
	}

	if resp.StatusCode() != 200 {
		l.logger.Error().Int("status", resp.StatusCode()).Msg("Failed to read events from CVN, unexpected status code")
		return nil, nil
	}

	if resp.JSON200 == nil || resp.JSON200.Data == nil {
		l.logger.Warn().Msg("No events found in response from CVN")
		return nil, nil
	}

	var eventList = *resp.JSON200.Data
	if len(eventList) > 0 {
		l.lastReadTimestamp = *eventList[len(eventList)-1].CreatedAt
		l.lastReadEventId = *eventList[len(eventList)-1].EventId
	}

	return resp.JSON200.Data, nil
}

func (l *EventReader) Reset() {
	l.logger.Info().Msg("Resetting event reader state")
	l.lastReadTimestamp = 0
	l.lastReadEventId = uuid.Nil
}
