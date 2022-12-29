package server

import (
	"fmt"
	"net/http"
)

type GenericErrorResponse struct {
	Type         bool   `json:"shards_acknowledged"`
	Reason       string `json:"reason"`
	ResourceType string `json:"resource.type"`
	ResourceId   string `json:"resource.id"`
	Index        string `json:"index"`
}

func handleErrorResponse(w http.ResponseWriter, err error) {
	// Generic error handler
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprint(w, err.Error())
}
