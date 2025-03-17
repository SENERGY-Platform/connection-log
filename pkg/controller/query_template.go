package controller

import (
	"bytes"
	"text/template"
	"time"
)

const (
	statePrevStr            = `SELECT "time", LAST("connected") AS "connected" FROM "{{.Kind}}" WHERE time < '{{.Timestamp}}' AND ({{range $index, $element := .Id}} {{if $index}} OR {{end}} "{{$.Kind}}" = '{{$element}}' {{end}}) GROUP BY "{{.Kind}}";`
	stateNextStr            = `SELECT "time", FIRST("connected") AS "connected" FROM "{{.Kind}}" WHERE time > '{{.Timestamp}}' AND ({{range $index, $element := .Id}} {{if $index}} OR {{end}} "{{$.Kind}}" = '{{$element}}' {{end}}) GROUP BY "{{.Kind}}";`
	statesTimeGrtEqStr      = `SELECT "time", "connected" FROM "{{.Kind}}" WHERE time >= '{{.Timestamp}}' AND ({{range $index, $element := .Id}} {{if $index}} OR {{end}} "{{$.Kind}}" = '{{$element}}' {{end}}) GROUP BY "{{.Kind}}";`
	statesTimeLesEqStr      = `SELECT "time", "connected" FROM "{{.Kind}}" WHERE time <= '{{.Timestamp}}' AND ({{range $index, $element := .Id}} {{if $index}} OR {{end}} "{{$.Kind}}" = '{{$element}}' {{end}}) GROUP BY "{{.Kind}}";`
	statesTimeGrtEqLesEqStr = `SELECT "time", "connected" FROM "{{.Kind}}" WHERE time >= '{{.TimestampA}}' AND time <= '{{.TimestampB}}' AND ({{range $index, $element := .Id}} {{if $index}} OR {{end}} "{{$.Kind}}" = '{{$element}}' {{end}}) GROUP BY "{{.Kind}}";`
)

type queryTemplates struct {
	statePrev            *template.Template
	stateNext            *template.Template
	statesTimeGrtEq      *template.Template
	statesTimeLesEq      *template.Template
	statesTimeGrtEqLesEq *template.Template
}

func newQueryTemplates() (*queryTemplates, error) {
	var err error
	var qt queryTemplates
	qt.statePrev, err = template.New("statePrev").Parse(statePrevStr)
	if err != nil {
		return nil, err
	}
	qt.stateNext, err = template.New("stateNext").Parse(stateNextStr)
	if err != nil {
		return nil, err
	}
	qt.statesTimeGrtEq, err = template.New("statesTimeGrtEq").Parse(statesTimeGrtEqStr)
	if err != nil {
		return nil, err
	}
	qt.statesTimeLesEq, err = template.New("statesTimeLesEq").Parse(statesTimeLesEqStr)
	if err != nil {
		return nil, err
	}
	qt.statesTimeGrtEqLesEq, err = template.New("statesTimeGrtEqLesEq").Parse(statesTimeGrtEqLesEqStr)
	if err != nil {
		return nil, err
	}
	return &qt, nil
}

func (t *queryTemplates) StatePrevQuery(ids []string, kind string, timestamp time.Time) (string, error) {
	return execTemplate(t.statePrev, map[string]any{"Id": ids, "Kind": kind, "Timestamp": formatTimestamp(timestamp)})
}

func (t *queryTemplates) StateNextQuery(ids []string, kind string, timestamp time.Time) (string, error) {
	return execTemplate(t.stateNext, map[string]any{"Id": ids, "Kind": kind, "Timestamp": formatTimestamp(timestamp)})
}

func (t *queryTemplates) StatesTimeGrtEqQuery(ids []string, kind string, timestamp time.Time) (string, error) {
	return execTemplate(t.statesTimeGrtEq, map[string]any{"Id": ids, "Kind": kind, "Timestamp": formatTimestamp(timestamp)})
}

func (t *queryTemplates) StatesTimeLesEqQuery(ids []string, kind string, timestamp time.Time) (string, error) {
	return execTemplate(t.statesTimeLesEq, map[string]any{"Id": ids, "Kind": kind, "Timestamp": formatTimestamp(timestamp)})
}

func (t *queryTemplates) StatesTimeGrtEqLesEqQuery(ids []string, kind string, timestampA, timestampB time.Time) (string, error) {
	return execTemplate(t.statesTimeGrtEqLesEq, map[string]any{"Id": ids, "Kind": kind, "TimestampA": formatTimestamp(timestampA), "TimestampB": formatTimestamp(timestampB)})
}

func execTemplate(t *template.Template, data any) (string, error) {
	var buffer bytes.Buffer
	err := t.Execute(&buffer, data)
	if err != nil {
		return "", err
	}
	return buffer.String(), nil
}

func formatTimestamp(t time.Time) string {
	return t.Format(time.RFC3339)
}
