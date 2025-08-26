.PHONY: tools
tools:
	go install github.com/ethereum/go-ethereum/cmd/abigen@latest

.PHONY: generate
generate:
	abigen --abi interfaces/abi/IERC20.abi.json --pkg erc20 --out interfaces/erc20/erc20.go
	abigen --abi interfaces/abi/IHoldManager.abi.json --pkg holdmanager --out interfaces/holdmanager/holdmanager.go

.PHONY: test
test:
	go test ./...