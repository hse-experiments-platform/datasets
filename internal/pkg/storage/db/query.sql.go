// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: query.sql

package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createDataset = `-- name: CreateDataset :one
insert into datasets (name, creator_id)
values ($1, $2)
returning id
`

type CreateDatasetParams struct {
	Name      string
	CreatorID int64
}

func (q *Queries) CreateDataset(ctx context.Context, arg CreateDatasetParams) (int64, error) {
	row := q.db.QueryRow(ctx, createDataset, arg.Name, arg.CreatorID)
	var id int64
	err := row.Scan(&id)
	return id, err
}

const getDataset = `-- name: GetDataset :one
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
where id = $1
`

type GetDatasetRow struct {
	ID          int64
	Name        string
	Version     string
	Status      DatasetStatus
	CreatorID   int64
	CreatedAt   pgtype.Timestamptz
	UpdatedAt   pgtype.Timestamptz
	RowsCount   int64
	UploadError pgtype.Text
}

func (q *Queries) GetDataset(ctx context.Context, id int64) (GetDatasetRow, error) {
	row := q.db.QueryRow(ctx, getDataset, id)
	var i GetDatasetRow
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Version,
		&i.Status,
		&i.CreatorID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.RowsCount,
		&i.UploadError,
	)
	return i, err
}

const getDatasetCreator = `-- name: GetDatasetCreator :one
select creator_id
from datasets
where id = $1
`

func (q *Queries) GetDatasetCreator(ctx context.Context, id int64) (int64, error) {
	row := q.db.QueryRow(ctx, getDatasetCreator, id)
	var creator_id int64
	err := row.Scan(&creator_id)
	return creator_id, err
}

const getDatasetSchema = `-- name: GetDatasetSchema :many
select column_name, column_type
from dataset_schemas
where dataset_id = $1
order by column_number
`

type GetDatasetSchemaRow struct {
	ColumnName string
	ColumnType string
}

func (q *Queries) GetDatasetSchema(ctx context.Context, datasetID int64) ([]GetDatasetSchemaRow, error) {
	rows, err := q.db.Query(ctx, getDatasetSchema, datasetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetDatasetSchemaRow
	for rows.Next() {
		var i GetDatasetSchemaRow
		if err := rows.Scan(&i.ColumnName, &i.ColumnType); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getDatasetStatus = `-- name: GetDatasetStatus :one
select status
from datasets
where id = $1
`

func (q *Queries) GetDatasetStatus(ctx context.Context, id int64) (DatasetStatus, error) {
	row := q.db.QueryRow(ctx, getDatasetStatus, id)
	var status DatasetStatus
	err := row.Scan(&status)
	return status, err
}

const getUserDatasets = `-- name: GetUserDatasets :many
select id,
       name,
       version,
       status,
       count(1) over () as count
from datasets
where creator_id = $1
  and name like $4
  and status = any ($5::dataset_status[])
order by created_at desc
limit $2 offset $3
`

type GetUserDatasetsParams struct {
	CreatorID       int64
	Limit           int64
	Offset          int64
	Name            string
	AllowedStatuses []DatasetStatus
}

type GetUserDatasetsRow struct {
	ID      int64
	Name    string
	Version string
	Status  DatasetStatus
	Count   int64
}

func (q *Queries) GetUserDatasets(ctx context.Context, arg GetUserDatasetsParams) ([]GetUserDatasetsRow, error) {
	rows, err := q.db.Query(ctx, getUserDatasets,
		arg.CreatorID,
		arg.Limit,
		arg.Offset,
		arg.Name,
		arg.AllowedStatuses,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetUserDatasetsRow
	for rows.Next() {
		var i GetUserDatasetsRow
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Version,
			&i.Status,
			&i.Count,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const setDatasetSchema = `-- name: SetDatasetSchema :exec
insert into dataset_schemas (dataset_id, column_number, column_name, column_type)
select $1,
       unnest($2::int[]),
       unnest($3::text[]),
       unnest($4::text[])
on conflict (dataset_id, column_number) do update
    set column_name = excluded.column_name,
        column_type = excluded.column_type
`

type SetDatasetSchemaParams struct {
	DatasetID   int64
	Indexes     []int32
	ColumnNames []string
	ColumnTypes []string
}

func (q *Queries) SetDatasetSchema(ctx context.Context, arg SetDatasetSchemaParams) error {
	_, err := q.db.Exec(ctx, setDatasetSchema,
		arg.DatasetID,
		arg.Indexes,
		arg.ColumnNames,
		arg.ColumnTypes,
	)
	return err
}

const setErrorStatus = `-- name: SetErrorStatus :exec
update datasets
set status       = $2,
    updated_at   = now(),
    upload_error = $3
where id = $1
`

type SetErrorStatusParams struct {
	ID          int64
	Status      DatasetStatus
	UploadError pgtype.Text
}

func (q *Queries) SetErrorStatus(ctx context.Context, arg SetErrorStatusParams) error {
	_, err := q.db.Exec(ctx, setErrorStatus, arg.ID, arg.Status, arg.UploadError)
	return err
}

const setStatus = `-- name: SetStatus :exec
update datasets
set status     = $2,
    updated_at = now()
where id = $1
`

type SetStatusParams struct {
	ID     int64
	Status DatasetStatus
}

func (q *Queries) SetStatus(ctx context.Context, arg SetStatusParams) error {
	_, err := q.db.Exec(ctx, setStatus, arg.ID, arg.Status)
	return err
}

const updateAfterUpload = `-- name: UpdateAfterUpload :exec
update datasets
set status     = $2,
    version    = $3,
    rows_count = $4,
    updated_at = now()
where id = $1
`

type UpdateAfterUploadParams struct {
	ID        int64
	Status    DatasetStatus
	Version   string
	RowsCount int64
}

func (q *Queries) UpdateAfterUpload(ctx context.Context, arg UpdateAfterUploadParams) error {
	_, err := q.db.Exec(ctx, updateAfterUpload,
		arg.ID,
		arg.Status,
		arg.Version,
		arg.RowsCount,
	)
	return err
}

const updateData = `-- name: UpdateData :exec
update datasets
set version    = $2,
    rows_count = $3,
    updated_at = now()
where id = $1
`

type UpdateDataParams struct {
	ID        int64
	Version   string
	RowsCount int64
}

func (q *Queries) UpdateData(ctx context.Context, arg UpdateDataParams) error {
	_, err := q.db.Exec(ctx, updateData, arg.ID, arg.Version, arg.RowsCount)
	return err
}
