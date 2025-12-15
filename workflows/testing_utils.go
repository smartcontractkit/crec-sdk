package workflows

import (
	"testing"

	"github.com/smartcontractkit/cre-sdk-go/cre/testutils"
)

// PrepareTestingRuntime creates a test runtime with standard secrets configured.
func PrepareTestingRuntime(t *testing.T) *testutils.TestRuntime {
	return testutils.NewRuntime(t, testutils.Secrets{
		"": map[testutils.ID]string{
			"courier": "API-KEY",
		},
	})
}
