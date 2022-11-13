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
	resp := &CreateTemplateResponse{
		Acknowledged: true,
	}
	j, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	w.Write(j)
}
