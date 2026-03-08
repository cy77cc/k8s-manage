package router

import (
	"context"
	"fmt"
	"math"
	"strings"
	"unicode"
	"unicode/utf8"

	einomodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/cy77cc/k8s-manage/internal/ai/tools"
)

const classifierSystemPrompt = `You classify platform operation requests into one domain.
Respond with exactly one token from:
infrastructure
service
cicd
monitor
config
general`

// IntentClassifier maps user input to a tool domain.
type IntentClassifier struct {
	model  einomodel.ToolCallingChatModel
	config []DomainRouteConfig
}

func NewIntentClassifier(model einomodel.ToolCallingChatModel, config []DomainRouteConfig) *IntentClassifier {
	if len(config) == 0 {
		config = DefaultDomainRoutingConfig()
	}
	return &IntentClassifier{
		model:  model,
		config: append([]DomainRouteConfig(nil), config...),
	}
}

func (c *IntentClassifier) Classify(ctx context.Context, input string) (tools.ToolDomain, error) {
	classification, err := c.ClassifyWithDetail(ctx, input)
	if err != nil {
		return tools.DomainGeneral, err
	}
	return classification.Domain, nil
}

func (c *IntentClassifier) ClassifyWithDetail(ctx context.Context, input string) (Classification, error) {
	normalized := normalizeInput(input)
	if normalized == "" {
		return Classification{
			Domain:     tools.DomainGeneral,
			Normalized: normalized,
			Confidence: 0,
		}, nil
	}

	if c.model != nil {
		if domain, ok, err := c.classifyWithModel(ctx, normalized); err == nil && ok {
			return Classification{
				Domain:      domain,
				Normalized:  normalized,
				Confidence:  0.9,
				MatchedRule: "model",
			}, nil
		}
	}

	bestDomain := tools.DomainGeneral
	bestScore := 0
	bestRule := ""
	for _, rule := range c.config {
		score, matched := scoreRule(normalized, rule)
		if score > bestScore || (score == bestScore && bestDomain == tools.DomainGeneral && rule.Domain != tools.DomainGeneral) {
			bestDomain = rule.Domain
			bestScore = score
			bestRule = matched
		}
	}

	confidence := 0.15
	if bestScore > 0 {
		confidence = math.Min(0.95, 0.3+float64(bestScore)*0.2)
	}

	return Classification{
		Domain:      bestDomain,
		Normalized:  normalized,
		Confidence:  confidence,
		MatchedRule: bestRule,
	}, nil
}

func (c *IntentClassifier) classifyWithModel(ctx context.Context, normalized string) (tools.ToolDomain, bool, error) {
	msg, err := c.model.Generate(ctx, []*schema.Message{
		schema.SystemMessage(classifierSystemPrompt),
		schema.UserMessage(normalized),
	})
	if err != nil {
		return tools.DomainGeneral, false, err
	}

	domain, ok := parseModelDomain(msg.Content)
	return domain, ok, nil
}

func parseModelDomain(content string) (tools.ToolDomain, bool) {
	switch strings.TrimSpace(strings.ToLower(content)) {
	case string(tools.DomainInfrastructure):
		return tools.DomainInfrastructure, true
	case string(tools.DomainService):
		return tools.DomainService, true
	case string(tools.DomainCICD):
		return tools.DomainCICD, true
	case string(tools.DomainMonitor):
		return tools.DomainMonitor, true
	case string(tools.DomainConfig):
		return tools.DomainConfig, true
	case string(tools.DomainGeneral):
		return tools.DomainGeneral, true
	default:
		return tools.DomainGeneral, false
	}
}

func scoreRule(normalized string, rule DomainRouteConfig) (int, string) {
	totalScore := 0
	bestMatch := ""
	for _, keyword := range rule.Keywords {
		term := normalizeInput(keyword)
		if term == "" {
			continue
		}
		score := 0
		switch {
		case hasToken(normalized, term):
			score = 3
		case len(strings.Fields(term)) > 1 && strings.Contains(normalized, term):
			score = 2
		case utf8.RuneCountInString(term) >= 4 && strings.Contains(normalized, term):
			score = 1
		}
		totalScore += score
		if score > 0 && bestMatch == "" {
			bestMatch = keyword
		}
	}
	return totalScore, bestMatch
}

func hasToken(normalized, term string) bool {
	for _, token := range strings.Fields(normalized) {
		if token == term {
			return true
		}
	}
	return false
}

func normalizeInput(input string) string {
	input = strings.TrimSpace(strings.ToLower(input))
	if input == "" {
		return ""
	}

	var b strings.Builder
	b.Grow(len(input))
	lastSpace := false
	for _, r := range input {
		switch {
		case unicode.IsLetter(r), unicode.IsNumber(r):
			b.WriteRune(r)
			lastSpace = false
		case unicode.IsSpace(r):
			if !lastSpace {
				b.WriteByte(' ')
				lastSpace = true
			}
		default:
			if !lastSpace {
				b.WriteByte(' ')
				lastSpace = true
			}
		}
	}

	return strings.TrimSpace(b.String())
}

func (c *IntentClassifier) String() string {
	return fmt.Sprintf("IntentClassifier{rules:%d}", len(c.config))
}
