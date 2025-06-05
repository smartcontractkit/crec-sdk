package types

type VerifiableEvent struct {
	Type       string   `json:"type"`
	Payload    string   `json:"verifiable_event"`
	Report     string   `json:"ocr_report"`
	Context    string   `json:"ocr_context"`
	Signatures []string `json:"signatures"`
}
