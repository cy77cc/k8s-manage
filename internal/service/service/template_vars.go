package service

import (
	"regexp"
	"sort"
	"strings"
)

var templateVarPattern = regexp.MustCompile(`\{\{\s*([a-zA-Z_][a-zA-Z0-9_\.\-]*)(?:\|default:([^}]+))?\s*\}\}`)

func detectTemplateVars(content string) []TemplateVar {
	matches := templateVarPattern.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return nil
	}
	uniq := make(map[string]TemplateVar)
	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		name := strings.TrimSpace(m[1])
		if name == "" {
			continue
		}
		def := ""
		if len(m) > 2 {
			def = strings.TrimSpace(m[2])
		}
		item := TemplateVar{
			Name:       name,
			Required:   def == "",
			Default:    def,
			SourcePath: "template",
		}
		uniq[name] = item
	}
	out := make([]TemplateVar, 0, len(uniq))
	for _, v := range uniq {
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func resolveTemplateVars(content string, reqValues map[string]string, envValues map[string]string) (string, []string) {
	out := templateVarPattern.ReplaceAllStringFunc(content, func(token string) string {
		m := templateVarPattern.FindStringSubmatch(token)
		if len(m) < 2 {
			return token
		}
		name := strings.TrimSpace(m[1])
		def := ""
		if len(m) > 2 {
			def = strings.TrimSpace(m[2])
		}
		if v, ok := reqValues[name]; ok && strings.TrimSpace(v) != "" {
			return v
		}
		if v, ok := envValues[name]; ok && strings.TrimSpace(v) != "" {
			return v
		}
		if def != "" {
			return def
		}
		return token
	})
	var unresolved []string
	for _, v := range detectTemplateVars(out) {
		if v.Required {
			unresolved = append(unresolved, v.Name)
		}
	}
	return out, unresolved
}
