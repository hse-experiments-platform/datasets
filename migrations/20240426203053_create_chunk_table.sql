-- +goose Up
-- +goose StatementBegin
create table if not exists datasets_data_chunks
(
    dataset_id     bigint not null,
    chunk_number   bigint not null,
    raw_data_chunk bytea  not null,
    primary key (dataset_id, chunk_number)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists datasets_data_chunks;
-- +goose StatementEnd
