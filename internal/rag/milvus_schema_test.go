package rag

import (
	"context"
	"testing"

	milvusclient "github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

type fakeCollectionAdmin struct {
	hasMap  map[string]bool
	created []string
}

func (f *fakeCollectionAdmin) HasCollection(_ context.Context, collName string) (bool, error) {
	return f.hasMap[collName], nil
}

func (f *fakeCollectionAdmin) CreateCollection(_ context.Context, schema *entity.Schema, _ int32, _ ...milvusclient.CreateCollectionOption) error {
	f.created = append(f.created, schema.CollectionName)
	return nil
}

func TestCollectionSpecs(t *testing.T) {
	specs := DefaultCollectionSpecs()
	if len(specs) != 3 {
		t.Fatalf("expected 3 specs, got %d", len(specs))
	}
	expect := map[string]bool{
		CollectionToolExamples:    false,
		CollectionPlatformAssets:  false,
		CollectionTroubleshooting: false,
	}
	for _, spec := range specs {
		expect[spec.Name] = true
		if spec.Schema == nil {
			t.Fatalf("schema should not be nil for %s", spec.Name)
		}
		foundEmbedding := false
		for _, field := range spec.Schema.Fields {
			if field.Name == "embedding" {
				foundEmbedding = true
				if field.TypeParams[entity.TypeParamDim] != "1536" {
					t.Fatalf("embedding dim mismatch for %s: %v", spec.Name, field.TypeParams[entity.TypeParamDim])
				}
				break
			}
		}
		if !foundEmbedding {
			t.Fatalf("embedding field not found in %s", spec.Name)
		}
	}
	for name, ok := range expect {
		if !ok {
			t.Fatalf("missing collection spec for %s", name)
		}
	}
}

func TestEnsureCollections(t *testing.T) {
	admin := &fakeCollectionAdmin{
		hasMap: map[string]bool{
			CollectionToolExamples: true,
		},
	}
	if err := EnsureCollections(context.Background(), admin, DefaultCollectionSpecs()); err != nil {
		t.Fatalf("ensure collections: %v", err)
	}
	if len(admin.created) != 2 {
		t.Fatalf("expected 2 created collections, got %d (%v)", len(admin.created), admin.created)
	}
}

func TestDefaultCollectionSpecsWithDim(t *testing.T) {
	specs := DefaultCollectionSpecsWithDim(1024)
	if len(specs) != 3 {
		t.Fatalf("expected 3 specs, got %d", len(specs))
	}
	for _, spec := range specs {
		foundEmbedding := false
		for _, field := range spec.Schema.Fields {
			if field.Name == "embedding" {
				foundEmbedding = true
				if field.TypeParams[entity.TypeParamDim] != "1024" {
					t.Fatalf("embedding dim mismatch for %s: %v", spec.Name, field.TypeParams[entity.TypeParamDim])
				}
			}
		}
		if !foundEmbedding {
			t.Fatalf("embedding field not found in %s", spec.Name)
		}
	}
}
