-- +goose Up
-- +goose StatementBegin
alter table datasets_data
    drop column if exists raw_data_chunk,
    add column if not exists chunk_len int not null default 0;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table datasets_data
    add column if not exists raw_data_chunk bytea not null default '',
    drop column if exists chunk_len;
-- +goose StatementEnd
