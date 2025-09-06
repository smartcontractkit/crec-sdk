.PHONY: tools
tools:
	go install github.com/ethereum/go-ethereum/cmd/abigen@latest
	go install github.com/vektra/mockery/v2@latest

.PHONY: generate
generate:
	mockery --name=KMSClient --dir=./transact/signer/kms --output=./transact/signer/kms/mocks --outpkg=mocks
	abigen --abi interfaces/abi/IERC20.abi.json --pkg erc20 --out interfaces/erc20/erc20.go
	abigen --abi interfaces/abi/IHoldManager.abi.json --pkg holdmanager --out interfaces/holdmanager/holdmanager.go
	go generate ./...

.PHONY: test
test:
	go test ./...