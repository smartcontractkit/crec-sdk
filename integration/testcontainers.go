package integration

import (
	"context"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/redpanda"
	"github.com/testcontainers/testcontainers-go/network"
)

func createDockerNetwork(ctx context.Context) *testcontainers.DockerNetwork {
	docketNetwork, err := network.New(ctx)
	if err != nil {
		log.Fatalf("Failed to create Docker network: %v", err)
	}

	log.Printf("Docker network created: %s", docketNetwork.Name)
	return docketNetwork
}

func startRedpandaContainer(ctx context.Context, dockerNetwork *testcontainers.DockerNetwork) (
	testcontainers.Container, string,
) {
	log.Printf("Starting RedPanda container")
	rpCtr, err := redpanda.Run(
		ctx, "redpandadata/redpanda:v23.2.18",
		network.WithNetwork([]string{"redpanda-host"}, dockerNetwork),
		redpanda.WithListener("redpanda:29092"),
		redpanda.WithAutoCreateTopics(),
	)
	if err != nil {
		log.Fatalf("failed to start container: %v", err)
	}

	seedBroker, err := rpCtr.KafkaSeedBroker(ctx)
	if err != nil {
		log.Fatalf("failed to get seed broker: %v", err)
	}

	log.Printf("RedPanda container started successfully, seed broker: %s", seedBroker)
	return rpCtr, seedBroker
}

func startKcatContainer(ctx context.Context, docketNetwork *testcontainers.DockerNetwork) testcontainers.Container {
	kcatCtr, err := testcontainers.GenericContainer(
		ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Image:      "confluentinc/cp-kcat:7.4.1",
				Entrypoint: []string{"sh"},
				Networks:   []string{docketNetwork.Name},
				Cmd:        []string{"-c", "tail -f /dev/null"},
			},
			Started: true,
		},
	)
	if err != nil {
		log.Fatalf("failed to start kcat container: %v", err)
	}

	err = kcatCtr.CopyDirToContainer(ctx, "kafka_events", "/", 0o755)
	if err != nil {
		log.Fatalf("failed to copy kafka_events directory to container: %v", err)
	}

	log.Printf("kcat container started successfully")
	return kcatCtr
}

func sendKafkaEvents(ctx context.Context, kcat testcontainers.Container, topic string) {
	kcat.Exec(ctx, []string{"/kafka_events/publish_events.sh", topic})
}
