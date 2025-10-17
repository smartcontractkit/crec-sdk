package events

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"

	apiClient "github.com/smartcontractkit/crec-api-go/client"
)

func LoadMockEvent(filename string) (*apiClient.Event, error) {
	var event apiClient.Event
	err := LoadJson(filename, &event)
	if err != nil {
		return nil, fmt.Errorf("failed to load mock event from %s: %w", filename, err)
	}
	return &event, nil
}

func LoadJson(filename string, target any) error {
	_, b, _, ok := runtime.Caller(0)
	if !ok {
		return errors.New("could not get caller information")
	}
	currentFileDir := filepath.Dir(b)

	file, err := os.Open(fmt.Sprintf("%s/data/%s", currentFileDir, filename))
	if err != nil {
		log.Fatal(err)
	}
	data, err := io.ReadAll(file)
	return json.Unmarshal(data, target)
}
