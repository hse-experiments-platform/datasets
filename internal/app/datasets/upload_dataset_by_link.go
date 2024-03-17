package datasets

import (
	"context"
	"fmt"

	"github.com/hse-experiments-platform/datasets/internal/pkg/storage/db"
	pb "github.com/hse-experiments-platform/datasets/pkg/datasets"
	"github.com/jackc/pgx/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var availableStatuses = map[db.DatasetStatus]bool{
	db.DatasetStatusInitializing: true,
	db.DatasetStatusError:        true,
}

func (d *datasetsService) prepareForUploadFunc(ctx context.Context, datasetID int64) func(tx pgx.Tx) error {
	return func(tx pgx.Tx) error {

		txDB := d.commonDB.WithTx(tx)

		st, err := txDB.GetDatasetStatus(ctx, datasetID)
		if err != nil {
			return fmt.Errorf("txDB.GetDatasetStatus: %w", err)
		} else if !availableStatuses[st] {
			return status.Errorf(codes.InvalidArgument, "invalid dataset status to upload: expected initializing/error, got: %v", st)
		}

		if err := txDB.SetStatus(ctx, db.SetStatusParams{
			ID:     datasetID,
			Status: db.DatasetStatusLoading,
		}); err != nil {
			return fmt.Errorf("txDB.SetStatus: %w", err)
		}

		return nil
	}
}

func (d *datasetsService) UploadDatasetByLink(ctx context.Context, request *pb.UploadDatasetByLinkRequest) (*pb.UploadDatasetByLinkResponse, error) {
	if request.GetDatasetID() == 0 {
		return nil, status.Error(codes.InvalidArgument, "datasetID must be not 0")
	} else if request.GetUrl() == "" {
		return nil, status.Error(codes.InvalidArgument, "url must be not empty")
	}

	if err := d.checkDatasetAccess(ctx, request.GetDatasetID()); err != nil {
		return nil, err
	}

	if err := pgx.BeginTxFunc(ctx, d.commonDBConn, pgx.TxOptions{
		IsoLevel:   pgx.Serializable,
		AccessMode: pgx.ReadWrite,
	}, d.prepareForUploadFunc(ctx, request.GetDatasetID())); err != nil {
		return nil, fmt.Errorf("pgx.BeginTxFunc: %w", err)
	}

	if err := d.uploadFromURL(context.Background(), request.GetUrl(), request.GetDatasetID()); err != nil {
		return nil, fmt.Errorf("d.uploadFromURL: %w", err)
	}

	return &pb.UploadDatasetByLinkResponse{}, nil
}
