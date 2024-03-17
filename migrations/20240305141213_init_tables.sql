-- +goose Up
-- +goose StatementBegin

create table datasets_data
(
    chunk_number   bigint not null,
    dataset_id     bigint not null,
    raw_data_chunk bytea  not null,
    min_row_number bigint not null,
    max_row_number bigint not null,
    prefix_len     int    not null,
    primary key (dataset_id, chunk_number)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table datasets_data;
-- +goose StatementEnd
