package graph

import (
	"context"

	"github.com/cloudwego/eino/compose"
)

const (
	nodeRoute           = "route"
	nodePrimary         = "primary"
	nodeHelpersParallel = "helpers_parallel"
	nodeHelpersSeq      = "helpers_sequential"
	nodeAggregate       = "aggregate"
)

// Builder creates the expert orchestration graph.
type Builder struct {
	primaryRunner PrimaryRunner
	helperRunner  HelperRunner
}

func NewBuilder() *Builder {
	return &Builder{}
}

func NewBuilderWithRunners(primary PrimaryRunner, helper HelperRunner) *Builder {
	return &Builder{primaryRunner: primary, helperRunner: helper}
}

func (b *Builder) Build(ctx context.Context) (*compose.Graph[*GraphInput, *GraphOutput], error) {
	_ = ctx
	g := compose.NewGraph[*GraphInput, *GraphOutput]()

	route := compose.InvokableLambda(func(_ context.Context, in *GraphInput) (*GraphInput, error) { return in, nil })
	primary := compose.InvokableLambda(func(ctx context.Context, in *GraphInput) (*GraphInput, error) {
		return runPrimary(ctx, b.primaryRunner, in)
	})
	parallelHelpers := compose.InvokableLambda(func(ctx context.Context, in *GraphInput) (*GraphInput, error) {
		return runHelpersParallel(ctx, b.helperRunner, in)
	})
	sequentialHelpers := compose.InvokableLambda(func(ctx context.Context, in *GraphInput) (*GraphInput, error) {
		return runHelpersSequential(ctx, b.helperRunner, in)
	})
	aggregate := compose.InvokableLambda(func(_ context.Context, in *GraphInput) (*GraphOutput, error) {
		return aggregateResults(in)
	})

	if err := g.AddLambdaNode(nodeRoute, route); err != nil {
		return nil, err
	}
	if err := g.AddLambdaNode(nodePrimary, primary); err != nil {
		return nil, err
	}
	if err := g.AddLambdaNode(nodeHelpersParallel, parallelHelpers); err != nil {
		return nil, err
	}
	if err := g.AddLambdaNode(nodeHelpersSeq, sequentialHelpers); err != nil {
		return nil, err
	}
	if err := g.AddLambdaNode(nodeAggregate, aggregate); err != nil {
		return nil, err
	}

	if err := g.AddEdge(compose.START, nodeRoute); err != nil {
		return nil, err
	}
	if err := g.AddEdge(nodeRoute, nodePrimary); err != nil {
		return nil, err
	}
	if err := g.AddBranch(nodePrimary, helperStrategyBranch()); err != nil {
		return nil, err
	}
	if err := g.AddEdge(nodeHelpersParallel, nodeAggregate); err != nil {
		return nil, err
	}
	if err := g.AddEdge(nodeHelpersSeq, nodeAggregate); err != nil {
		return nil, err
	}
	if err := g.AddEdge(nodeAggregate, compose.END); err != nil {
		return nil, err
	}
	return g, nil
}
