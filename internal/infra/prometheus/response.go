package prometheus

import (
	"encoding/json"
	"fmt"
)

type queryEnvelope struct {
	Status    string          `json:"status"`
	Data      json.RawMessage `json:"data"`
	ErrorType string          `json:"errorType"`
	Error     string          `json:"error"`
}

type queryData struct {
	ResultType string        `json:"resultType"`
	Result     []VectorPoint `json:"result"`
}

type rangeData struct {
	ResultType string        `json:"resultType"`
	Result     []MatrixPoint `json:"result"`
}

type VectorPoint struct {
	Metric map[string]string `json:"metric"`
	Value  []any             `json:"value"`
}

type MatrixPoint struct {
	Metric map[string]string `json:"metric"`
	Values [][]any           `json:"values"`
}

type QueryResult struct {
	ResultType string
	Vector     []VectorPoint
	Matrix     []MatrixPoint
}

type MetadataItem struct {
	Metric string `json:"metric"`
	Type   string `json:"type"`
	Help   string `json:"help"`
	Unit   string `json:"unit"`
}

func parseQueryEnvelope(body []byte) (*queryEnvelope, error) {
	var env queryEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, err
	}
	if env.Status != "success" {
		if env.Error != "" {
			return nil, fmt.Errorf("prometheus %s: %s", env.ErrorType, env.Error)
		}
		return nil, fmt.Errorf("prometheus status: %s", env.Status)
	}
	return &env, nil
}
