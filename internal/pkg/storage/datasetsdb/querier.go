// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0

package datasetsdb

import (
	"context"
)

type Querier interface {
	DeleteDatasetData(ctx context.Context, datasetID int64) error
	DeleteOldDatasetData(ctx context.Context, arg DeleteOldDatasetDataParams) error
	GetDatasetChunkBorders(ctx context.Context, datasetID int64) (GetDatasetChunkBordersRow, error)
	GetDatasetChunks(ctx context.Context, arg GetDatasetChunksParams) ([]GetDatasetChunksRow, error)
	UploadDatasetChunk(ctx context.Context, arg UploadDatasetChunkParams) error
}

var _ Querier = (*Queries)(nil)
