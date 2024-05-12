-- +goose Up
-- +goose StatementBegin
insert into datasets_data_chunks (dataset_id, chunk_number, raw_data_chunk)
select dataset_id, chunk_number, raw_data_chunk
from datasets_data;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
truncate datasets_data_chunks;
-- +goose StatementEnd
