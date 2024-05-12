-- name: GetDatasetChunks :many
select raw_data_chunk, min_row_number, max_row_number, prefix_len
from datasets_data dd
join datasets_data_chunks ddc on dd.chunk_number = ddc.chunk_number
where dd.dataset_id = $1
  and @l <= dd.chunk_number
  and dd.chunk_number < @r;

-- name: ClearChunks :exec
DELETE
from datasets_data_chunks
where dataset_id = $1;

-- name: ClearOldChunks :exec
DELETE
from datasets_data_chunks
where dataset_id = $1 and chunk_number >= $2;