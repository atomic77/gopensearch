package server

import (
	"encoding/json"
	"regexp"
	"testing"

	require "github.com/alecthomas/assert/v2"
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

func TestRegexp(t *testing.T) {
	// ES and Golang regexps need to be made to play nicely
	pat := "*jaeger-service-*"
	regex := regexp.MustCompile(pat)
	s := "jaeger-service-2023-01-01"
	// match, _ := regexp.MatchString(,
	match := regex.FindString(s)

	require.Equal(t, match, s)

}
