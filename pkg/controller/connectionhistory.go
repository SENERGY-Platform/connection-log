/*
 * Copyright 2018 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package controller

import (
	"bytes"
	"encoding/json"
	"errors"
	influx "github.com/influxdata/influxdb1-client/v2"
	"log"
	"reflect"
	"text/template"
)


func parseTemplate(tmplName string, tmplString string, values interface{}) (result string, err error) {
	tmpl, err := template.New(tmplName).Parse(tmplString)
	if err != nil {
		return result, err
	}
	var buffer bytes.Buffer
	err = tmpl.Execute(&buffer, values)
	result = buffer.String()
	return
}

//duration in influxdb format https://docs.influxdata.com/influxdb/v1.5/query_language/spec/#durations
func (this *Controller) GetResourceHistory(id string, kind string, duration string) (result interface{}, err error) {
	templString := `SELECT * FROM "{{.Kind}}" WHERE time > now() - {{.Duration}} AND  "{{.Kind}}" = '{{.Id}}'`
	query, err := parseTemplate("getResourceHistory", templString, map[string]string{"Id": id, "Kind": kind, "Duration": duration})
	if err != nil {
		return result, err
	}
	q := influx.NewQuery(query, this.config.InfluxdbDb, "s")
	resp, err := this.influx.Query(q)
	if err != nil {
		log.Println("ERROR:", err, query)
		return result, err
	}
	err = resp.Error()
	if err != nil {
		log.Println("ERROR:", err, query)
	}
	return resp.Results, err
}

//duration in influxdb format https://docs.influxdata.com/influxdb/v1.5/query_language/spec/#durations
func (this *Controller) GetResourcesHistory(ids []string, kind string, duration string) (result interface{}, err error) {
	if len(ids) == 0 {
		return []interface{}{map[string]interface{}{"Series": []interface{}{}}}, nil
	}
	templString := `SELECT * FROM "{{.Kind}}" WHERE time > now() - {{.Duration}} AND ({{range $index, $element := .Id}} {{if $index}} OR {{end}} "{{$.Kind}}" = '{{$element}}' {{end}})  GROUP BY "{{.Kind}}"`
	query, err := parseTemplate("getResourcesHistory", templString, map[string]interface{}{"Id": ids, "Kind": kind, "Duration": duration})
	if err != nil {
		return result, err
	}
	q := influx.NewQuery(query, this.config.InfluxdbDb, "s")
	resp, err := this.influx.Query(q)
	if err != nil {
		log.Println("ERROR:", err, query)
		return result, err
	}
	err = resp.Error()
	if err != nil {
		log.Println("ERROR:", err, query)
	}
	return resp.Results, err
}

type HistoryResult struct {
	Series []HistorySeries `json:"Series"`
}

type HistorySeries struct {
	Name    string            `json:"name"`
	Tags    map[string]string `json:"tags"`
	Columns []string          `json:"columns"`
	Values  [][]interface{}   `json:"values"`
}

func (this *Controller) GetResourcesLogstart(ids []string, kind string) (result map[string]float64, err error) {
	result = map[string]float64{}
	templString := `SELECT FIRST(*) FROM "{{.Kind}}" WHERE {{range $index, $element := .Id}} {{if $index}} OR {{end}} "{{$.Kind}}" = '{{$element}}' {{end}} GROUP BY "{{.Kind}}"`
	query, err := parseTemplate("getResourcesHistory", templString, map[string]interface{}{"Id": ids, "Kind": kind})
	if err != nil {
		return result, err
	}
	q := influx.NewQuery(query, this.config.InfluxdbDb, "s")
	resp, err := this.influx.Query(q)
	if err != nil {
		return result, err
	}
	err = resp.Error()
	if err != nil {
		return result, err
	}
	b, _ := json.Marshal(resp.Results)
	temp := []HistoryResult{}
	json.Unmarshal(b, &temp)
	if len(temp) == 0 {
		err = errors.New("error while interpreting database result (series)")
		return
	}
	for _, series := range temp[0].Series {
		if len(series.Values) == 0 || len(series.Values[0]) == 0 {
			err = errors.New("error while interpreting database result (series.Values)")
			return
		}
		var ok bool
		result[series.Tags[kind]], ok = series.Values[0][0].(float64)
		if !ok {
			err = errors.New("error while interpreting database result (cast)" + reflect.TypeOf(series.Values[0][0]).String())
			return
		}
	}
	return
}

func (this *Controller) GetResourcesLogEdge(ids []string, kind string, duration string) (result map[string]interface{}, err error) {
	if len(ids) == 0 {
		return map[string]interface{}{}, nil
	}
	result = map[string]interface{}{}
	templString := `SELECT LAST(*) FROM "{{.Kind}}" WHERE time < now() - {{.Duration}} AND ({{range $index, $element := .Id}} {{if $index}} OR {{end}} "{{$.Kind}}" = '{{$element}}' {{end}})  GROUP BY "{{.Kind}}"`
	query, err := parseTemplate("getResourcesHistory", templString, map[string]interface{}{"Id": ids, "Kind": kind, "Duration": duration})
	if err != nil {
		return result, err
	}
	q := influx.NewQuery(query, this.config.InfluxdbDb, "s")
	resp, err := this.influx.Query(q)
	if err != nil {
		return result, err
	}
	err = resp.Error()
	if err != nil {
		return result, err
	}
	b, _ := json.Marshal(resp.Results)
	temp := []HistoryResult{}
	json.Unmarshal(b, &temp)
	if len(temp) == 0 {
		err = errors.New("error while interpreting database result (series)")
		return
	}
	for _, series := range temp[0].Series {
		if len(series.Values) == 0 {
			err = errors.New("error while interpreting database result (series.Values)")
			return
		}
		result[series.Tags[kind]] = series.Values[0]
	}
	return
}

//duration in influxdb format https://docs.influxdata.com/influxdb/v1.5/query_language/spec/#durations
func (this *Controller) GetResourceKindHistory(kind string, duration string) (result interface{}, err error) {
	templString := `SELECT * FROM "{{.Kind}}" WHERE time > now() - {{.Duration}} GROUP BY "{{.Kind}}"`
	query, err := parseTemplate("getResourceKindHistory", templString, map[string]string{"Kind": kind, "Duration": duration})
	if err != nil {
		return result, err
	}
	q := influx.NewQuery(query, this.config.InfluxdbDb, "s")
	resp, err := this.influx.Query(q)
	if err != nil {
		return result, err
	}
	err = resp.Error()
	return resp.Results, err
}
