package server

import (
	"encoding/json"
	"net/http"
)

type CreateTemplateResponse struct {
	Acknowledged bool `json:"acknowledged"`
}

func (s *Server) CreateTemplateHandler(w http.ResponseWriter, r *http.Request) {
	/* TODO Implement me ; swallow the response for now and pretend everything
	is ok */
	// Jaeger example template
	// Will need at the very least to ensure that all date types are stored internally
	// in a consistent manner to make comparisons easier
	// https://github.com/jaegertracing/jaeger/blob/458c1ce7ef7e0a3d374711187d96514600689754/plugin/storage/es/mappings/jaeger-span.json
	resp := &CreateTemplateResponse{
		Acknowledged: true,
	}
	j, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	w.Write(j)
}
