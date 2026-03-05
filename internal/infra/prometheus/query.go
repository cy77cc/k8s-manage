package prometheus

import (
	"fmt"
	"sort"
	"strings"
)

var allowedAgg = map[string]struct{}{
	"avg": {}, "sum": {}, "max": {}, "min": {}, "count": {}, "stddev": {}, "stdvar": {},
}

type QueryBuilder struct {
	metric    string
	labels    map[string]string
	rangeExpr string
	aggFunc   string
}

func NewQueryBuilder(metric string) *QueryBuilder {
	return &QueryBuilder{metric: strings.TrimSpace(metric), labels: make(map[string]string)}
}

func (b *QueryBuilder) WithLabel(key, value string) *QueryBuilder {
	k := strings.TrimSpace(key)
	if k == "" {
		return b
	}
	b.labels[k] = strings.TrimSpace(value)
	return b
}

func (b *QueryBuilder) WithRange(expr string) *QueryBuilder {
	b.rangeExpr = strings.TrimSpace(expr)
	return b
}

func (b *QueryBuilder) WithAggregation(agg string) *QueryBuilder {
	a := strings.ToLower(strings.TrimSpace(agg))
	if _, ok := allowedAgg[a]; ok {
		b.aggFunc = a
	}
	return b
}

func (b *QueryBuilder) Build() string {
	base := b.metric
	if len(b.labels) > 0 {
		keys := make([]string, 0, len(b.labels))
		for k := range b.labels {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		parts := make([]string, 0, len(keys))
		for _, k := range keys {
			parts = append(parts, fmt.Sprintf("%s=\"%s\"", k, b.labels[k]))
		}
		base = fmt.Sprintf("%s{%s}", base, strings.Join(parts, ","))
	}
	if b.rangeExpr != "" {
		base = fmt.Sprintf("%s[%s]", base, b.rangeExpr)
	}
	if b.aggFunc != "" {
		base = fmt.Sprintf("%s(%s)", b.aggFunc, base)
	}
	return base
}
