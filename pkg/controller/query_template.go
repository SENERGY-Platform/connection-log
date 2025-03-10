package controller

import (
	"bytes"
	"text/template"
	"time"
)

const (
	resourcesStatesFirstStr = `SELECT "time", FIRST("connected") AS "connected" FROM "{{.Kind}}" WHERE {{range $index, $element := .Id}} {{if $index}} OR {{end}} "{{$.Kind}}" = '{{$element}}' {{end}} GROUP BY "{{.Kind}}";`
	resourcesStatesLastStr  = `SELECT "time", LAST("connected") AS "connected" FROM "{{.Kind}}" WHERE time < '{{.Timestamp}}' AND ({{range $index, $element := .Id}} {{if $index}} OR {{end}} "{{$.Kind}}" = '{{$element}}' {{end}})  GROUP BY "{{.Kind}}";`
	resourcesStatesRangeStr = `SELECT "time", "connected" FROM "{{.Kind}}" WHERE time > '{{.Timestamp}}' AND ({{range $index, $element := .Id}} {{if $index}} OR {{end}} "{{$.Kind}}" = '{{$element}}' {{end}})  GROUP BY "{{.Kind}}";`
)

type queryTemplates struct {
	resStatesFirst *template.Template
	resStatesLast  *template.Template
	resStatesRange *template.Template
}

func newQueryTemplates() (*queryTemplates, error) {
	var err error
	var qt queryTemplates
	qt.resStatesFirst, err = template.New("resStatesFirst").Parse(resourcesStatesFirstStr)
	if err != nil {
		return nil, err
	}
	qt.resStatesLast, err = template.New("resStatesLast").Parse(resourcesStatesLastStr)
	if err != nil {
		return nil, err
	}
	qt.resStatesRange, err = template.New("resStatesRange").Parse(resourcesStatesRangeStr)
	if err != nil {
		return nil, err
	}
	return &qt, nil
}

func (t *queryTemplates) GetResStatesInitQuery(ids []string, kind string) (string, error) {
	return execTemplate(t.resStatesFirst, map[string]any{"Id": ids, "Kind": kind})
}

func (t *queryTemplates) GetResStatesPrevQuery(ids []string, kind string, timestamp time.Time) (string, error) {
	return execTemplate(t.resStatesLast, map[string]any{"Id": ids, "Kind": kind, "Timestamp": timestamp.Format(time.RFC3339)})
}

func (t *queryTemplates) GetResStatesRangeQuery(ids []string, kind string, timestamp time.Time) (string, error) {
	return execTemplate(t.resStatesRange, map[string]any{"Id": ids, "Kind": kind, "Timestamp": timestamp.Format(time.RFC3339)})
}

func execTemplate(t *template.Template, data any) (string, error) {
	var buffer bytes.Buffer
	err := t.Execute(&buffer, data)
	if err != nil {
		return "", err
	}
	return buffer.String(), nil
}
