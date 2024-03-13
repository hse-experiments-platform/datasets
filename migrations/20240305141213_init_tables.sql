-- +goose Up
-- +goose StatementBegin

create table datasets_data
(
    dataset_id      bigint primary key,
    raw_data_chunks bytea[] not null,
    min_row_number  bigint[]   not null,
    max_row_number  bigint[]   not null
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table datasets_data;
-- +goose StatementEnd
