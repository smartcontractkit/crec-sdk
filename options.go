package crec

import (
	"log/slog"
	"net/http"
	"time"
)

// DefaultMinRequiredSignatures is the minimum number of valid signatures required
// to verify an event. Calculated as F + 1, where F is the maximum number of faulty
// nodes the DON can tolerate. With F+1 signatures, at least one must be from an
// honest node, guaranteeing the data passed consensus.
//
// The DON transmits once this minimum is reached, so reports will never contain
// more than F+1 signatures. Do not increase this value unless you are certain
// about the implications—doing so will cause signature verification to fail.
const DefaultMinRequiredSignatures = 4

// DefaultValidSigners contains the production Zone A workflow node public keys.
// These are the keys used by the DON to sign events on Ethereum Mainnet.
// Updated keys can be found in the chainlink-deployments repository.
//
// These keys rarely change. When they do, update the SDK to get new defaults.
var DefaultValidSigners = []string{
	"0xff9b062fccb2f042311343048b9518068370f837", // chainlayer-wf-zone-a-1
	"0xe55fcaf921e76c6bbcf9415bba12b1236f07b0c3", // clp-cre-wf-zone-a-0
	"0x4d6cfd44f94408a39fb1af94a53c107a730ba161", // dextrac-wf-zone-a-3
	"0xde5cd1dd4300a0b4854f8223add60d20e1dfe21b", // fiews-wf-zone-a-2
	"0xf3baa9a99b5ad64f50779f449bac83baac8bfdb6", // inotel-wf-zone-a-4
	"0xd7f22fb5382ff477d2ff5c702cab0ef8abf18233", // linkforest-wf-zone-a-5
	"0xcdf20f8ffd41b02c680988b20e68735cc8c1ca17", // linkpool-wf-zone-a-0
	"0x4d7d71c7e584cfa1f5c06275e5d283b9d3176924", // linkriver-wf-zone-a-6
	"0x1a89c98e75983ec384ad8e83eaf7d0176eeaf155", // piertwo-wf-zone-a-7
	"0x4f99b550623e77b807df7cbed9c79d55e1163b48", // simplyvc-wf-zone-a-8
}

// clientConfig holds the internal configuration for the Client.
type clientConfig struct {
	httpClient                       *http.Client
	logger                           *slog.Logger
	minRequiredSignatures            int
	validSigners                     []string
	disableEventVerification         bool
	orgID                            string
	workflowOwner                    string
	creTenantID                      string
	watcherPollInterval              time.Duration
	watcherEventualConsistencyWindow time.Duration
}

// Option is a functional option for configuring the Client.
type Option func(*clientConfig)

// WithHTTPClient sets a custom HTTP client for API requests.
// If not provided, http.DefaultClient is used.
func WithHTTPClient(client *http.Client) Option {
	return func(cfg *clientConfig) {
		cfg.httpClient = client
	}
}

// WithLogger sets a custom logger for the SDK.
// If not provided, slog.Default() is used.
func WithLogger(logger *slog.Logger) Option {
	return func(cfg *clientConfig) {
		cfg.logger = logger
	}
}

// WithEventVerification configures custom event verification settings.
// By default, the SDK uses DefaultValidSigners and DefaultMinRequiredSignatures.
// Use this option to override with custom keys or signature requirements.
//
// Parameters:
//   - minRequiredSignatures: Minimum number of valid signatures required to verify an event
//   - validSigners: List of valid signer addresses (as hex strings)
func WithEventVerification(minRequiredSignatures int, validSigners []string) Option {
	return func(cfg *clientConfig) {
		cfg.minRequiredSignatures = minRequiredSignatures
		cfg.validSigners = validSigners
	}
}

// WithoutEventVerification disables event verification entirely.
// Use this option if you don't need to verify event signatures.
func WithoutEventVerification() Option {
	return func(cfg *clientConfig) {
		cfg.disableEventVerification = true
	}
}

// WithOrgID sets the default organization ID for the events client.
// When set, [Client.Events].Verify and [Client.Events].VerifyOperationStatus can be called
// without passing an org ID. For multi-org use, omit this and use VerifyWithOrgID or
// VerifyOperationStatusWithOrgID with an explicit org ID per call.
func WithOrgID(orgID string) Option {
	return func(cfg *clientConfig) {
		cfg.orgID = orgID
	}
}

// WithWorkflowOwner sets the default workflow owner address for the events client.
// When set, [Client.Events].VerifyWithWorkflowOwner and [Client.Events].VerifyOperationStatusWithWorkflowOwner
// can be called without passing the workflow owner. For multi-org or per-event workflow owner, pass it explicitly.
func WithWorkflowOwner(workflowOwner string) Option {
	return func(cfg *clientConfig) {
		cfg.workflowOwner = workflowOwner
	}
}

// WithCRETenantID sets the CRE tenant ID referring to different environments of CRE
// for workflow owner address derivation. Defaults to events.CreMainlineTenantID ("1") if not provided.
func WithCRETenantID(creTenantID string) Option {
	return func(cfg *clientConfig) {
		cfg.creTenantID = creTenantID
	}
}

// WithWatcherPolling configures the Watchers client polling behavior.
//
// Parameters:
//   - pollInterval: Duration between polling attempts when waiting for watcher state changes
//   - eventualConsistencyWindow: Duration to tolerate 404 errors after creation due to eventual consistency
func WithWatcherPolling(pollInterval, eventualConsistencyWindow time.Duration) Option {
	return func(cfg *clientConfig) {
		cfg.watcherPollInterval = pollInterval
		cfg.watcherEventualConsistencyWindow = eventualConsistencyWindow
	}
}
