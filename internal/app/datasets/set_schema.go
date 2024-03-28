package datasets

import (
	"context"
	"fmt"

	"github.com/hse-experiments-platform/datasets/internal/pkg/models"
	"github.com/hse-experiments-platform/datasets/internal/pkg/storage/db"
	pb "github.com/hse-experiments-platform/datasets/pkg/datasets"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (d *datasetsService) constructSchema(rows []db.GetDatasetSchemaRow) map[string]string {
	res := make(map[string]string, len(rows))

	for _, row := range rows {
		res[row.ColumnName] = row.ColumnType
	}

	return res
}

func (d *datasetsService) SetDatasetColumnTypes(ctx context.Context, req *pb.SetDatasetColumnTypesRequest) (*pb.SetDatasetColumnTypesResponse, error) {
	if req.GetDatasetID() == 0 {
		return nil, status.Error(codes.InvalidArgument, "datasetID must be not 0")
	}
	if err := d.checkDatasetAccess(ctx, req.GetDatasetID()); err != nil {
		return nil, err
	}

	dataset, err := d.commonDB.GetDataset(ctx, req.GetDatasetID())
	if err != nil {
		return nil, fmt.Errorf("d.commonDB.GetDataset: %w", err)
	} else if dataset.Status != db.DatasetStatusWaitsConvertation && dataset.Status != db.DatasetStatusConvertationError {
		return nil, status.Errorf(codes.InvalidArgument, "dataset must be in status %v, got %v", db.DatasetStatusWaitsConvertation, dataset.Status)
	}

	schemaRows, err := d.commonDB.GetDatasetSchema(ctx, req.GetDatasetID())
	if err != nil {
		return nil, fmt.Errorf("d.commonDB.GetDatasetSchema: %w", err)
	} else if len(schemaRows) == 0 {
		return nil, status.Error(codes.Internal, "dataset schema empty")
	}

	dbReq := db.SetDatasetSchemaParams{
		DatasetID:   req.GetDatasetID(),
		Indexes:     make([]int32, len(schemaRows)),
		ColumnNames: make([]string, len(schemaRows)),
		ColumnTypes: make([]string, len(schemaRows)),
	}

	columnsMap := req.GetColumnTypes()

	for i, row := range schemaRows {
		t, ok := columnsMap[row.ColumnName]
		if !ok {
			return nil, status.Errorf(codes.InvalidArgument, "column type for column %s not specified", row.ColumnName)
		}

		dbReq.Indexes[i] = int32(i)
		dbReq.ColumnNames[i] = row.ColumnName
		if t == pb.ColumnType_ColumnTypeUndefined {
			return nil, status.Errorf(codes.InvalidArgument, "column type for column %s not selected", row.ColumnName)
		}
		dbReq.ColumnTypes[i] = models.TypeToString[convertColumnTypePB(t)]
	}

	err = pgx.BeginTxFunc(ctx, d.commonDBConn, pgx.TxOptions{}, func(tx pgx.Tx) error {
		defer func() {
			if err != nil {
				if e := tx.Rollback(ctx); e != nil {
					log.Error().Err(e).Msg("rollback error")
				}
			}
		}()

		txdb := d.commonDB.WithTx(tx)

		log.Debug().Any("req", dbReq).Msg("debug")
		if err = txdb.SetDatasetSchema(ctx, dbReq); err != nil {
			return fmt.Errorf("txdb.SetDatasetSchema: %w", err)
		}

		if err = txdb.SetStatus(ctx, db.SetStatusParams{
			ID:     req.GetDatasetID(),
			Status: db.DatasetStatusConvertationInProgress,
		}); err != nil {
			return fmt.Errorf("txdb.SetStatus: %w", err)
		}

		if err = d.setColumnTypes(ctx, dbReq, dataset.RowsCount); err != nil {
			return fmt.Errorf("d.setColumnTypes: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("pgx.BeginTxFunc: %w", err)
	}

	return &pb.SetDatasetColumnTypesResponse{}, nil
}
