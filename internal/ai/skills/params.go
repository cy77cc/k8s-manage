package skills

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var templatePattern = regexp.MustCompile(`\{\{\s*([^{}]+?)\s*\}\}`)

type EnumValueProvider interface {
	ListValues(source string) ([]string, error)
}

func validateParams(defs []SkillParameter, input map[string]any) (map[string]any, error) {
	params := make(map[string]any, len(defs))
	for k, v := range input {
		params[k] = v
	}
	for _, def := range defs {
		name := strings.TrimSpace(def.Name)
		if name == "" {
			continue
		}
		value, exists := params[name]
		if (!exists || value == nil || strings.TrimSpace(fmt.Sprintf("%v", value)) == "") && def.Default != nil {
			value = def.Default
			exists = true
		}
		if def.Required && (!exists || value == nil || strings.TrimSpace(fmt.Sprintf("%v", value)) == "") {
			return nil, fmt.Errorf("missing required parameter: %s", name)
		}
		if !exists || value == nil {
			continue
		}
		casted, err := castParamValue(strings.TrimSpace(def.Type), value)
		if err != nil {
			return nil, fmt.Errorf("invalid parameter %s: %w", name, err)
		}
		params[name] = casted
	}
	return params, nil
}

func extractParams(message string, defs []SkillParameter, enumProvider EnumValueProvider) (map[string]any, error) {
	parsed := parseInlineParams(message)
	out := make(map[string]any, len(defs))
	for _, def := range defs {
		name := strings.TrimSpace(def.Name)
		if name == "" {
			continue
		}
		raw, ok := parsed[name]
		if !ok {
			continue
		}
		value := any(raw)
		if strings.TrimSpace(def.EnumSource) != "" {
			resolved, err := resolveEnumValue(raw, def.EnumSource, enumProvider)
			if err != nil {
				return nil, fmt.Errorf("resolve enum source for %s: %w", name, err)
			}
			value = resolved
		}
		out[name] = value
	}
	return validateParams(defs, out)
}

func renderParamsTemplate(tpl map[string]any, params map[string]any, stepResults map[string]any) (map[string]any, error) {
	if len(tpl) == 0 {
		return map[string]any{}, nil
	}
	out := make(map[string]any, len(tpl))
	for k, v := range tpl {
		rendered, err := renderTemplateValue(v, params, stepResults)
		if err != nil {
			return nil, err
		}
		out[k] = rendered
	}
	return out, nil
}

func renderTemplateValue(value any, params map[string]any, stepResults map[string]any) (any, error) {
	switch node := value.(type) {
	case string:
		matches := templatePattern.FindAllStringSubmatch(node, -1)
		if len(matches) == 0 {
			return node, nil
		}
		if len(matches) == 1 && strings.TrimSpace(matches[0][0]) == strings.TrimSpace(node) {
			resolved, ok := resolveTemplateRef(matches[0][1], params, stepResults)
			if !ok {
				return nil, fmt.Errorf("template reference not found: %s", matches[0][1])
			}
			return resolved, nil
		}
		rendered := node
		for _, match := range matches {
			resolved, ok := resolveTemplateRef(match[1], params, stepResults)
			if !ok {
				return nil, fmt.Errorf("template reference not found: %s", match[1])
			}
			rendered = strings.ReplaceAll(rendered, match[0], fmt.Sprintf("%v", resolved))
		}
		return rendered, nil
	case map[string]any:
		child := make(map[string]any, len(node))
		for k, v := range node {
			rv, err := renderTemplateValue(v, params, stepResults)
			if err != nil {
				return nil, err
			}
			child[k] = rv
		}
		return child, nil
	case []any:
		child := make([]any, 0, len(node))
		for _, v := range node {
			rv, err := renderTemplateValue(v, params, stepResults)
			if err != nil {
				return nil, err
			}
			child = append(child, rv)
		}
		return child, nil
	default:
		return value, nil
	}
}

func resolveTemplateRef(ref string, params map[string]any, stepResults map[string]any) (any, bool) {
	ref = strings.TrimSpace(ref)
	if strings.HasPrefix(ref, "params.") {
		v, ok := pickPath(params, strings.TrimPrefix(ref, "params."))
		return v, ok
	}
	if strings.HasPrefix(ref, "steps.") {
		v, ok := pickPath(stepResults, strings.TrimPrefix(ref, "steps."))
		return v, ok
	}
	if v, ok := pickPath(params, ref); ok {
		return v, true
	}
	return pickPath(stepResults, ref)
}

func pickPath(root map[string]any, path string) (any, bool) {
	if root == nil {
		return nil, false
	}
	parts := strings.Split(strings.TrimSpace(path), ".")
	if len(parts) == 0 {
		return nil, false
	}
	var current any = root
	for _, part := range parts {
		key := strings.TrimSpace(part)
		if key == "" {
			return nil, false
		}
		m, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		value, exists := m[key]
		if !exists {
			return nil, false
		}
		current = value
	}
	return current, true
}

func parseInlineParams(message string) map[string]string {
	out := map[string]string{}
	for _, token := range strings.Fields(message) {
		k, v, ok := strings.Cut(token, "=")
		if !ok {
			continue
		}
		key := strings.TrimSpace(k)
		val := strings.TrimSpace(v)
		if key == "" || val == "" {
			continue
		}
		out[key] = strings.Trim(val, "\"'")
	}
	return out
}

func resolveEnumValue(input, source string, provider EnumValueProvider) (string, error) {
	if provider == nil || strings.TrimSpace(source) == "" {
		return input, nil
	}
	candidates, err := provider.ListValues(source)
	if err != nil {
		return "", err
	}
	needle := strings.ToLower(strings.TrimSpace(input))
	if needle == "" {
		return input, nil
	}
	for _, item := range candidates {
		v := strings.TrimSpace(item)
		if strings.EqualFold(v, input) {
			return v, nil
		}
	}
	for _, item := range candidates {
		v := strings.TrimSpace(item)
		if strings.Contains(strings.ToLower(v), needle) {
			return v, nil
		}
	}
	return "", fmt.Errorf("value %q not found in enum source %q", input, source)
}

func castParamValue(kind string, value any) (any, error) {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "", "string":
		return fmt.Sprintf("%v", value), nil
	case "int", "integer":
		return castToInt(value)
	case "number", "float":
		return castToFloat(value)
	case "bool", "boolean":
		return castToBool(value)
	case "array<int>", "[]int":
		return castToIntSlice(value)
	case "array<string>", "[]string":
		return castToStringSlice(value), nil
	default:
		return value, nil
	}
}

func castToInt(v any) (int, error) {
	s := strings.TrimSpace(fmt.Sprintf("%v", v))
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func castToFloat(v any) (float64, error) {
	s := strings.TrimSpace(fmt.Sprintf("%v", v))
	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func castToBool(v any) (bool, error) {
	s := strings.TrimSpace(strings.ToLower(fmt.Sprintf("%v", v)))
	switch s {
	case "1", "true", "yes", "y", "on":
		return true, nil
	case "0", "false", "no", "n", "off":
		return false, nil
	default:
		return false, fmt.Errorf("invalid bool value: %v", v)
	}
}

func castToIntSlice(v any) ([]int, error) {
	s := strings.TrimSpace(fmt.Sprintf("%v", v))
	if s == "" {
		return nil, nil
	}
	parts := strings.Split(s, ",")
	out := make([]int, 0, len(parts))
	for _, part := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, nil
}

func castToStringSlice(v any) []string {
	s := strings.TrimSpace(fmt.Sprintf("%v", v))
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item != "" {
			out = append(out, item)
		}
	}
	return out
}
