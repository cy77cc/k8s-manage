package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

const DefaultSkillsConfigPath = "configs/skills.yaml"

type SkillConfig struct {
	Version string  `yaml:"version"`
	Skills  []Skill `yaml:"skills"`
}

type Skill struct {
	Name                string           `yaml:"name"`
	DisplayName         string           `yaml:"display_name"`
	Description         string           `yaml:"description"`
	TriggerPatterns     []string         `yaml:"trigger_patterns"`
	RiskLevel           string           `yaml:"risk_level"`
	RequiredPermissions []string         `yaml:"required_permissions"`
	TimeoutMinutes      int              `yaml:"timeout_minutes"`
	Parameters          []SkillParameter `yaml:"parameters"`
	Steps               []SkillStep      `yaml:"steps"`
}

type SkillParameter struct {
	Name        string `yaml:"name"`
	Type        string `yaml:"type"`
	Required    bool   `yaml:"required"`
	Description string `yaml:"description"`
	Default     any    `yaml:"default"`
	EnumSource  string `yaml:"enum_source"`
}

type SkillStep struct {
	Name           string         `yaml:"name"`
	Type           string         `yaml:"type"`
	Tool           string         `yaml:"tool"`
	ParamsTemplate map[string]any `yaml:"params_template"`
	TimeoutSeconds int            `yaml:"timeout_seconds"`
}

type Registry struct {
	configPath string

	mu     sync.RWMutex
	config SkillConfig
	byName map[string]Skill
}

func NewRegistry(configPath string) (*Registry, error) {
	if strings.TrimSpace(configPath) == "" {
		configPath = DefaultSkillsConfigPath
	}
	r := &Registry{
		configPath: configPath,
		byName:     make(map[string]Skill),
	}
	if err := r.Reload(); err != nil {
		return nil, err
	}
	return r, nil
}

func LoadSkills(configPath string) (*SkillConfig, error) {
	if strings.TrimSpace(configPath) == "" {
		configPath = DefaultSkillsConfigPath
	}
	raw, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	var cfg SkillConfig
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return nil, err
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (r *Registry) Reload() error {
	cfg, err := LoadSkills(r.configPath)
	if err != nil {
		return err
	}
	index := make(map[string]Skill, len(cfg.Skills))
	for _, skill := range cfg.Skills {
		index[skill.Name] = skill
	}
	r.mu.Lock()
	r.config = *cfg
	r.byName = index
	r.mu.Unlock()
	return nil
}

func (r *Registry) List() []Skill {
	if r == nil {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Skill, 0, len(r.byName))
	for _, skill := range r.byName {
		out = append(out, skill)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (r *Registry) Get(name string) (Skill, bool) {
	if r == nil {
		return Skill{}, false
	}
	key := strings.TrimSpace(name)
	r.mu.RLock()
	defer r.mu.RUnlock()
	skill, ok := r.byName[key]
	return skill, ok
}

func (r *Registry) MatchSkill(message string) (*Skill, float64) {
	if r == nil {
		return nil, 0
	}
	text := strings.ToLower(strings.TrimSpace(message))
	if text == "" {
		return nil, 0
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	var (
		best      *Skill
		bestScore float64
	)
	for _, skill := range r.byName {
		score := matchScore(text, skill.TriggerPatterns)
		if score <= 0 {
			continue
		}
		if best == nil || score > bestScore {
			copy := skill
			best = &copy
			bestScore = score
		}
	}
	return best, bestScore
}

func (c *SkillConfig) Validate() error {
	if c == nil {
		return fmt.Errorf("skills config is nil")
	}
	if strings.TrimSpace(c.Version) == "" {
		return fmt.Errorf("version is required")
	}
	if len(c.Skills) == 0 {
		return fmt.Errorf("at least one skill is required")
	}
	seenSkill := make(map[string]struct{}, len(c.Skills))
	for i, skill := range c.Skills {
		if err := validateSkill(skill); err != nil {
			return fmt.Errorf("skills[%d]: %w", i, err)
		}
		if _, ok := seenSkill[skill.Name]; ok {
			return fmt.Errorf("duplicate skill name: %s", skill.Name)
		}
		seenSkill[skill.Name] = struct{}{}
	}
	return nil
}

func validateSkill(skill Skill) error {
	if strings.TrimSpace(skill.Name) == "" {
		return fmt.Errorf("name is required")
	}
	if len(skill.TriggerPatterns) == 0 {
		return fmt.Errorf("trigger_patterns is required")
	}
	if len(skill.Steps) == 0 {
		return fmt.Errorf("steps is required")
	}
	seenParams := make(map[string]struct{}, len(skill.Parameters))
	for _, param := range skill.Parameters {
		if strings.TrimSpace(param.Name) == "" {
			return fmt.Errorf("parameter name is required")
		}
		if strings.TrimSpace(param.Type) == "" {
			return fmt.Errorf("parameter %s type is required", param.Name)
		}
		if _, ok := seenParams[param.Name]; ok {
			return fmt.Errorf("duplicate parameter name: %s", param.Name)
		}
		seenParams[param.Name] = struct{}{}
	}
	for _, step := range skill.Steps {
		if strings.TrimSpace(step.Name) == "" {
			return fmt.Errorf("step name is required")
		}
		switch strings.ToLower(strings.TrimSpace(step.Type)) {
		case "tool":
			if strings.TrimSpace(step.Tool) == "" {
				return fmt.Errorf("step %s tool is required for tool step", step.Name)
			}
		case "approval", "resolver":
		default:
			return fmt.Errorf("step %s has unsupported type: %s", step.Name, step.Type)
		}
		if step.TimeoutSeconds < 0 {
			return fmt.Errorf("step %s timeout_seconds must be >= 0", step.Name)
		}
	}
	return nil
}

func matchScore(message string, patterns []string) float64 {
	if len(patterns) == 0 {
		return 0
	}
	best := 0.0
	for _, raw := range patterns {
		pattern := strings.ToLower(strings.TrimSpace(raw))
		if pattern == "" {
			continue
		}
		if strings.Contains(message, pattern) {
			score := float64(len(pattern)) / float64(len(message)+1)
			if score > best {
				best = score
			}
			continue
		}
		if ok, _ := filepath.Match(pattern, message); ok {
			if best < 0.2 {
				best = 0.2
			}
		}
	}
	return best
}
