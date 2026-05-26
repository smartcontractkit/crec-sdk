// Package queries provides SDK helpers for asynchronous CRE Connect chain queries.
//
// Chain queries are channel-scoped requests for DON-backed blockchain reads. The
// current implementation supports EVM callContract-style reads ("evm_call"). A
// query is submitted to CREC, moves through accepted/sent/terminal states, and
// terminal callbacks may include a signed query.status event with a base64
// verifiable_result.
//
// # What is implemented
//
// The package wraps the generated CREC API client and adds:
//
//   - block-selection helpers: package vars [Latest] and [Finalized], plus
//     [BlockNumber] and [BlockNumberFromString] for explicit block numbers;
//   - submission helpers: [Client.Create] for generic query creation and
//     [Client.CreateEVMCall] for raw EVM call queries;
//   - lookup helpers: [Client.Get] and [Client.List];
//   - completion helper: [Client.Wait], which polls [Client.Get] until the query
//     is completed, failed, or expired;
//   - wrapped call helpers: [Client.CallContract] and
//     [Client.CallContractWithABI];
//   - decoding helpers: [ResultFromQuery], [DecodeVerifiableResult], and
//     [DecodeVerifiableResultBytes].
//
// [Client.Wait] uses the client's poll interval (2 seconds by default) and
// retries transient 429, 5xx, and common network errors. It stops immediately for
// validation errors, missing channels, missing queries, or context cancellation.
// Bound a single [Client.Wait] call either by setting Options.WaitTimeout when
// constructing the client or by passing a context.WithTimeout per call.
//
// # What each returned part means
//
// Query lifecycle data comes from two places:
//
//   - the query resource, fetched through [Client.Get], [Client.List], and
//     [Client.Wait]; it contains the query ID, channel ID, status, kind, chain
//     selector, timestamps, and API-side fields such as event_hash, proof, and
//     verifiable_result when available;
//   - channel events, fetched through the events client; terminal query.status
//     events carry OCR proof material in headers and the same
//     verifiable_result in the payload.
//
// [CallContractResult] is the SDK's decoded view of a terminal EVM call query:
//
//   - QueryID, Status, ChainSelector, and Query identify the API query record.
//     The display-friendly target (contract address for evm_call, transaction
//     hash for tx-style reads, etc.) lives on Query.Target — derived server-side
//     from the request per query kind — so consumers that need it should read
//     Query.Target (or branch on Query.QueryKind);
//   - VerifiableResult is the original base64 string returned by CREC;
//   - VerifiableQuery is the decoded JSON bytes of VerifiableResult for logging
//     or archival;
//   - RawReturnData is the EVM call return bytes decoded from raw_return_data;
//   - Proof and EventHash are the OCR proof/hash fields returned on the query
//     resource;
//   - Block is the resolved block number/hash/timestamp used for the read;
//   - Error is non-nil when the terminal chain execution failed or expired.
//
// Transport/admission/poll/decode failures are returned as Go errors. A
// successful Go error return with result.Error != nil means the query reached a
// terminal chain-query error state and the error details are part of the signed
// result.
//
// # Quick wrapped raw call
//
// [Client.CallContract] is the simplest raw-calldata path. It submits the query,
// waits for completion, decodes verifiable_result, and returns a
// [CallContractResult].
//
//	result, err := client.Queries.CallContract(ctx, queries.CallContractInput{
//	    ChannelID:       channelID,
//	    ChainSelector:   "16015286601757825753",
//	    ContractAddress: "0x1234567890123456789012345678901234567890",
//	    CallData:        []byte{0x18, 0x16, 0x0d, 0xdd}, // totalSupply()
//	    BlockSelection:  queries.Finalized,
//	    IdempotencyKey:  "total-supply-finalized-001",
//	})
//	if err != nil {
//	    // API, polling, or decode error.
//	    return err
//	}
//	if result.Error != nil {
//	    // The EVM call reached a signed terminal error, such as CALL_REVERTED.
//	    return fmt.Errorf("query failed: %s: %s", result.Error.Code, result.Error.Message)
//	}
//	totalSupply := new(big.Int).SetBytes(result.RawReturnData)
//
// Set FromAddress or Metadata directly on the input when needed:
//
//	fromAddress := "0x000000000000000000000000000000000000dEaD"
//	result, err = client.Queries.CallContract(ctx, queries.CallContractInput{
//	    ChannelID:       channelID,
//	    ChainSelector:   chainSelector,
//	    ContractAddress: tokenAddress,
//	    CallData:        "0x70a08231...", // balanceOf(address) calldata
//	    BlockSelection:  queries.Latest,
//	    IdempotencyKey:  "balance-query-001",
//	    FromAddress:     &fromAddress,
//	    Metadata:        map[string]interface{}{"client_reference_id": "balance-ui"},
//	})
//
// # Full ABI wrapper
//
// [Client.CallContractWithABI] packs function arguments, calls
// [Client.CallContract], and ABI-unpacks successful return data into Outputs.
// Human-readable function fragments and JSON ABI fragments are both accepted.
//
//	result, err := client.Queries.CallContractWithABI(ctx, queries.CallContractWithABIInput{
//	    ChannelID:       channelID,
//	    ChainSelector:   chainSelector,
//	    ContractAddress: tokenAddress,
//	    ABIFragment:     "function balanceOf(address owner) view returns (uint256)",
//	    FunctionName:    "balanceOf",
//	    Args:            []any{"0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"},
//	    BlockSelection:  queries.Finalized,
//	    IdempotencyKey:  "balance-finalized-001",
//	})
//	if err != nil {
//	    return err
//	}
//	if result.Error != nil {
//	    return fmt.Errorf("query failed: %s", result.Error.Message)
//	}
//	balance := result.Outputs[0].(*big.Int)
//
// # Manual lifecycle/raw wrappers
//
// Use [Client.CreateEVMCall] when you want to submit now and process completion
// later. It returns only the accepted query ID and status.
//
//	accepted, err := client.Queries.CreateEVMCall(ctx, queries.CallContractInput{
//	    ChannelID:       channelID,
//	    ChainSelector:   chainSelector,
//	    ContractAddress: tokenAddress,
//	    CallData:        []byte{0x18, 0x16, 0x0d, 0xdd},
//	    BlockSelection:  queries.Latest,
//	    IdempotencyKey:  "total-supply-async-001",
//	})
//	if err != nil {
//	    return err
//	}
//
// Later, await the terminal query resource and decode it:
//
//	query, err := client.Queries.Wait(ctx, channelID, accepted.QueryId)
//	if err != nil {
//	    return err
//	}
//	result, err := queries.ResultFromQuery(query)
//	if err != nil {
//	    return err
//	}
//
// For the lowest-level path, build the generated EVM params yourself and call
// [Client.Create]. The current SDK accepts only QueryKindEVMCall.
//
//	params := apiClient.EVMCallQueryParams{
//	    ContractAddress: apiClient.EthereumAddress(tokenAddress),
//	    CallData:        "0x18160ddd",
//	    BlockSelection:  queries.Finalized,
//	}
//	accepted, err = client.Queries.Create(ctx, queries.CreateInput{
//	    ChannelID:      channelID,
//	    IdempotencyKey: "manual-evm-call-001",
//	    QueryKind:      apiClient.QueryKindEVMCall,
//	    ChainSelector:  chainSelector,
//	    Params:         params,
//	})
//
// # Query all and find the one you need
//
// Use [Client.List] to page through query resources in a channel. Filter by
// status when possible, then match by QueryId or by your own stored idempotency
// key mapping.
//
//	limit := 100
//	statuses := []apiClient.QueryStatus{apiClient.QueryStatusCompleted, apiClient.QueryStatusFailed}
//	for offset := int64(0); ; offset += int64(limit) {
//	    page, hasMore, err := client.Queries.List(ctx, queries.ListInput{
//	        ChannelID: channelID,
//	        Status:    &statuses,
//	        Limit:     &limit,
//	        Offset:    &offset,
//	    })
//	    if err != nil {
//	        return err
//	    }
//	    for i := range page {
//	        if page[i].QueryId == queryID {
//	            result, err := queries.ResultFromQuery(&page[i])
//	            _ = result
//	            return err
//	        }
//	    }
//	    if !hasMore {
//	        break
//	    }
//	}
//
// # Detect completion from channel events
//
// Query status callbacks are emitted as query.status events on the same CREC
// channel. Use the events client when you need the signed event object, want to
// verify the DON OCR proof, or are already consuming the channel event stream.
//
//	limit := 100
//	eventTypes := []apiClient.EventType{apiClient.EventTypeQueryStatus}
//	channelEvents, hasMore, err := client.Events.SearchEvents(
//	    ctx,
//	    channelID,
//	    &apiClient.SearchChannelEventsParams{
//	        Type:  &eventTypes,
//	        Limit: &limit,
//	    },
//	)
//	_ = hasMore
//	if err != nil {
//	    return err
//	}
//	for i := range channelEvents {
//	    event := &channelEvents[i]
//	    payload, err := event.Payload.AsQueryStatusPayload()
//	    if err != nil || payload.QueryId != queryID {
//	        continue
//	    }
//
//	    verified, err := client.Events.VerifyQueryStatus(event)
//	    if err != nil || !verified {
//	        return err
//	    }
//
//	    decoded, err := client.Events.DecodeQueryStatusVerifiableEvent(&payload)
//	    if err != nil {
//	        return err
//	    }
//	    _ = decoded.Data
//	}
//
// For live polling, the events client's Poll method can be used instead of
// SearchEvents; filter returned events by EventTypeQueryStatus and
// payload.QueryId in the same way.
//
// # Detect completion without scanning events
//
// If you do not need the event envelope, use the query-specific resource APIs:
// [Client.Get] for a single check, [Client.List] for pages, or [Client.Wait] for
// polling until a terminal status. These methods read
// /channels/{channel_id}/queries and /channels/{channel_id}/queries/{query_id}.
// The SDK does not expose a separate query-only event stream; signed terminal
// callbacks are still channel events.
//
// # Decode and verify
//
// Decoding and verification are intentionally separate:
//
//   - The queries package decodes terminal query data. Use
//     [ResultFromQuery] for a query resource or [DecodeVerifiableResult]
//     for a base64 verifiable_result string.
//   - The events package verifies OCR signatures on query.status events. Create
//     the root client with crec.WithOrgID or crec.WithWorkflowOwner plus valid
//     signer configuration, then call VerifyQueryStatus on the event before
//     trusting it.
//
// Direct decode from a query resource:
//
//	query, err := client.Queries.Get(ctx, channelID, queryID)
//	if err != nil {
//	    return err
//	}
//	result, err := queries.ResultFromQuery(query)
//	if err != nil {
//	    return err
//	}
//	decoded, err := queries.DecodeVerifiableResult(result.VerifiableResult)
//
// Verified event decode:
//
//	verified, err := client.Events.VerifyQueryStatus(event)
//	if err != nil || !verified {
//	    return err
//	}
//	payload, err := event.Payload.AsQueryStatusPayload()
//	if err != nil {
//	    return err
//	}
//	decoded, err := client.Events.DecodeQueryStatusVerifiableEvent(&payload)
//
// After decoding, raw_return_data is still ABI-encoded EVM output. Use
// [Client.CallContractWithABI] for automatic output unpacking, or decode
// result.RawReturnData with go-ethereum/accounts/abi when using raw wrappers.
//
// # Input validation and normalization
//
// The SDK validates channel ID, idempotency key, query kind, chain selector,
// contract address, call data, block selection, from address, list limit, and
// offset before sending requests. Contract and from addresses are normalized to
// lower-case hex; call data may be []byte or a 0x-prefixed even-length hex
// string. If no from address is supplied for an EVM call, the zero address is
// sent.
//
// # Error handling
//
// All sentinel errors can be inspected with errors.Is:
//
//	if errors.Is(err, queries.ErrChannelNotFound) {
//	    // Missing channel.
//	}
//	if errors.Is(err, queries.ErrQueryNotFound) {
//	    // Missing query.
//	}
//	if errors.Is(err, queries.ErrDecodeVerifiableResult) {
//	    // Invalid or missing terminal verifiable_result.
//	}
package queries
