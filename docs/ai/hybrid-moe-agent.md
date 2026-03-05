# Hybrid MOE Agent

## Overview

The AI agent routing and execution path has been refactored to a Hybrid MOE + graph design:

1. `ExpertRegistry` loads expert definitions from `configs/experts.yaml`
2. `HybridRouter` selects experts using scene -> keyword -> domain -> default fallback
3. `graph.Builder` compiles a declarative expert graph for `parallel`/`sequential`/`primary_led` strategies
4. `Orchestrator` remains as a compatibility fallback and is marked deprecated
5. `ResultAggregator` merges expert outputs into a final response
6. `callbacks` module provides unified event emission for `tool_call`/`tool_result`/`expert_progress`

## Configuration

- Experts: `configs/experts.yaml`
- Scene mappings: `configs/scene_mappings.yaml`

Both files are resolved from current working directory with upward path fallback, so tests and runtime can load them from different process roots.

## Runtime Integration

`PlatformAgent` now owns:

- `registry` (`experts.ExpertRegistry`)
- `router` (`*experts.HybridRouter`)
- `orchestrator` (`*experts.Orchestrator`)
- `graphRunner` (`compose.Runnable[*graph.GraphInput, *graph.GraphOutput]`)

`Stream()` and `Generate()` route with `HybridRouter`, execute with `graphRunner` first, and fallback to `Orchestrator`/`react.Agent` on failure.

## Expert Collaboration as Tools

`ExpertRegistry` now injects expert-as-tool adapters into each expert's toolset (excluding self), so a primary expert can call helper experts using standard tool-calling semantics without regex directives.

## Testing

- Unit tests: registry/router/orchestrator/aggregator behavior
- Unit tests: callbacks + graph + expert tool adapter behavior
- Integration tests: end-to-end pipeline with real YAML config
- Regression tests: key scene mappings compatibility checks
- Performance baseline: benchmark for routing path (`BenchmarkHybridRouterRoute`)
