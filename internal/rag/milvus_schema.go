package rag

import (
	"context"
	"fmt"

	milvusclient "github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

const (
	DefaultVectorDim = int64(1536)

	CollectionToolExamples       = "tool_examples"
	CollectionPlatformAssets     = "platform_assets"
	CollectionTroubleshooting    = "troubleshooting_cases"
	defaultVarcharMaxLength int64 = 4096
)

type CollectionSpec struct {
	Name   string
	Schema *entity.Schema
	Shards int32
}

func ToolExamplesCollectionSpec() CollectionSpec {
	return ToolExamplesCollectionSpecWithDim(DefaultVectorDim)
}

func ToolExamplesCollectionSpecWithDim(dim int64) CollectionSpec {
	schema := entity.NewSchema().WithName(CollectionToolExamples).WithDescription("Historical successful tool usage examples.").
		WithField(entity.NewField().WithName("id").WithDataType(entity.FieldTypeInt64).WithIsPrimaryKey(true).WithIsAutoID(true)).
		WithField(entity.NewField().WithName("tool_name").WithDataType(entity.FieldTypeVarChar).WithMaxLength(256)).
		WithField(entity.NewField().WithName("intent").WithDataType(entity.FieldTypeVarChar).WithMaxLength(defaultVarcharMaxLength)).
		WithField(entity.NewField().WithName("params_json").WithDataType(entity.FieldTypeVarChar).WithMaxLength(defaultVarcharMaxLength)).
		WithField(entity.NewField().WithName("result_summary").WithDataType(entity.FieldTypeVarChar).WithMaxLength(defaultVarcharMaxLength)).
		WithField(entity.NewField().WithName("embedding").WithDataType(entity.FieldTypeFloatVector).WithDim(normalizeVectorDim(dim)))
	return CollectionSpec{Name: CollectionToolExamples, Schema: schema, Shards: entity.DefaultShardNumber}
}

func PlatformAssetsCollectionSpec() CollectionSpec {
	return PlatformAssetsCollectionSpecWithDim(DefaultVectorDim)
}

func PlatformAssetsCollectionSpecWithDim(dim int64) CollectionSpec {
	schema := entity.NewSchema().WithName(CollectionPlatformAssets).WithDescription("Platform assets searchable by semantic name and tags.").
		WithField(entity.NewField().WithName("id").WithDataType(entity.FieldTypeInt64).WithIsPrimaryKey(true).WithIsAutoID(true)).
		WithField(entity.NewField().WithName("asset_type").WithDataType(entity.FieldTypeVarChar).WithMaxLength(64)).
		WithField(entity.NewField().WithName("asset_id").WithDataType(entity.FieldTypeVarChar).WithMaxLength(128)).
		WithField(entity.NewField().WithName("name").WithDataType(entity.FieldTypeVarChar).WithMaxLength(512)).
		WithField(entity.NewField().WithName("status").WithDataType(entity.FieldTypeVarChar).WithMaxLength(64)).
		WithField(entity.NewField().WithName("metadata_json").WithDataType(entity.FieldTypeVarChar).WithMaxLength(defaultVarcharMaxLength)).
		WithField(entity.NewField().WithName("embedding").WithDataType(entity.FieldTypeFloatVector).WithDim(normalizeVectorDim(dim)))
	return CollectionSpec{Name: CollectionPlatformAssets, Schema: schema, Shards: entity.DefaultShardNumber}
}

func TroubleshootingCasesCollectionSpec() CollectionSpec {
	return TroubleshootingCasesCollectionSpecWithDim(DefaultVectorDim)
}

func TroubleshootingCasesCollectionSpecWithDim(dim int64) CollectionSpec {
	schema := entity.NewSchema().WithName(CollectionTroubleshooting).WithDescription("Troubleshooting knowledge cases for diagnosis.").
		WithField(entity.NewField().WithName("id").WithDataType(entity.FieldTypeInt64).WithIsPrimaryKey(true).WithIsAutoID(true)).
		WithField(entity.NewField().WithName("title").WithDataType(entity.FieldTypeVarChar).WithMaxLength(512)).
		WithField(entity.NewField().WithName("symptom").WithDataType(entity.FieldTypeVarChar).WithMaxLength(defaultVarcharMaxLength)).
		WithField(entity.NewField().WithName("diagnosis").WithDataType(entity.FieldTypeVarChar).WithMaxLength(defaultVarcharMaxLength)).
		WithField(entity.NewField().WithName("solution").WithDataType(entity.FieldTypeVarChar).WithMaxLength(defaultVarcharMaxLength)).
		WithField(entity.NewField().WithName("embedding").WithDataType(entity.FieldTypeFloatVector).WithDim(normalizeVectorDim(dim)))
	return CollectionSpec{Name: CollectionTroubleshooting, Schema: schema, Shards: entity.DefaultShardNumber}
}

func DefaultCollectionSpecs() []CollectionSpec {
	return DefaultCollectionSpecsWithDim(DefaultVectorDim)
}

func DefaultCollectionSpecsWithDim(dim int64) []CollectionSpec {
	normDim := normalizeVectorDim(dim)
	return []CollectionSpec{
		ToolExamplesCollectionSpecWithDim(normDim),
		PlatformAssetsCollectionSpecWithDim(normDim),
		TroubleshootingCasesCollectionSpecWithDim(normDim),
	}
}

type collectionAdmin interface {
	HasCollection(ctx context.Context, collName string) (bool, error)
	CreateCollection(ctx context.Context, schema *entity.Schema, shardsNum int32, opts ...milvusclient.CreateCollectionOption) error
}

func EnsureCollections(ctx context.Context, client collectionAdmin, specs []CollectionSpec) error {
	if client == nil {
		return fmt.Errorf("milvus collection admin is nil")
	}
	for _, spec := range specs {
		if spec.Schema == nil {
			return fmt.Errorf("nil schema for collection %s", spec.Name)
		}
		has, err := client.HasCollection(ctx, spec.Name)
		if err != nil {
			return fmt.Errorf("check collection %s: %w", spec.Name, err)
		}
		if has {
			continue
		}
		if err := client.CreateCollection(ctx, spec.Schema, spec.Shards); err != nil {
			return fmt.Errorf("create collection %s: %w", spec.Name, err)
		}
	}
	return nil
}

func (m *MilvusClient) EnsureCollections(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	ensureCtx, cancel := context.WithTimeout(ctx, m.operationTimeout()*3)
	defer cancel()
	if err := m.Connect(ensureCtx); err != nil {
		return err
	}
	m.mu.RLock()
	client := m.client
	m.mu.RUnlock()
	return EnsureCollections(ensureCtx, client, DefaultCollectionSpecsWithDim(int64(m.cfg.Dimension)))
}

func normalizeVectorDim(dim int64) int64 {
	if dim <= 0 {
		return DefaultVectorDim
	}
	return dim
}
