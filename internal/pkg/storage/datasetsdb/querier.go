// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0

package datasetsdb

import (
	"context"
)

type Querier interface {
	CreateEmptyDatasetData(ctx context.Context, datasetID int64) error
	GetDatasetChunkBorders(ctx context.Context, datasetID int64) (GetDatasetChunkBordersRow, error)
	GetDatasetChunks(ctx context.Context, arg GetDatasetChunksParams) ([][]byte, error)
	UploadDatasetChunks(ctx context.Context, arg UploadDatasetChunksParams) error
}

var _ Querier = (*Queries)(nil)
