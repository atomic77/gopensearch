package server

import (
	"encoding/json"
	"io"
	"net/http"
	"regexp"

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
	IndexPatterns string
	Fields        map[string]Property
}

func makeTemplateMapping() TemplateMapping {
	t := TemplateMapping{}
	t.Fields = make(map[string]Property)
	return t
}

func (s *Server) CreateTemplateHandler(w http.ResponseWriter, r *http.Request) {
	buf, err := io.ReadAll(r.Body)

	req := CreateTemplateRequest{}
	err = json.Unmarshal(buf, &req)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	tm := createTemplateMappingForReq(req)
	s.TemplateMappings = append(s.TemplateMappings, tm)
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
	sb.Define("index_pattern", "text")
	sb.Define("body", "text")

	_, err := s.db.Exec(sb.String())
	if err != nil {
		panic(err)
	}
}

func (s *Server) loadTemplateMetadata() {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select("index_pattern", "body").From("__templates")

	rows, err := s.db.Queryx(sb.String())

	if err != nil {
		panic(err)
	}

	for rows.Next() {
		tm := makeTemplateMapping()
		var flds string
		rows.Scan(&tm.IndexPatterns, &flds)
		err := json.Unmarshal([]byte(flds), &tm.Fields)
		if err != nil {
			panic(err)
		}
		s.TemplateMappings = append(s.TemplateMappings, tm)
	}
}

func (s *Server) saveTemplateMetadata() {

	tx, _ := s.db.Begin()
	tx.Exec("DELETE FROM __templates;")
	qry := `INSERT INTO __templates (index_pattern, body) VALUES (?, json(?))`

	for _, tpl := range s.TemplateMappings {
		b, err := json.Marshal(tpl.Fields)
		if err != nil {
			panic(err)
		}
		sqlr, err := tx.Exec(qry, tpl.IndexPatterns, string(b))
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
		match, _ := regexp.MatchString(tm.IndexPatterns, index)
		if match {
			return &tm
		}
	}
	return nil
}
