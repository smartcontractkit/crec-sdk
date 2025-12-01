.PHONY: tools
tools:
	go install github.com/ethereum/go-ethereum/cmd/abigen@latest
	go install github.com/vektra/mockery/v2@latest

.PHONY: generate
generate:
	mockery --name=KMSClient --dir=./transact/signer/kms --output=./transact/signer/kms/mocks --outpkg=mocks
	abigen --abi interfaces/abi/IERC20.abi.json --pkg erc20 --out interfaces/erc20/erc20.go
	abigen --abi interfaces/abi/IHoldManager.abi.json --pkg holdmanager --out interfaces/holdmanager/holdmanager.go

.PHONY: test
test:
	go test ./...

.PHONY: coverage
coverage:
	go test ./... -cover

.PHONY: coverage-report
coverage-report:
	go test ./... -coverprofile=coverage.out -covermode=atomic
	go tool cover -func=coverage.out

.PHONY: vendor
vendor:
	go mod tidy
	go mod download
	go mod vendor

.PHONY: docs
docs:
	@echo "Docs server running at http://localhost:8080/github.com/smartcontractkit/crec-sdk"
	pkgsite -http :8080
