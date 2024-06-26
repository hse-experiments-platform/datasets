// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0

package common

import (
	"context"
)

type Querier interface {
	CreateDataset(ctx context.Context, arg CreateDatasetParams) (int64, error)
	DropColumnsByType(ctx context.Context, arg DropColumnsByTypeParams) error
	DropDatasetSchema(ctx context.Context, datasetID int64) error
	GetDataset(ctx context.Context, id int64) (GetDatasetRow, error)
	GetDatasetCreator(ctx context.Context, id int64) (int64, error)
	GetDatasetSchema(ctx context.Context, datasetID int64) ([]GetDatasetSchemaRow, error)
	GetDatasetStatus(ctx context.Context, id int64) (DatasetStatus, error)
	GetUserDatasets(ctx context.Context, arg GetUserDatasetsParams) ([]GetUserDatasetsRow, error)
	SetDatasetSchema(ctx context.Context, arg SetDatasetSchemaParams) error
	SetErrorStatus(ctx context.Context, arg SetErrorStatusParams) error
	SetRowsCount(ctx context.Context, arg SetRowsCountParams) error
	SetStatus(ctx context.Context, arg SetStatusParams) error
	UpdateAfterUpload(ctx context.Context, arg UpdateAfterUploadParams) error
	UpdateData(ctx context.Context, arg UpdateDataParams) error
}

var _ Querier = (*Queries)(nil)
