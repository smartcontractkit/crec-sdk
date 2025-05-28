package integration

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"

	"github.com/smartcontractkit/cvm-sdk/events/listener"
)

const (
	KafkaTopicName = "beholder_otlp_logs"
	KafkaGroupId   = "cvn-sdk-ci"
)

func setupTestEnv(ctx context.Context) (
	string, *testcontainers.DockerNetwork, *testcontainers.Container, *testcontainers.Container,
) {
	fmt.Println("Setting up integration test environment")

	dockerNetwork := createDockerNetwork(ctx)
	rpCtr, rpSeedBroker := startRedpandaContainer(ctx, dockerNetwork)
	kcatCtr := startKcatContainer(ctx, dockerNetwork)

	sendKafkaEvents(ctx, kcatCtr, KafkaTopicName)

	return rpSeedBroker, dockerNetwork, &rpCtr, &kcatCtr
}

func teardownTestEnv(
	ctx context.Context, dockerNetwork *testcontainers.DockerNetwork, rpCtr testcontainers.Container,
	kcatCtr testcontainers.Container,
) {
	fmt.Println("Terminating integration test environment")

	err := rpCtr.Terminate(ctx)
	if err != nil {
		log.Printf("failed to terminate redpanda container: %s", err)
	}
	if err := kcatCtr.Terminate(ctx); err != nil {
		log.Printf("failed to terminate kcat container: %s", err)
	}
	if err := dockerNetwork.Remove(ctx); err != nil {
		log.Printf("failed to remove network: %s", err)
	}
}

func TestEventListener(t *testing.T) {
	ctx := context.Background()

	broker, dockerNetwork, rpCtr, kcatCtr := setupTestEnv(ctx)

	l := listener.NewEventListener(
		&listener.EventListenerOptions{
			Brokers: []string{broker},
			Topic:   KafkaTopicName,
			GroupID: KafkaGroupId,
		},
	)

	evt, err := l.Read()
	if err != nil {
		t.Fatalf("failed to read event: %v", err)
	}
	log.Printf("Event received: %v", evt)

	assert.Equal(t, evt.Type, "SettlementAccepted")

	teardownTestEnv(ctx, dockerNetwork, *rpCtr, *kcatCtr)
}
