# Queries

Use `queries` when your app needs a trustworthy chain read without wiring its own RPC polling.

Query creation is asynchronous: the API accepts the request and returns a `query_id` before the
chain read has finished. To await the terminal result, create the query, call `Wait` with the
returned `query_id`, then decode the terminal query with `ResultFromQuery`. The convenience helpers
`CallContract` and `CallContractWithABI` perform this create-and-wait flow internally.

To verify the emitted terminal `query.status` event, configure an organization ID or workflow owner,
then retrieve the event, verify its OCR proof, and make sure its `verifiable_result` matches the
terminal query result before decoding the returned bytes.

```go
package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/google/uuid"
	apiClient "github.com/smartcontractkit/crec-api-go/client"

	"github.com/smartcontractkit/crec-sdk"
	"github.com/smartcontractkit/crec-sdk/queries"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	client, err := crec.NewClient(
		"https://api.crec.example.com",
		"your-api-key",
		crec.WithOrgID("your-org-id"),
	)
	if err != nil {
		log.Fatal(err)
	}

	channelID := uuid.MustParse("00000000-0000-0000-0000-000000000000")
	token := "0x1234567890123456789012345678901234567890"

	accepted, err := client.Queries.CreateEVMCall(
		ctx,
		queries.CallContractInput{
			ChannelID:       channelID,
			ChainSelector:   "16015286601757825753",
			ContractAddress: token,
			CallData:        []byte{0x18, 0x16, 0x0d, 0xdd}, // totalSupply()
			BlockSelection:  queries.Finalized(),
			IdempotencyKey:  "total-supply-finalized-demo",
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	query, err := client.Queries.Wait(ctx, channelID, accepted.QueryId)
	if err != nil {
		log.Fatal(err)
	}

	result, err := client.Queries.ResultFromQuery(query)
	if err != nil {
		log.Fatal(err)
	}
	if result.Error != nil {
		log.Fatalf("query failed: %s %s", result.Error.Code, result.Error.Message)
	}

	queryStatusEvent, err := findQueryStatusEvent(ctx, client, channelID, accepted.QueryId)
	if err != nil {
		log.Fatal(err)
	}

	verified, err := client.Events.VerifyQueryStatus(queryStatusEvent)
	if err != nil {
		log.Fatal(err)
	}
	if !verified {
		log.Fatal("query.status event did not have enough valid DON signatures")
	}

	queryStatusPayload, err := queryStatusEvent.Payload.AsQueryStatusPayload()
	if err != nil {
		log.Fatal(err)
	}
	if queryStatusPayload.VerifiableResult == nil || *queryStatusPayload.VerifiableResult != result.VerifiableResult {
		log.Fatal("verified query.status event does not match the terminal query result")
	}

	decodedQueryEvent, err := client.Events.DecodeQueryStatusVerifiableEvent(&queryStatusPayload)
	if err != nil {
		log.Fatal(err)
	}
	if decodedQueryEvent.Data.QueryId != accepted.QueryId {
		log.Fatal("verified query.status event has a different query_id")
	}

	totalSupply := new(big.Int).SetBytes(result.RawReturnData)
	fmt.Println("finalized total supply:", totalSupply)
}

func findQueryStatusEvent(
	ctx context.Context,
	sdkClient *crec.Client,
	channelID uuid.UUID,
	queryID uuid.UUID,
) (*apiClient.Event, error) {
	limit := 100
	eventTypes := []apiClient.EventType{apiClient.EventTypeQueryStatus}

	for offset := int64(0); ; offset += int64(limit) {
		channelEvents, hasMore, err := sdkClient.Events.SearchEvents(
			ctx,
			channelID,
			&apiClient.SearchChannelEventsParams{
				Limit:  &limit,
				Offset: &offset,
				Type:   &eventTypes,
			},
		)
		if err != nil {
			return nil, err
		}

		for i := range channelEvents {
			if channelEvents[i].Headers.Type != apiClient.EventTypeQueryStatus {
				continue
			}

			payload, err := channelEvents[i].Payload.AsQueryStatusPayload()
			if err != nil {
				return nil, fmt.Errorf("decode query.status payload: %w", err)
			}
			if payload.QueryId == queryID {
				return &channelEvents[i], nil
			}
		}

		if !hasMore {
			break
		}
	}

	return nil, fmt.Errorf("query.status event for query %s not found", queryID)
}
```
