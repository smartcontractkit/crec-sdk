.PHONY: tools
tools:
	go install github.com/atombender/go-jsonschema@latest
	go install github.com/ethereum/go-ethereum/cmd/abigen@latest

.PHONY: generate
generate:
	go generate ./client
	go generate ./test_consumer/mockserver/api
	go-jsonschema services/dvp/schema/dvp.json -p events -o services/dvp/gen/events/events.go -t
	abigen --abi services/dvp/abi/CCIPDVPCoordinator.abi.json --pkg contract --out services/dvp/gen/contract/contract.go
	abigen --abi services/ccip/abi/IRouterClient.abi.json --pkg routerclient --out services/ccip/gen/routerclient/routerclient.go
	abigen --abi interfaces/abi/IERC20.abi.json --pkg erc20 --out interfaces/erc20/erc20.go
	abigen --abi interfaces/abi/IHoldManager.abi.json --pkg holdmanager --out interfaces/holdmanager/holdmanager.go

.PHONY: test
test:
	go test ./...