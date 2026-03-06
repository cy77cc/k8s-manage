package tools

import "testing"

func TestValidateResolvedParamsMissingAndEnumHints(t *testing.T) {
	meta := ToolMeta{
		Name:        "k8s_query",
		Required:    []string{"resource"},
		EnumSources: map[string]string{"resource": "cluster_list_inventory"},
		ParamHints:  map[string]string{"resource": "use pod/service/deployment"},
		Schema: map[string]any{
			"properties": map[string]any{
				"resource": map[string]any{"type": "string", "enum": []any{"pod", "service"}},
			},
		},
	}
	err := validateResolvedParams(meta, map[string]any{})
	if err == nil || err.Error() == "" {
		t.Fatalf("expected missing param error")
	}
	if err.Error() != "missing required parameter `resource`; you can call `cluster_list_inventory` to get candidate values; use pod/service/deployment" {
		t.Fatalf("unexpected missing param message: %q", err.Error())
	}

	err = validateResolvedParams(meta, map[string]any{"resource": "node"})
	if err == nil || err.Error() == "" {
		t.Fatalf("expected enum validation error")
	}
}

func TestValidateTypeAndConversions(t *testing.T) {
	cases := []struct {
		name  string
		field string
		value any
		prop  map[string]any
		ok    bool
	}{
		{name: "integer ok", field: "limit", value: "10", prop: map[string]any{"type": "integer"}, ok: true},
		{name: "integer bad", field: "limit", value: "ten", prop: map[string]any{"type": "integer"}, ok: false},
		{name: "number ok", field: "ratio", value: "3.14", prop: map[string]any{"type": "number"}, ok: true},
		{name: "boolean string ok", field: "force", value: "true", prop: map[string]any{"type": "boolean"}, ok: true},
		{name: "array ok", field: "ids", value: []string{"1", "2"}, prop: map[string]any{"type": "array"}, ok: true},
		{name: "array bad", field: "ids", value: "1,2", prop: map[string]any{"type": "array"}, ok: false},
		{name: "string bad", field: "target", value: 12, prop: map[string]any{"type": "string"}, ok: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateType(tc.field, tc.value, tc.prop)
			if tc.ok && err != nil {
				t.Fatalf("expected success, got %v", err)
			}
			if !tc.ok && err == nil {
				t.Fatalf("expected validation error")
			}
		})
	}

	if v, ok := toInt64("42"); !ok || v != 42 {
		t.Fatalf("unexpected integer conversion: %v %v", v, ok)
	}
	if _, ok := toInt64("4.2"); ok {
		t.Fatalf("expected invalid integer conversion")
	}
	if v, ok := toFloat64("4.2"); !ok || v != 4.2 {
		t.Fatalf("unexpected float conversion: %v %v", v, ok)
	}
	if _, ok := toFloat64(""); ok {
		t.Fatalf("expected empty float conversion to fail")
	}
}
