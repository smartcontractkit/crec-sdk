# Queries

Use `queries` when your app needs a trustworthy chain read without wiring its own RPC polling.

```go
package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/google/uuid"

	"github.com/smartcontractkit/crec-sdk"
	"github.com/smartcontractkit/crec-sdk/queries"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	client, err := crec.NewClient("https://api.crec.example.com", "your-api-key")
	if err != nil {
		log.Fatal(err)
	}

	channelID := uuid.MustParse("00000000-0000-0000-0000-000000000000")
	token := "0x1234567890123456789012345678901234567890"

	result, err := client.Queries.CallContractWithABI(
		ctx,
		channelID,
		"16015286601757825753",
		token,
		"function totalSupply() view returns (uint256)",
		"totalSupply",
		nil,
		queries.Finalized(),
		"total-supply-finalized-demo",
	)
	if err != nil {
		log.Fatal(err)
	}
	if result.Error != nil {
		log.Fatalf("query failed: %s %s", result.Error.Code, result.Error.Message)
	}

	totalSupply := result.Outputs[0].(*big.Int)
	fmt.Println("finalized total supply:", totalSupply)
}
```
