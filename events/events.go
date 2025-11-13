package events

import (
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/smartcontractkit/crec-sdk/client"
)

// ClientOptions holds the configuration options for the CREC events client.
// It includes options for logging and event retrieval.
//   - Logger: Optional logger instance.
//   - CRECClient: A client instance for interacting with the CREC system.
//   - EventsAfter: Unix timestamp to start retrieving events from. Defaults to current time if not set.
type ClientOptions struct {
	Logger      *zerolog.Logger
	CRECClient  *client.CRECClient
	EventsAfter int64
}

type Client struct {
	crecClient        *client.CRECClient
	logger            *zerolog.Logger
	eventsAfter       int64
	lastReadTimestamp int64
	lastReadEventId   uuid.UUID
}

// NewClient creates a new CREC events client with the provided CREC client and options.
// Returns a pointer to the Client and an error if any issues occur during initialization.
// If the CREC client or options are nil, it returns an error.
//   - opts: Options for configuring the CREC events client, see ClientOptions for details.
func NewClient(opts *ClientOptions) (*Client, error) {
	if opts == nil {
		return nil, fmt.Errorf("ClientOptions is required")
	}
	if opts.CRECClient == nil {
		return nil, fmt.Errorf("a valid CRECClient must be provided")
	}

	logger := opts.Logger
	if logger == nil {
		lgr := zerolog.New(os.Stdout).With().Timestamp().Logger()
		logger = &lgr
	}

	logger.Debug().Msg("Creating CREC events client")

	eventsAfter := opts.EventsAfter
	if eventsAfter == 0 {
		eventsAfter = time.Now().Unix()
	}

	return &Client{
		crecClient:        opts.CRECClient,
		logger:            logger,
		eventsAfter:       eventsAfter,
		lastReadTimestamp: 0,
		lastReadEventId:   uuid.Nil,
	}, nil
}

// Reset resets the internal state of the CREC events client.
// It clears the last read timestamp and event ID, allowing the client to start reading events from scratch.
func (c *Client) Reset() {
	c.logger.Debug().Msg("Resetting event reader state")
	c.lastReadTimestamp = 0
	c.lastReadEventId = uuid.Nil
}
