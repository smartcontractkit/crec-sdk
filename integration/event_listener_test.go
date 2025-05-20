package integration

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/testcontainers/testcontainers-go"

	"github.com/smartcontractkit/cvm-sdk/events"
)

const (
	KafkaTopicName = "beholder_otlp_logs"
	KafkaGroupId   = "cvm-example"
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

	listener := events.NewEventListener(
		&events.EventListenerOptions{
			Brokers: []string{broker},
			Topic:   KafkaTopicName,
			GroupID: KafkaGroupId,
		},
	)

	evt, err := listener.Read()
	if err != nil {
		t.Fatalf("failed to read event: %v", err)
	}
	log.Printf("Event received: %v", evt)

	teardownTestEnv(ctx, dockerNetwork, *rpCtr, *kcatCtr)
}
