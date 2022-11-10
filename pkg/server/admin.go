// Random administrative and health related apis
package server

import (
	"encoding/json"
	"net/http"
)

func (s *Server) HeadHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write(nil)
}

type VersionStatus struct {
	Number                           string `json:"number"`
	BuildFlavor                      string `json:"build_flavor"`
	MinimumIndexCompatibilityVersion string `json:"minimum_index_compatibility_version"`
	MinimumWireCompatibilityVersion  string `json:"minimum_wire_compatibility_version"`
}

type ClusterStatusResponse struct {
	Name        string         `json:"name"`
	ClusterName string         `json:"cluster_name"`
	ClusterUUID string         `json:"cluster_uuid"`
	Version     *VersionStatus `json:"version"`
	TagLine     string         `json:"tagline"`
}

func (s *Server) ClusterStatusHandler(w http.ResponseWriter, r *http.Request) {
	vs := &VersionStatus{
		Number:                           "7.17",
		BuildFlavor:                      "default",
		MinimumIndexCompatibilityVersion: "6.8.0",
		MinimumWireCompatibilityVersion:  "6.8.0",
	}
	cs := &ClusterStatusResponse{
		Name:        "asdfasdf",
		ClusterName: "qwerty",
		ClusterUUID: "asdf;ljkasdf",
		Version:     vs,
		TagLine:     "You Go, for search",
	}
	j, _ := json.Marshal(cs)
	w.Header().Set("Content-Type", "application/json")
	w.Write(j)
}
