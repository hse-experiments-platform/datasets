-- name: GetDatasetChunkBorders :many
SELECT min_row_number                                                                  as min_row_number,
       max_row_number                                                                  as max_row_number,
       chunk_len                                                                       as chunk_len,
       prefix_len                                                                      as prefix_len,
       coalesce(sum(chunk_len) over (rows unbounded preceding EXCLUDE CURRENT ROW), 0::bigint) as "offset"
from datasets_data
WHERE dataset_id = $1;

-- name: UploadDatasetChunk :exec
INSERT into datasets_data (dataset_id, min_row_number, max_row_number, chunk_number, prefix_len, chunk_len)
VALUES ($1, $2, $3, $4, $5, $6)
on conflict (dataset_id, chunk_number) do update
    set min_row_number = excluded.min_row_number,
        max_row_number = excluded.max_row_number,
        prefix_len     = excluded.prefix_len,
        chunk_len      =excluded.chunk_len;

-- name: DeleteDatasetData :exec
DELETE
from datasets_data
where dataset_id = $1;

-- name: DeleteOldDatasetData :exec
DELETE
from datasets_data
where dataset_id = $1
  and chunk_number >= $2;
