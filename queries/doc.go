// Package queries provides SDK helpers for asynchronous CRE Connect chain queries.
//
// Chain queries are scoped to CREC channels and currently support EVM
// callContract-style reads using raw calldata. The package wraps the generated
// API client and adds:
//   - typed block-selection helpers,
//   - raw EVM call creation and polling,
//   - decoding of terminal verifiable_result payloads,
//   - exposing the API-provided/fallback display target for terminal results,
//   - ABI calldata encoding / return-data decoding convenience helpers.
//
// Typical usage through the root SDK client:
//
//	result, err := client.Queries.CallContract(
//	    ctx,
//	    channelID,
//	    "16015286601757825753",
//	    "0x1234567890123456789012345678901234567890",
//	    []byte{0x18, 0x16, 0x0d, 0xdd}, // totalSupply()
//	    queries.LatestBlockSelection(),
//	    "total-supply-query-001",
//	)
//	if err != nil {
//	    // transport/admission/poll/decode error
//	}
//	if result.Error != nil {
//	    // terminal chain-query execution error
//	}
//
// ABI convenience:
//
//	result, err := client.Queries.CallContractWithABI(
//	    ctx,
//	    channelID,
//	    "16015286601757825753",
//	    tokenAddress,
//	    `[{"type":"function","name":"totalSupply","inputs":[],"outputs":[{"type":"uint256"}],"stateMutability":"view"}]`,
//	    "totalSupply",
//	    nil,
//	    queries.FinalizedBlockSelection(),
//	    "total-supply-finalized-001",
//	)
package queries
