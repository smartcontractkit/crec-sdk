.PHONY: tools
tools:
	go install github.com/atombender/go-jsonschema@latest
	go install github.com/ethereum/go-ethereum/cmd/abigen@latest

.PHONY: generate
generate:
	go generate ./client
	go generate ./internal/mockserver/api
	go-jsonschema services/dvp/schema/dvp.json -p events -o services/dvp/gen/events/events.go -t
	abigen --abi services/dvp/abi/CCIPDVPCoordinator.abi.json --pkg contract --out services/dvp/gen/contract/contract.go
	abigen --abi services/ccip/abi/IRouterClient.abi.json --pkg routerclient --out services/ccip/gen/routerclient/routerclient.go
	abigen --abi services/ccip/abi/IERC20.abi.json --pkg erc20 --out services/ccip/gen/erc20/erc20.go
