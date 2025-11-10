# Watchers Service

The Watchers Service provides a comprehensive Go SDK for managing blockchain event watchers in the CREC platform. Watchers monitor specific smart contract events on blockchain networks and trigger workflows when those events occur. They can be configured either with pre-defined domain ABIs or custom event ABIs.

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Service Configuration](#service-configuration)
- [Watcher Operations](#watcher-operations)
- [Usage Examples](#usage-examples)
- [Error Handling](#error-handling)

## Overview

The Watchers Service is a **blockchain monitoring service** that enables real-time event detection on smart contracts. Watchers act as:

- **Event Detectors**: Monitor blockchain for specific contract events
- **Workflow Triggers**: Automatically trigger actions when events are detected
- **Domain-Aware**: Support for pre-configured domain types (DvP, DTA, test consumers)
- **Custom ABI Support**: Define custom event structures for any smart contract

Think of watchers as **"event listeners"** that continuously monitor blockchain activity and notify your application when specific events occur.

### Key Benefits

- ✅ **Real-Time Monitoring** - Detect blockchain events as they happen
- ✅ **Domain Support** - Pre-configured ABIs for DvP, DTA, and test consumer contracts
- ✅ **Custom Events** - Define any event structure with custom ABIs
- ✅ **Automatic Deployment** - Watchers are deployed automatically after creation
- ✅ **Status Tracking** - Monitor watcher deployment and health status
- ✅ **Channel Organization** - Watchers are scoped to channels for better organization

## Architecture

```mermaid
graph TD
    A[Watchers Service Architecture] --> B[Service Layer - Go SDK]
    A --> C[CREC API Backend]
    A --> D[Channel Scoping]

    B --> B1[CreateWatcherWithDomain]
    B --> B2[CreateWatcherWithABI]
    B --> B3[FindWatchersByChannel]
    B --> B4[FindWatcherByID]
    B --> B5[UpdateWatcher]
    B --> B6[WaitForActive]
    B --> B7[DeleteWatcher]
    B --> B8[WaitForDeleted]

    C --> C1[POST /channels/:id/watchers]
    C --> C2[GET /channels/:id/watchers]
    C --> C3[GET /channels/:id/watchers/:watcher_id]
    C --> C4[PATCH /channels/:id/watchers/:watcher_id]
    C --> C5[DELETE /channels/:id/watchers/:watcher_id]

    D --> D1[Channel Validation]
    D --> D2[Event Filtering]
    D --> D3[Status Management]
```

## Service Configuration

### ServiceOptions

Configure the watchers service with the CREC API client:

```go
import (
    "github.com/smartcontractkit/crec-sdk/client"
    "github.com/smartcontractkit/crec-sdk/services/watchers"
)

// 1. Create CREC API client
crecClient, err := client.NewCRECClient(&client.ClientOptions{
    BaseURL: "https://api.crec.chainlink.com",
    APIKey:  "your-api-key",
})
if err != nil {
    log.Fatal(err)
}

// 2. Create Watchers service
watchersService, err := watchers.NewService(&watchers.ServiceOptions{
    CRECClient: crecClient,
    Logger:     logger, // Optional: zerolog.Logger instance
})
if err != nil {
    log.Fatal(err)
}
```

**Configuration Details:**

- **CRECClient**: Required. The authenticated CREC API client instance.
- **Logger**: Optional. A zerolog.Logger instance for service logging. If not provided, a default logger will be created.

## Watcher Operations

### CreateWatcherWithDomain

Creates a new watcher using a pre-defined domain type. Domains provide pre-configured event ABIs for known contract types.

**Supported Domains:**
- `dvp` - Delivery vs Payment contracts
- `dta` - Digital Trade Agreements contracts  
- `test_consumer` - Test consumer contracts

**Input Parameters:**

- `ChannelID`: UUID of the channel where the watcher will be created
- `Name`: Optional friendly name for the watcher
- `ChainSelector`: The chain selector to identify the blockchain
- `Address`: Smart contract address to watch for events
- `Domain`: Domain type (dvp, dta, or test_consumer)
- `Events`: List of event names to watch for within the domain

**Returns:**

- `*apiClient.Watcher`: The created watcher with ID and status
- `error`: Error if the operation fails

**Example:**

```go
name := "Production DVP Watcher"
watcher, err := watchersService.CreateWatcherWithDomain(ctx, channelID, watchers.CreateWatcherWithDomainInput{
    Name:          &name,
    ChainSelector: 1,
    Address:       "0x1234567890123456789012345678901234567890",
    Domain:        "dvp",
    Events:        []string{"OperationExecuted", "OperationCreated"},
})
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Created watcher: %s (Status: %s)\n", watcher.WatcherId, watcher.Status)
```

### Create a Watcher with Custom ABI

Creates a new watcher with a custom event ABI. Use this when monitoring contracts that don't fit into pre-defined domains.

**Input Parameters:**

- `ChannelID`: UUID of the channel where the watcher will be created
- `Name`: Optional friendly name for the watcher
- `ChainSelector`: The chain selector to identify the blockchain
- `Address`: Smart contract address to watch for events
- `Events`: List of event names to watch for
- `ABI`: Array of event ABI definitions

> **Note on ABI Types**: The SDK currently only supports event types. The `EventABI.Type` field **must be set to `"event"`**. The SDK validates this and will return an error if you provide any other value (e.g., `"function"`, `"error"`). Support for other ABI types may be added in future versions.

**Returns:**

- `*apiClient.Watcher`: The created watcher with ID and status
- `error`: Error if the operation fails

**Example:**

```go
name := "ERC20 Transfer Watcher"
watcher, err := watchersService.CreateWatcherWithABI(ctx, channelID, watchers.CreateWatcherWithABIInput{
    Name:          &name,
    ChainSelector: 1,
    Address:       "0x1234567890123456789012345678901234567890",
    Events:        []string{"Transfer"},
    ABI: []watchers.EventABI{
        {
            Name: "Transfer",
            Type: "event",  // Required: must be "event" (validated by SDK)
            Inputs: []watchers.EventABIInput{
                {
                    Name:         "from",
                    Type:         "address",
                    InternalType: "address",
                    Indexed:      true,
                },
                {
                    Name:         "to",
                    Type:         "address",
                    InternalType: "address",
                    Indexed:      true,
                },
                {
                    Name:         "value",
                    Type:         "uint256",
                    InternalType: "uint256",
                    Indexed:      false,
                },
            },
        },
    },
})
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Created watcher: %s\n", watcher.WatcherId)
```

### FindWatchersByChannel

Retrieves a list of watchers for a channel with optional filtering and pagination.

**Input Parameters:**

- `ChannelID`: UUID of the channel to list watchers from (required)
- `Filters`: Optional filtering criteria:
  - `Limit`: Maximum number of watchers to return (default: 20)
  - `Offset`: Number of watchers to skip for pagination (default: 0)
  - `Name`: Filter by name (partial match, case-insensitive)
  - `Status`: Filter by status (pending, active, failed, deleting, deleted)
  - `ChainSelector`: Filter by chain selector
  - `Address`: Filter by contract address
  - `Domain`: Filter by domain
  - `EventName`: Filter by specific event name being monitored

**Returns:**

- `*apiClient.WatcherList`: List of watchers with pagination info
- `error`: Error if the operation fails

**Example:**

```go
// List all watchers with default pagination
watchersList, err := watchersService.FindWatchersByChannel(ctx, channelID, watchers.WatcherFilters{})
if err != nil {
    log.Fatal(err)
}

for _, w := range watchersList.Data {
    fmt.Printf("- Watcher: %s (Status: %s, Address: %s)\n", 
        w.WatcherId, w.Status, w.Address)
}

if watchersList.HasMore {
    fmt.Println("More watchers available...")
}
```

**Example with Filters:**

```go
// Filter active watchers on Ethereum mainnet
status := watchers.StatusActive
chainSelector := uint64(1)
limit := 10

watchersList, err := watchersService.FindWatchersByChannel(ctx, channelID, watchers.WatcherFilters{
    Status:        &status,
    ChainSelector: &chainSelector,
    Limit:         &limit,
})
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Found %d active watchers on Ethereum\n", len(watchersList.Data))
```

### FindWatcherByID

Retrieves a specific watcher by its ID within a channel.

**Input Parameters:**

- `channelID`: UUID of the channel containing the watcher
- `watcherID`: UUID of the watcher to retrieve

**Returns:**

- `*apiClient.Watcher`: The watcher details
- `error`: Error if the watcher is not found or the operation fails

**Example:**

```go
watcher, err := watchersService.FindWatcherByID(ctx, channelID, watcherID)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Watcher: %s\n", watcher.WatcherId)
fmt.Printf("Status: %s\n", watcher.Status)
fmt.Printf("Address: %s\n", watcher.Address)
fmt.Printf("Events: %v\n", watcher.Events)
```

### UpdateWatcher

Updates a watcher's metadata (currently only the name can be updated).

**Input Parameters:**

- `channelID`: UUID of the channel containing the watcher
- `watcherID`: UUID of the watcher to update
- `input`: Update parameters
  - `Name`: New name for the watcher (required)

**Returns:**

- `*apiClient.Watcher`: The updated watcher
- `error`: Error if the operation fails

**Example:**

```go
updatedWatcher, err := watchersService.UpdateWatcher(ctx, channelID, watcherID, watchers.UpdateWatcherInput{
    Name: "Updated Production DVP Watcher",
})
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Updated watcher name to: %s\n", *updatedWatcher.Name)
```

### WaitForActive

Polls a watcher until it reaches active status or fails. Watchers are created in a `pending` state and must be deployed before becoming `active`.

**Input Parameters:**

- `channelID`: UUID of the channel containing the watcher
- `watcherID`: UUID of the watcher to monitor
- `maxWaitTime`: Maximum duration to wait for activation

**Returns:**

- `*apiClient.Watcher`: The active watcher
- `error`: Error if deployment fails or timeout occurs

**Example:**

```go
// Wait up to 2 minutes for the watcher to become active
activeWatcher, err := watchersService.WaitForActive(ctx, channelID, watcher.WatcherId, 2*time.Minute)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Watcher is now active: %s\n", activeWatcher.WatcherId)
```

### DeleteWatcher

Deletes a watcher from a channel. The watcher will stop monitoring events and be removed from the system.

**Async Support:** This operation can be either synchronous (204 No Content) or asynchronous (202 Accepted). 
For async deletions, use `WaitForDeleted` to wait for completion.

**Input Parameters:**

- `channelID`: UUID of the channel containing the watcher
- `watcherID`: UUID of the watcher to delete

**Returns:**

- `error`: Error if the watcher is not found or the operation fails

**Example:**

```go
// Initiate deletion
err := watchersService.DeleteWatcher(ctx, channelID, watcherID)
if err != nil {
    log.Fatal(err)
}

fmt.Println("Watcher deletion initiated")
```

### WaitForDeleted

Waits for a watcher to be fully deleted. This is useful when deletion is asynchronous (returns 202 Accepted).

The method polls the watcher status until it reaches "deleted" state, the watcher is not found (404), 
or the timeout is reached.

**Input Parameters:**

- `channelID`: UUID of the channel containing the watcher
- `watcherID`: UUID of the watcher to wait for
- `maxWaitTime`: Maximum time to wait for deletion

**Returns:**

- `error`: Error if the watcher is not deleted within the timeout or if an error occurs

**Example:**

```go
// Delete watcher (may return 202 for async)
err := watchersService.DeleteWatcher(ctx, channelID, watcherID)
if err != nil {
    log.Fatal(err)
}

// Wait for deletion to complete
err = watchersService.WaitForDeleted(ctx, channelID, watcherID, 30*time.Second)
if err != nil {
    log.Fatal(err)
}

fmt.Println("Watcher fully deleted")
```

## Usage Examples

### Complete Workflow: Creating and Monitoring a Watcher

This example demonstrates a complete workflow for creating a watcher with a domain, waiting for activation, and monitoring its status.

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/google/uuid"
    "github.com/smartcontractkit/crec-sdk/client"
    "github.com/smartcontractkit/crec-sdk/services/channels"
    "github.com/smartcontractkit/crec-sdk/services/watchers"
)

func main() {
    ctx := context.Background()

    // 1. Initialize CREC client
    crecClient, err := client.NewCRECClient(&client.ClientOptions{
        BaseURL: "https://api.crec.chainlink.com",
        APIKey:  "your-api-key",
    })
    if err != nil {
        log.Fatal(err)
    }

    // 2. Create services
    channelsService, err := channels.NewService(&channels.ServiceOptions{
        CRECClient: crecClient,
    })
    if err != nil {
        log.Fatal(err)
    }

    watchersService, err := watchers.NewService(&watchers.ServiceOptions{
        CRECClient: crecClient,
    })
    if err != nil {
        log.Fatal(err)
    }

    // 3. Create a channel
    channel, err := channelsService.CreateChannel(ctx, channels.CreateChannelInput{
        Name: "dvp-monitoring-channel",
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("✓ Created channel: %s\n", channel.ChannelId)

    // 4. Create a watcher with domain
    name := "DVP Operations Watcher"
    watcher, err := watchersService.CreateWatcherWithDomain(ctx, channel.ChannelId, watchers.CreateWatcherWithDomainInput{
        Name:          &name,
        ChainSelector: 1, // Ethereum mainnet
        Address:       "0x1234567890123456789012345678901234567890",
        Domain:        "dvp",
        Events:        []string{"OperationExecuted"},
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("✓ Created watcher: %s (Status: %s)\n", watcher.WatcherId, watcher.Status)

    // 5. Wait for watcher to become active
    fmt.Println("\nWaiting for watcher deployment...")
    activeWatcher, err := watchersService.WaitForActive(ctx, channel.ChannelId, watcher.WatcherId, 2*time.Minute)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("✓ Watcher is now active: %s\n", activeWatcher.WatcherId)

    // 6. Update watcher name
    updatedWatcher, err := watchersService.UpdateWatcher(ctx, channel.ChannelId, watcher.WatcherId, watchers.UpdateWatcherInput{
        Name: "Production DVP Watcher",
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("✓ Updated watcher name to: %s\n", *updatedWatcher.Name)

    // 7. List all active watchers
    fmt.Println("\nAll active watchers:")
    status := watchers.StatusActive
    watchersList, err := watchersService.FindWatchersByChannel(ctx, channel.ChannelId, watchers.WatcherFilters{
        Status: &status,
    })
    if err != nil {
        log.Fatal(err)
    }

    for _, w := range watchersList.Data {
        fmt.Printf("  - %s: %s (Address: %s)\n", w.WatcherId, w.Status, w.Address)
    }

    // 8. Delete the watcher (cleanup) and wait for completion
    fmt.Println("\nCleaning up...")
    err = watchersService.DeleteWatcher(ctx, channel.ChannelId, watcher.WatcherId)
    if err != nil {
        log.Fatal(err)
    }
    
    // Wait for deletion to complete (handles async deletions)
    err = watchersService.WaitForDeleted(ctx, channel.ChannelId, watcher.WatcherId, 30*time.Second)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("✓ Watcher fully deleted: %s\n", watcher.WatcherId)
}
```

### Creating a Custom ERC20 Watcher

```go
func createERC20Watcher(ctx context.Context, service *watchers.Service, channelID uuid.UUID, tokenAddress string) (*apiClient.Watcher, error) {
    name := "ERC20 Token Watcher"
    
    watcher, err := service.CreateWatcherWithABI(ctx, channelID, watchers.CreateWatcherWithABIInput{
        Name:          &name,
        ChainSelector: 1, // Ethereum mainnet
        Address:       tokenAddress,
        Events:        []string{"Transfer", "Approval"},
        ABI: []watchers.EventABI{
            {
                Name: "Transfer",
                Type: "event",
                Inputs: []watchers.EventABIInput{
                    {Name: "from", Type: "address", InternalType: "address", Indexed: true},
                    {Name: "to", Type: "address", InternalType: "address", Indexed: true},
                    {Name: "value", Type: "uint256", InternalType: "uint256", Indexed: false},
                },
            },
            {
                Name: "Approval",
                Type: "event",
                Inputs: []watchers.EventABIInput{
                    {Name: "owner", Type: "address", InternalType: "address", Indexed: true},
                    {Name: "spender", Type: "address", InternalType: "address", Indexed: true},
                    {Name: "value", Type: "uint256", InternalType: "uint256", Indexed: false},
                },
            },
        },
    })
    if err != nil {
        return nil, err
    }

    // Wait for activation
    return service.WaitForActive(ctx, channelID, watcher.WatcherId, 2*time.Minute)
}
```

### Pagination Example

```go
func listAllWatchers(ctx context.Context, service *watchers.Service, channelID uuid.UUID) ([]apiClient.Watcher, error) {
    var allWatchers []apiClient.Watcher
    limit := 20
    offset := 0

    for {
        watchersList, err := service.FindWatchersByChannel(ctx, channelID, watchers.WatcherFilters{
            Limit:  &limit,
            Offset: &offset,
        })
        if err != nil {
            return nil, err
        }

        allWatchers = append(allWatchers, watchersList.Data...)

        if !watchersList.HasMore {
            break
        }

        offset += limit
    }

    return allWatchers, nil
}
```

### Filtering Watchers by Event

```go
func findWatchersMonitoringEvent(ctx context.Context, service *watchers.Service, channelID uuid.UUID, eventName string) ([]apiClient.Watcher, error) {
    watchersList, err := service.FindWatchersByChannel(ctx, channelID, watchers.WatcherFilters{
        EventName: &eventName,
    })
    if err != nil {
        return nil, err
    }

    return watchersList.Data, nil
}

// Usage
transferWatchers, err := findWatchersMonitoringEvent(ctx, watchersService, channelID, "Transfer")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Found %d watchers monitoring Transfer events\n", len(transferWatchers))
```

## Error Handling

The Watchers Service provides typed sentinel errors for robust error handling using Go's `errors.Is()` pattern.

### Typed Errors

The service defines the following sentinel errors that can be checked using `errors.Is()`:

#### Validation Errors
| Error | Description |
|-------|-------------|
| `ErrChannelIDRequired` | Channel ID parameter is empty or nil |
| `ErrWatcherIDRequired` | Watcher ID parameter is empty or nil |
| `ErrNameRequired` | Name field is required but not provided |
| `ErrDomainRequired` | Domain field is required but not provided |
| `ErrAddressRequired` | Contract address is required but not provided |
| `ErrEventsRequired` | Events list is empty |
| `ErrABIRequired` | ABI list is empty for ABI-based watcher |
| `ErrServiceOptionsRequired` | ServiceOptions is nil when creating service |
| `ErrCRECClientRequired` | CRECClient is nil in ServiceOptions |
| `ErrChainSelectorRequired` | ChainSelector is 0 (invalid) |
| `ErrInvalidABIType` | ABI Type is not "event" (only event types supported) |

#### Watcher State Errors
| Error | Description |
|-------|-------------|
| `ErrWatcherNotFound` | Watcher with specified ID doesn't exist (404) |
| `ErrWatcherDeploymentFailed` | Watcher deployment failed (status = "failed") |
| `ErrWatcherIsDeleting` | Watcher is being deleted and cannot become active |
| `ErrWatcherAlreadyDeleted` | Watcher has been deleted |
| `ErrWatcherDeletionFailed` | Watcher deletion process failed |

#### Timeout Errors
| Error | Description |
|-------|-------------|
| `ErrWaitForActiveTimeout` | Watcher didn't activate within timeout period |
| `ErrWaitForDeletedTimeout` | Watcher wasn't deleted within timeout period |

#### API Response Errors
| Error | Description |
|-------|-------------|
| `ErrEmptyResponse` | API returned an unexpected empty response |
| `ErrUnexpectedStatus` | Watcher is in an unexpected status |
| `ErrUnexpectedStatusCode` | API returned an unexpected HTTP status code |

#### API Operation Errors (base errors for wrapping)
| Error | Description |
|-------|-------------|
| `ErrCreateWatcherRequest` | Failed to create watcher request payload |
| `ErrCreateWatcherDomain` | Failed to create domain-based watcher |
| `ErrCreateWatcherABI` | Failed to create ABI-based watcher |
| `ErrFindWatchers` | Failed to find/list watchers |
| `ErrFindWatcher` | Failed to find specific watcher by ID |
| `ErrUpdateWatcher` | Failed to update watcher |
| `ErrDeleteWatcher` | Failed to delete watcher |
| `ErrCheckWatcherStatus` | Failed to check watcher status during polling |

### Error Handling Best Practices

1. **Use `errors.Is()` for typed error checking**: Check for specific error types using `errors.Is()`
2. **Always check for errors**: Never ignore error returns
3. **Log errors with context**: Include watcher IDs and channel IDs in error logs
4. **Handle deployment failures**: Check watcher status after creation
5. **Implement timeouts**: Use reasonable timeouts when waiting for activation
6. **Validate inputs**: The SDK validates inputs, but pre-validation can improve UX

#### Example: Checking for Specific Errors

```go
import (
    "errors"
    "github.com/smartcontractkit/crec-sdk/services/watchers"
)

// Check for not found errors
watcher, err := watchersService.FindWatcherByID(ctx, channelID, watcherID)
if err != nil {
    if errors.Is(err, watchers.ErrWatcherNotFound) {
        // Handle not found case gracefully
        log.Warn().Str("watcher_id", watcherID.String()).Msg("Watcher not found")
        return nil
    }
    // Handle other errors
    log.Error().Err(err).Msg("Failed to find watcher")
    return err
}

// Check for deployment failures
activeWatcher, err := watchersService.WaitForActive(ctx, channelID, watcher.WatcherId, 5*time.Minute)
if err != nil {
    if errors.Is(err, watchers.ErrWatcherDeploymentFailed) {
        // Handle deployment failure
        log.Error().Str("watcher_id", watcher.WatcherId.String()).Msg("Watcher deployment failed")
        // Maybe delete the failed watcher
        _ = watchersService.DeleteWatcher(ctx, channelID, watcher.WatcherId)
        return err
    }
    if errors.Is(err, watchers.ErrWaitForActiveTimeout) {
        // Handle timeout
        log.Warn().Str("watcher_id", watcher.WatcherId.String()).Msg("Timeout waiting for watcher")
        return err
    }
    return err
}

// Check for validation errors
watcher, err = watchersService.CreateWatcherWithDomain(ctx, uuid.Nil, input)
if err != nil {
    if errors.Is(err, watchers.ErrChannelIDRequired) {
        // Handle missing channel ID
        log.Error().Msg("Channel ID is required")
        return err
    }
    return err
}
```

## Watcher Lifecycle

Watchers go through several states during their lifecycle:

| State | Description | Next States |
|-------|-------------|-------------|
| **Pending** | Watcher has been created and is being deployed | → Active (success)<br>→ Failed (error) |
| **Active** | Watcher is deployed and monitoring events | → Deleting (on delete) |
| **Failed** | Watcher deployment failed | Terminal state |
| **Deleting** | Watcher is being removed | → Deleted |
| **Deleted** | Watcher has been removed | Terminal state |


## Integration with Other CREC Services

Watchers integrate with other CREC services to provide end-to-end blockchain monitoring:

1. **Channels**: Watchers are created within channels for organization
2. **Events**: Detected blockchain events are stored and accessible via the Events service
3. **Operations**: Watchers can trigger operation workflows based on detected events

Example integration:

```go
// 1. Create a channel
channel, _ := channelsService.CreateChannel(ctx, channels.CreateChannelInput{
    Name: "dvp-settlements",
})

// 2. Create a watcher in the channel
name := "DVP Watcher"
watcher, _ := watchersService.CreateWatcherWithDomain(ctx, channel.ChannelId, watchers.CreateWatcherWithDomainInput{
    Name:          &name,
    ChainSelector: 1,
    Address:       dvpContractAddress,
    Domain:        "dvp",
    Events:        []string{"OperationExecuted"},
})

// 3. Wait for activation
activeWatcher, _ := watchersService.WaitForActive(ctx, channel.ChannelId, watcher.WatcherId, 2*time.Minute)

// 4. Query events detected by the watcher (using events service)
events, _ := eventsService.ListEvents(ctx, eventsInput)

// 5. Trigger operations based on events (using operations service)
for _, event := range events {
    operation, _ := operationsService.CreateOperation(ctx, operationInput)
}
```

## Best Practices

1. **Use Domains When Possible**: For DvP, DTA, and test consumer contracts, use domain-based creation for simplicity and correctness

2. **Wait for Activation**: Always wait for watchers to become active before assuming they're monitoring events

3. **Handle Deployment Failures**: Check watcher status and handle failures gracefully

4. **Avoid Duplicate Events**: The API prevents creating multiple watchers for the same event on the same address within a channel

5. **Set Reasonable Timeouts**: When using `WaitForActive`, set a timeout appropriate for your use case (typically 1-2 minutes)

6. **Use Descriptive Names**: Give watchers meaningful names to help identify them later

7. **Filter Appropriately**: Use filters when listing watchers to improve query performance

8. **Monitor Watcher Health**: Regularly check watcher status to ensure they remain active

9. **Organize by Channel**: Group related watchers in the same channel for better organization

10. **Validate ABIs**: When providing custom ABIs, ensure all required fields are populated and correctly typed

