package catalog

import catalogv1 "github.com/cy77cc/k8s-manage/api/catalog/v1"

type CategoryCreateRequest = catalogv1.CategoryCreateRequest
type CategoryUpdateRequest = catalogv1.CategoryUpdateRequest
type CategoryResponse = catalogv1.CategoryResponse

type TemplateCreateRequest = catalogv1.TemplateCreateRequest
type TemplateUpdateRequest = catalogv1.TemplateUpdateRequest
type TemplateResponse = catalogv1.TemplateResponse
type TemplateListResponse = catalogv1.TemplateListResponse

type ReviewActionRequest = catalogv1.ReviewActionRequest

type PreviewRequest = catalogv1.PreviewRequest
type PreviewResponse = catalogv1.PreviewResponse

type DeployRequest = catalogv1.DeployRequest
type DeployResponse = catalogv1.DeployResponse

type CatalogVariableSchema = catalogv1.CatalogVariableSchema
