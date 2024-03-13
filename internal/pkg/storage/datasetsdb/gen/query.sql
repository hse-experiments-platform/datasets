-- name: GetDatasetChunkBorders :one
SELECT min_row_number, max_row_number
from datasets_data
WHERE dataset_id = $1;

-- name: GetDatasetChunks :many
select UNNEST(raw_data_chunks[indexes.l : indexes.r]) as chunks
from (select $2::int as l, $3::int as r) as indexes
         join datasets_data d on d.dataset_id = $1;

-- name: CreateEmptyDatasetData :exec
INSERT INTO datasets_data (dataset_id, raw_data_chunks, min_row_number, max_row_number)
VALUES ($1, array []::bytea[], array []::bigint[], array []::bigint[]);

-- name: UploadDatasetChunks :exec
UPDATE datasets_data
set raw_data_chunks = raw_data_chunks || $2::bytea[],
    min_row_number = min_row_number || $3::bigint[],
    max_row_number = max_row_number || $4::bigint[]
where dataset_id = $1;