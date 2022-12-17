package server

import (
	"encoding/json"
	"testing"

	"github.com/alecthomas/repr"
)

func TestTemplateParse(t *testing.T) {
	j := `
	{
		"index_patterns": "*jaeger-service-*",
		"mappings":{
		  "dynamic_templates": [{"m": {"mapping": {}}}],
		  "properties":{
				"serviceName":{
					"type":"keyword",
					"ignore_above":256
				},
				"operationName":{
					"type":"keyword",
					"ignore_above":256
				}
			}
		}
	  }
	`
	req := CreateTemplateRequest{}
	json.Unmarshal([]byte(j), &req)

	repr.Println(req)
}
