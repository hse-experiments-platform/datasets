// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: query.sql

package datasetsdb

import (
	"context"
)

const clearChunks = `-- name: ClearChunks :exec
DELETE
from datasets_data_chunks
where dataset_id = $1
`

func (q *Queries) ClearChunks(ctx context.Context, datasetID int64) error {
	_, err := q.db.Exec(ctx, clearChunks, datasetID)
	return err
}

const clearOldChunks = `-- name: ClearOldChunks :exec
DELETE
from datasets_data_chunks
where dataset_id = $1 and chunk_number >= $2
`

type ClearOldChunksParams struct {
	DatasetID   int64
	ChunkNumber int64
}

func (q *Queries) ClearOldChunks(ctx context.Context, arg ClearOldChunksParams) error {
	_, err := q.db.Exec(ctx, clearOldChunks, arg.DatasetID, arg.ChunkNumber)
	return err
}

const getDatasetChunks = `-- name: GetDatasetChunks :many
select raw_data_chunk, min_row_number, max_row_number, prefix_len
from datasets_data dd
join datasets_data_chunks ddc on dd.chunk_number = ddc.chunk_number
where dd.dataset_id = $1
  and $2 <= dd.chunk_number
  and dd.chunk_number < $3
`

type GetDatasetChunksParams struct {
	DatasetID int64
	L         int64
	R         int64
}

type GetDatasetChunksRow struct {
	RawDataChunk []byte
	MinRowNumber int64
	MaxRowNumber int64
	PrefixLen    int32
}

func (q *Queries) GetDatasetChunks(ctx context.Context, arg GetDatasetChunksParams) ([]GetDatasetChunksRow, error) {
	rows, err := q.db.Query(ctx, getDatasetChunks, arg.DatasetID, arg.L, arg.R)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetDatasetChunksRow
	for rows.Next() {
		var i GetDatasetChunksRow
		if err := rows.Scan(
			&i.RawDataChunk,
			&i.MinRowNumber,
			&i.MaxRowNumber,
			&i.PrefixLen,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
