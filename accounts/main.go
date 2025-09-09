// Example Application: Account Deployment and Registration
//
// This example demonstrates how to:
// 1. Create a CVN client and accounts service
// 2. Prepare an ECDSA account deployment operation
// 3. Sign and submit the operation using the transact client
// 4. Monitor operation status until completion
// 5. Register the deployed account with the CVN backend
//
// Before running this example:
// - Replace all placeholder constants with your actual values
// - Ensure your CVN service is running and accessible
// - Deploy the required smart contracts to your target chain
// - Fund the account owner address with sufficient gas tokens
//
// Security Warning:
// - Never use the example private key in production
// - Generate secure keys and store them safely
// - Use proper key management systems for production deployments

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"

	apiClient "github.com/smartcontractkit/cvn-api-go/client"
	"github.com/smartcontractkit/cvn-sdk/client"
	"github.com/smartcontractkit/cvn-sdk/services/accounts"
	"github.com/smartcontractkit/cvn-sdk/transact"
	"github.com/smartcontractkit/cvn-sdk/transact/signer/local"
)

// Example constants - replace with actual values for your environment
const (
	// CVN service configuration
	CVN_API_URL = "http://localhost:8088" // Replace with your CVN service URL
	CVN_API_KEY = "your-api-key-here"     // Replace with your actual API key
	CHAIN_ID    = "1337"                  // Replace with your target chain ID (e.g., "1" for mainnet)

	// Smart contract addresses - replace with your deployed contract addresses
	KEYSTONE_FORWARDER_ADDRESS                  = "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE" // Keystone forwarder contract
	ACCOUNT_FACTORY_ADDRESS                     = "0x7Eb6D2Bf84C394A1718a60f0f89FBc4626eCdbA1" // Account factory contract
	ECDSA_SIGNATURE_VERIFYING_ACCOUNT_IMPL_ADDR = "0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc" // ECDSA account implementation
	RSA_SIGNATURE_VERIFYING_ACCOUNT_IMPL_ADDR   = "0xeb457346d2218f7f77aa23ac6d9e394b505dd621" // RSA account implementation

	// Account configuration
	ACCOUNT_OWNER_ADDRESS = "0x742d35Cc6841C8532b39f87290c2a3e7C5F0b1b2" // Replace with actual account owner
	ACCOUNT_ID            = "example-trading-account"                    // Unique identifier for your account

	// WARNING: This is an example private key - NEVER use this in production!
	// Generate a secure private key for your environment
	EXAMPLE_PRIVATE_KEY = "0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318"
)

var (
	// Example allowed signers for ECDSA account - replace with actual signer addresses
	ALLOWED_SIGNERS = []string{
		"0x8ba1f109551bD432803012645Hac136c5C3bc7C8", // Replace with actual signer address
		"0x9Ac9c636404C8db5A1c6B2AC3b9f5C7f56789eEf", // Replace with actual signer address
	}
)

func main() {
	ctx := context.Background()

	// Step 1: Instantiate CVN client
	fmt.Println("1. Creating CVN client...")
	cvnClient, err := client.NewCVNClient(CVN_API_URL, CVN_API_KEY)
	if err != nil {
		log.Fatalf("Failed to create CVN client: %v", err)
	}
	fmt.Println("   ✓ CVN client created successfully")

	// Step 2: Instantiate accounts service
	fmt.Println("2. Creating accounts service...")
	accountsService, err := accounts.NewService(&accounts.ServiceOptions{
		KeystoneForwarderAddress:                  KEYSTONE_FORWARDER_ADDRESS,
		AccountFactoryAddress:                     ACCOUNT_FACTORY_ADDRESS,
		ECDSASignatureVerifyingAccountImplAddress: ECDSA_SIGNATURE_VERIFYING_ACCOUNT_IMPL_ADDR,
		RSASignatureVerifyingAccountImplAddress:   RSA_SIGNATURE_VERIFYING_ACCOUNT_IMPL_ADDR,
	})
	if err != nil {
		log.Fatalf("Failed to create accounts service: %v", err)
	}
	fmt.Println("   ✓ Accounts service created successfully")

	// Step 3: Prepare ECDSA account deployment operation
	fmt.Println("3. Preparing ECDSA account deployment operation...")
	operation, err := accountsService.PrepareDeployNewECDSAAccountOperation(
		ACCOUNT_OWNER_ADDRESS,
		ALLOWED_SIGNERS,
		ACCOUNT_ID,
	)
	if err != nil {
		log.Fatalf("Failed to prepare ECDSA account operation: %v", err)
	}
	fmt.Printf("   ✓ Operation prepared with ID: %s\n", operation.ID.String())

	// Create transact client for signing and sending operations
	fmt.Println("4. Creating transact client...")
	transactClient, err := transact.NewClient(&transact.ClientOptions{
		CVNClient: cvnClient,
		ChainId:   CHAIN_ID,
	})
	if err != nil {
		log.Fatalf("Failed to create transact client: %v", err)
	}
	fmt.Println("   ✓ Transact client created successfully")

	// Create signer
	privateKey, err := crypto.HexToECDSA(EXAMPLE_PRIVATE_KEY)
	if err != nil {
		log.Fatalf("Failed to create private key: %v", err)
	}
	operationSigner := local.NewSigner(privateKey)

	// Step 4: Sign and send the operation
	fmt.Println("5. Signing and sending operation...")
	operationHash, operationSignature, err := transactClient.SignOperation(ctx, operation, operationSigner)
	if err != nil {
		log.Fatalf("Failed to sign operation: %v", err)
	}
	fmt.Printf("   ✓ Operation signed with hash: %s\n", operationHash.Hex())

	txHash, err := transactClient.SendSignedOperation(ctx, operation, operationSignature)
	if err != nil {
		log.Fatalf("Failed to send operation: %v", err)
	}
	fmt.Printf("   ✓ Operation sent with ID: %s\n", txHash.OperationId)

	// Step 5: Wait for operation to be executed successfully
	fmt.Println("6. Waiting for operation execution...")
	operationUUID, err := uuid.Parse(txHash.OperationId.String())
	if err != nil {
		log.Fatalf("Failed to parse operation UUID: %v", err)
	}
	operationID := operationUUID.String()

	for {
		fmt.Printf("   Checking operation status (ID: %s)...\n", operationID)

		status, err := transactClient.GetOperation(ctx, operationUUID)
		if err != nil {
			log.Printf("   Error getting operation status: %v", err)
		} else if status != nil {
			fmt.Printf("   Operation status: %s\n", status.Status)

			if status.Status == "settled" {
				fmt.Println("   ✓ Operation executed successfully!")
				break
			} else if status.Status == "failed" {
				log.Fatalf("   ✗ Operation failed")
			}
		} else {
			fmt.Println("   Operation not found yet...")
		}

		fmt.Println("   Waiting 3 seconds before next check...")
		time.Sleep(3 * time.Second)
	}

	// Step 6: Register the account using PostAccountsWithResponse endpoint
	fmt.Println("7. Registering account with backend...")

	// Create account registration data
	accountData := apiClient.CreateAccount{
		Address: ACCOUNT_OWNER_ADDRESS,
		ChainId: CHAIN_ID,
	}

	response, err := cvnClient.PostAccountsWithResponse(ctx, accountData)
	if err != nil {
		log.Fatalf("Failed to register account: %v", err)
	}

	if response.StatusCode() != 201 {
		log.Fatalf("Failed to register account, status: %d", response.StatusCode())
	}

	fmt.Println("   ✓ Account registered successfully!")
	fmt.Printf("   Registered Account ID: %s\n", response.JSON201.AccountId)

	fmt.Println("\n🎉 Account deployment and registration completed successfully!")
	fmt.Printf("   Account ID: %s\n", ACCOUNT_ID)
	fmt.Printf("   Owner: %s\n", ACCOUNT_OWNER_ADDRESS)
	fmt.Printf("   Operation ID: %s\n", operationID)
	fmt.Printf("   CVN Operation ID: %s\n", txHash.OperationId)
}
