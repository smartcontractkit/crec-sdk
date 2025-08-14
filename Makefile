.PHONY: tools
tools:
	go install github.com/atombender/go-jsonschema@latest
	go install github.com/ethereum/go-ethereum/cmd/abigen@latest

.PHONY: generate
generate:
	go generate ./client
	go generate ./internal/mockserver/api
	go-jsonschema services/dvp/schema/dvp.json -p events -o services/dvp/gen/events/events.go -t
	go-jsonschema services/dta/schema/dta.json -p events -o services/dta/gen/events/events.go -t
	go-jsonschema services/ace/ccid/identityregistry/schema/identity_registry.json -p events -o services/ace/ccid/identityregistry/gen/events/events.go -t
	go-jsonschema services/ace/ccid/credentialregistry/schema/credential_registry.json -p events -o services/ace/ccid/credentialregistry/gen/events/events.go -t
	abigen --abi services/dvp/abi/CCIPDVPCoordinator.abi.json --pkg contract --out services/dvp/gen/contract/contract.go
	abigen --abi services/ccip/abi/IRouterClient.abi.json --pkg routerclient --out services/ccip/gen/routerclient/routerclient.go
	abigen --abi interfaces/abi/IERC20.abi.json --pkg erc20 --out interfaces/erc20/erc20.go
	abigen --abi interfaces/abi/IHoldManager.abi.json --pkg holdmanager --out interfaces/holdmanager/holdmanager.go
	abigen --abi services/dta/abi/DTAOpenMarketplaceU.abi.json --pkg dtaopenmarketplace --out services/dta/gen/dtaopenmarketplace/dtaopenmarketplace.go
	abigen --abi services/dta/abi/DTAWalletU.abi.json --pkg dtawallet --out services/dta/gen/dtawallet/dtawallet.go
	abigen --abi services/ace/ccid/identityregistry/abi/IdentityRegistry.abi.json --pkg identityregistry --out services/ace/ccid/identityregistry/gen/identityregistry/identityregistry.go
	abigen --abi services/ace/ccid/credentialregistry/abi/CredentialRegistry.abi.json --pkg credentialregistry --out services/ace/ccid/credentialregistry/gen/credentialregistry/credentialregistry.go

.PHONY: test
test:
	go test ./...