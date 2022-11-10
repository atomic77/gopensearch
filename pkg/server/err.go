package server

type GenericErrorResponse struct {
	Type         bool   `json:"shards_acknowledged"`
	Reason       string `json:"reason"`
	ResourceType string `json:"resource.type"`
	ResourceId   string `json:"resource.id"`
	Index        string `json:"index"`
}
