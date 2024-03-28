-- name: CreateDataset :one
insert into datasets (name, creator_id)
values ($1, $2)
returning id;

-- name: UpdateAfterUpload :exec
update datasets
set status     = $2,
    version    = $3,
    rows_count = $4,
    updated_at = now()
where id = $1;

-- name: UpdateData :exec
update datasets
set version    = $2,
    rows_count = $3,
    updated_at = now()
where id = $1;

-- name: SetStatus :exec
update datasets
set status     = $2,
    updated_at = now()
where id = $1;

-- name: SetErrorStatus :exec
update datasets
set status       = $2,
    updated_at   = now(),
    upload_error = $3
where id = $1;

-- name: GetUserDatasets :many
select id,
       name,
       version,
       status,
       count(1) over () as count
from datasets
where creator_id = $1
  and name like $4
  and status = any (sqlc.arg(allowed_statuses)::dataset_status[])
order by created_at desc
limit $2 offset $3;

-- name: GetDataset :one
select id,
       name,
       version,
       status,
       creator_id,
       created_at,
       updated_at,
       rows_count,
       upload_error
from datasets
where id = $1;

-- name: GetDatasetCreator :one
select creator_id
from datasets
where id = $1;

-- name: GetDatasetStatus :one
select status
from datasets
where id = $1;

-- name: SetDatasetSchema :exec
insert into dataset_schemas (dataset_id, column_number, column_name, column_type)
select $1,
       unnest(sqlc.arg(indexes)::int[]),
       unnest(sqlc.arg(column_names)::text[]),
       unnest(sqlc.arg(column_types)::text[])
on conflict (dataset_id, column_number) do update
    set column_name = excluded.column_name,
        column_type = excluded.column_type
;

-- name: GetDatasetSchema :many
select column_name, column_type
from dataset_schemas
where dataset_id = $1
order by column_number;
