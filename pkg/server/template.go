package server

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/gorilla/mux"
	"github.com/huandu/go-sqlbuilder"
)

// We'll just interpret a subset of the mapping api for now
type CreateTemplateRequest struct {
	IndexPatterns string                 `json:"index_patterns"`
	Settings      map[string]interface{} `json:"settings,omitempty"`
	Mappings      Mappings               `json:"mappings,omitempty"`
}

type Mappings struct {
	DynamicTemplates []map[string]interface{} `json:"dynamic_templates"`
	Properties       map[string]Property      `json:"properties"`
}

type Property struct {
	Type        string `json:"type"`
	IgnoreAbove int    `json:"ignore_above"`
	Format      string `json:"format"`
}
type CreateTemplateResponse struct {
	Acknowledged bool `json:"acknowledged"`
}

type TemplateMapping struct {
	IndexPatterns string              `json:"index_patterns"`
	Fields        map[string]Property `json:"properties"`
}

func makeTemplateMapping() TemplateMapping {
	t := TemplateMapping{}
	t.Fields = make(map[string]Property)
	return t
}

// TODO Elasticsearch supports patterns like *-my-idx-*, which golang's RE
// engine doesn't accept. Quick and dirty transform all '*' to '.*',
// look into a proper solution
func cleanseIndexPattern(idx string) string {
	return strings.ReplaceAll(idx, "*", ".*")
}

func (s *Server) CreateTemplateHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	target, ok := vars["target"]
	if !ok {
		handleErrorResponse(w, errors.New("no target provided"))
		return
	}

	buf, err := io.ReadAll(r.Body)

	req := CreateTemplateRequest{}
	err = json.Unmarshal(buf, &req)
	req.IndexPatterns = cleanseIndexPattern(req.IndexPatterns)

	if err != nil {
		handleErrorResponse(w, errors.New("unable to parse json "+err.Error()))
		return
	}

	tm := createTemplateMappingForReq(req)
	s.TemplateMappings[target] = tm
	s.saveTemplateMetadata()

	resp := &CreateTemplateResponse{
		Acknowledged: true,
	}
	j, _ := json.Marshal(resp)

	w.Header().Set("Content-Type", "application/json")
	w.Write(j)
}

func createTemplateMappingForReq(req CreateTemplateRequest) TemplateMapping {
	tm := makeTemplateMapping()
	// FIXME Look into how ES style template patterns like *-idx-* can be made to work nicely with golang's RE package
	// Can replace all `*` with `.*` but there may be a better way
	tm.IndexPatterns = req.IndexPatterns
	for fld, prop := range req.Mappings.Properties {
		// We're only interested in date types for now
		if prop.Type == "date" {
			tm.Fields[fld] = prop
		}
	}
	return tm
}

func (s *Server) createMetadata() {
	// Quick and dirty way to persist the templates we need to keep track of.
	// Eventually there are other things we'll likely need to
	sb := sqlbuilder.NewCreateTableBuilder()
	sb.CreateTable("__templates").IfNotExists()
	sb.Define("target", "text")
	sb.Define("index_pattern", "text")
	sb.Define("body", "text")

	_, err := s.db.Exec(sb.String())
	if err != nil {
		panic(err)
	}
}

func (s *Server) loadTemplateMetadata() {
	s.TemplateMappings = make(map[string]TemplateMapping, 0)
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select("target", "index_pattern", "body").From("__templates")

	rows, err := s.db.Queryx(sb.String())

	if err != nil {
		panic(err)
	}

	for rows.Next() {
		tm := makeTemplateMapping()
		var flds string
		var target string
		rows.Scan(&target, &tm.IndexPatterns, &flds)
		err := json.Unmarshal([]byte(flds), &tm.Fields)
		if err != nil {
			panic(err)
		}
		s.TemplateMappings[target] = tm
	}

	if s.Cfg.Debug {
		log.Printf("Loaded %d templates from local datastore\n", len(s.TemplateMappings))
	}
}

func (s *Server) saveTemplateMetadata() {
	tx, _ := s.db.Begin()
	tx.Exec("DELETE FROM __templates;")
	qry := `INSERT INTO __templates (target, index_pattern, body) VALUES (?, ?, json(?))`

	for targ, tpl := range s.TemplateMappings {
		b, err := json.Marshal(tpl.Fields)
		if err != nil {
			panic(err)
		}
		sqlr, err := tx.Exec(qry, targ, tpl.IndexPatterns, string(b))
		if err != nil {
			panic(sqlr)
		}

	}
	tx.Commit()
}

func (s *Server) findMatchingTemplate(index string) *TemplateMapping {
	// This is not efficient, but will do for now given how few of these
	// there will likely to be. Iterate over all Template mappings
	// And check if the regex matches any

	for _, tm := range s.TemplateMappings {
		match, err := regexp.MatchString(tm.IndexPatterns, index)
		if err != nil {
			// Probably a regexp we haven't properly cleansed
			panic(err)
		}
		if match {
			return &tm
		}
	}
	return nil
}

func (s *Server) GetMappingDefinitionHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	target, ok := vars["target"]

	var targMappings map[string]TemplateMapping
	if ok {
		templ := s.findMatchingTemplate(target)
		targMappings = make(map[string]TemplateMapping, 0)
		if templ != nil {
			targMappings[target] = *templ
		}
	} else {
		targMappings = s.TemplateMappings
	}

	j, _ := json.Marshal(targMappings)

	w.Header().Set("Content-Type", "application/json")
	w.Write(j)
}
