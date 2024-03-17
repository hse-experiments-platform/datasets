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
set status    = $2,
    updated_at = now()
where id = $1;

-- name: GetUserDatasets :many
select id,
       name,
       version,
       status,
       count(1) over () as count
from datasets
where creator_id = $1 and name like $4
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
       rows_count
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

