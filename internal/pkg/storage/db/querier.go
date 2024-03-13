// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0

package db

import (
	"context"
)

type Querier interface {
	CreateDataset(ctx context.Context, arg CreateDatasetParams) (int64, error)
	GetDataset(ctx context.Context, id int64) (GetDatasetRow, error)
	GetDatasetCreator(ctx context.Context, id int64) (int64, error)
	GetUserDatasets(ctx context.Context, arg GetUserDatasetsParams) ([]GetUserDatasetsRow, error)
	UpdateAfterUpload(ctx context.Context, arg UpdateAfterUploadParams) error
	UpdateData(ctx context.Context, arg UpdateDataParams) error
}

var _ Querier = (*Queries)(nil)
