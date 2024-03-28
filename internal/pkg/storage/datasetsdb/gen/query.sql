-- name: GetDatasetChunkBorders :one
SELECT array_agg(min_row_number) as min_row_number, array_agg(max_row_number) as max_row_number
from datasets_data
WHERE dataset_id = $1
GROUP BY dataset_id;

-- name: GetDatasetChunks :many
select raw_data_chunk, min_row_number, max_row_number, prefix_len
from datasets_data
where dataset_id = $1
  and @l <= chunk_number
  and chunk_number < @r;

-- name: UploadDatasetChunk :exec
INSERT into datasets_data (dataset_id, raw_data_chunk, min_row_number, max_row_number, chunk_number, prefix_len)
VALUES ($1, $2, $3, $4, $5, $6)
on conflict (dataset_id, chunk_number) do update
    set raw_data_chunk = excluded.raw_data_chunk,
        min_row_number = excluded.min_row_number,
        max_row_number = excluded.max_row_number,
        prefix_len     = excluded.prefix_len;

-- name: DeleteDatasetData :exec
DELETE
from datasets_data
where dataset_id = $1;

-- name: DeleteOldDatasetData :exec
DELETE
from datasets_data
where dataset_id = $1 and chunk_number >= $2;