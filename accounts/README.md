# Account Deployment Example

This directory contains an example application demonstrating how to deploy and register accounts using the CVN SDK.

## Overview

The example application (`main.go`) shows the complete workflow for:

1. **CVN Client Setup** - Initialize connection to CVN service
2. **Accounts Service Creation** - Configure the accounts service with contract addresses
3. **Operation Preparation** - Prepare an ECDSA account deployment operation
4. **Transaction Signing** - Sign the operation using a local private key
5. **Operation Submission** - Submit the signed operation to CVN
6. **Status Monitoring** - Poll operation status until completion
7. **Account Registration** - Register the deployed account with CVN backend

## Prerequisites

Before running this example, ensure you have:

-   [ ] **CVN Service Running** - The CVN service accessible at the configured URL
-   [ ] **Smart Contracts Deployed** - All required contracts deployed to your target chain:
    -   Keystone Forwarder
    -   Account Factory
    -   ECDSA Signature Verifying Account Implementation
    -   RSA Signature Verifying Account Implementation (if using RSA accounts)
-   [ ] **Funded Account** - The account owner address has sufficient gas tokens
-   [ ] **API Key** - Valid API key for the CVN service

## Configuration

Update the constants in `main.go` with your actual values:

```go
const (
    CVN_API_URL = "https://your-cvn-service.com"  // Your CVN service URL
    CVN_API_KEY = "your-actual-api-key"           // Your CVN API key
    CHAIN_ID    = "1"                             // Target chain ID

    // Replace with your deployed contract addresses
    KEYSTONE_FORWARDER_ADDRESS = "0x..."
    ACCOUNT_FACTORY_ADDRESS = "0x..."
    // ... other addresses

    ACCOUNT_OWNER_ADDRESS = "0x..."               // Account owner
    EXAMPLE_PRIVATE_KEY = "0x..."                 // Private key for signing
)
```

## Running the Example

```bash
# From the accounts directory
go run main.go
```

## Security Warning

⚠️ **Never use the example private key in production!**

The example includes a hardcoded private key for demonstration purposes only. In production:

-   Generate secure private keys
-   Use proper key management systems
-   Consider hardware security modules (HSMs)
-   Implement proper access controls

## Expected Output

When successful, you should see output similar to:

```
1. Creating CVN client...
   ✓ CVN client created successfully

2. Creating accounts service...
   ✓ Accounts service created successfully

3. Preparing ECDSA account deployment operation...
   ✓ Operation prepared with ID: 1694123456

4. Creating transact client...
   ✓ Transact client created successfully

5. Signing and sending operation...
   ✓ Operation signed with hash: 0xabc123...
   ✓ Operation sent with ID: a47760fc-6eae-456f-84b9-e6e349d13281

6. Waiting for operation execution...
   Checking operation status (ID: a47760fc-6eae-456f-84b9-e6e349d13281)...
   Operation status: pending
   Waiting 3 seconds before next check...
   Operation status: settled
   ✓ Operation executed successfully!

7. Registering account with backend...
   ✓ Account registered successfully!
   Registered Account ID: df72db52-fa74-4dd8-8c53-a9742be70caa

🎉 Account deployment and registration completed successfully!
   Account ID: example-trading-account
   Owner: 0x742d35Cc6841C8532b39f87290c2a3e7C5F0b1b2
   Operation ID: a47760fc-6eae-456f-84b9-e6e349d13281
   CVN Operation ID: a47760fc-6eae-456f-84b9-e6e349d13281
```

## Troubleshooting

### Common Issues

**Connection errors:**

-   Verify CVN service URL and API key
-   Check network connectivity
-   Ensure CVN service is running

**Transaction failures:**

-   Check account has sufficient gas
-   Verify contract addresses are correct
-   Ensure private key corresponds to funded account

**Operation timeouts:**

-   Network congestion may cause delays
-   Increase timeout values if needed
-   Check blockchain explorer for transaction status

### Getting Help

For additional support:

-   Check the main [CVN SDK documentation](../README.md)
-   Review the [accounts service documentation](../services/accounts/README.md)
-   Open an issue on the repository

## Related Examples

-   [DTA Service Examples](../services/dta/README.md) - Digital Transfer Agent operations
-   [DVP Service Examples](../services/dvp/README.md) - Delivery vs Payment operations
