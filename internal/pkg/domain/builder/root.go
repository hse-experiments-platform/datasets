package builder

import (
	"context"

	"github.com/hse-experiments-platform/datasets/internal/pkg/models"
	"github.com/hse-experiments-platform/datasets/internal/pkg/storage/datasets"
)

type Builder struct {
	DatasetsDB datasets.Querier
	DatasetID  int64
	Schema     models.Schema
	Ctx        context.Context

	prevRowPrefix []byte
	curRow        int64
	chunkNumber   int64
}

func NewBuilder(ctx context.Context, db datasets.Querier, datasetID int64) *Builder {
	return &Builder{
		DatasetsDB: db,
		DatasetID:  datasetID,
		Ctx:        ctx,
	}
}

type DatasetChunk struct {
	DatasetID    int64
	ChunkNumber  int64
	RawDataChunk []byte
}
