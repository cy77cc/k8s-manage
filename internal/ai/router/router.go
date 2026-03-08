package router

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/compose"
	"github.com/cy77cc/k8s-manage/internal/ai/tools"
)

const (
	nodePreprocess = "preprocess"
	nodeClassify   = "classify"
	nodeGeneral    = "domain_general"
	nodeInfra      = "domain_infrastructure"
	nodeService    = "domain_service"
	nodeCICD       = "domain_cicd"
	nodeMonitor    = "domain_monitor"
	nodeConfig     = "domain_config"
)

// IntentRouter routes user input into a tool domain via compose.Graph.
type IntentRouter struct {
	classifier *IntentClassifier
	graph      *compose.Graph[string, tools.ToolDomain]
	runnable   compose.Runnable[string, tools.ToolDomain]
}

func NewIntentRouter(ctx context.Context, classifier *IntentClassifier) (*IntentRouter, error) {
	if classifier == nil {
		classifier = NewIntentClassifier(nil, nil)
	}

	graph := compose.NewGraph[string, tools.ToolDomain]()
	if err := graph.AddLambdaNode(nodePreprocess, compose.InvokableLambda(preprocessInput)); err != nil {
		return nil, fmt.Errorf("add preprocess node: %w", err)
	}
	if err := graph.AddLambdaNode(nodeClassify, compose.InvokableLambda(classifier.Classify)); err != nil {
		return nil, fmt.Errorf("add classify node: %w", err)
	}
	for key, domain := range map[string]tools.ToolDomain{
		nodeGeneral: tools.DomainGeneral,
		nodeInfra:   tools.DomainInfrastructure,
		nodeService: tools.DomainService,
		nodeCICD:    tools.DomainCICD,
		nodeMonitor: tools.DomainMonitor,
		nodeConfig:  tools.DomainConfig,
	} {
		d := domain
		if err := graph.AddLambdaNode(key, compose.InvokableLambda(func(ctx context.Context, input tools.ToolDomain) (tools.ToolDomain, error) {
			if input == "" {
				return d, nil
			}
			return input, nil
		})); err != nil {
			return nil, fmt.Errorf("add domain node %s: %w", key, err)
		}
	}

	if err := graph.AddEdge(compose.START, nodePreprocess); err != nil {
		return nil, fmt.Errorf("add start edge: %w", err)
	}
	if err := graph.AddEdge(nodePreprocess, nodeClassify); err != nil {
		return nil, fmt.Errorf("add classify edge: %w", err)
	}
	if err := graph.AddBranch(nodeClassify, compose.NewGraphBranch(selectDomainNode, map[string]bool{
		nodeGeneral: true,
		nodeInfra:   true,
		nodeService: true,
		nodeCICD:    true,
		nodeMonitor: true,
		nodeConfig:  true,
	})); err != nil {
		return nil, fmt.Errorf("add branch: %w", err)
	}
	for _, node := range []string{nodeGeneral, nodeInfra, nodeService, nodeCICD, nodeMonitor, nodeConfig} {
		if err := graph.AddEdge(node, compose.END); err != nil {
			return nil, fmt.Errorf("add end edge for %s: %w", node, err)
		}
	}

	runnable, err := graph.Compile(ctx)
	if err != nil {
		return nil, fmt.Errorf("compile router graph: %w", err)
	}

	return &IntentRouter{
		classifier: classifier,
		graph:      graph,
		runnable:   runnable,
	}, nil
}

func (r *IntentRouter) Route(ctx context.Context, input string) (tools.ToolDomain, error) {
	if r == nil || r.runnable == nil {
		return tools.DomainGeneral, fmt.Errorf("intent router is not initialized")
	}
	return r.runnable.Invoke(ctx, input)
}

func (r *IntentRouter) Graph() *compose.Graph[string, tools.ToolDomain] {
	if r == nil {
		return nil
	}
	return r.graph
}

func preprocessInput(_ context.Context, input string) (string, error) {
	return normalizeInput(input), nil
}

func selectDomainNode(_ context.Context, domain tools.ToolDomain) (string, error) {
	switch domain {
	case tools.DomainInfrastructure:
		return nodeInfra, nil
	case tools.DomainService:
		return nodeService, nil
	case tools.DomainCICD:
		return nodeCICD, nil
	case tools.DomainMonitor:
		return nodeMonitor, nil
	case tools.DomainConfig:
		return nodeConfig, nil
	default:
		return nodeGeneral, nil
	}
}
