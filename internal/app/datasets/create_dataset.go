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

func (d *datasetsService) CreateDataset(ctx context.Context, request *pb.CreateDatasetRequest) (*pb.CreateDatasetResponse, error) {
	if len(request.GetName()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty name")
	}

	userID, err := getUserID(ctx)
	if err != nil {
		return nil, err
	}

	id, err := d.commonDB.CreateDataset(ctx, db.CreateDatasetParams{
		Name:      request.GetName(),
		CreatorID: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("d.commonDB.CreateDataset: %w", err)
	}

	return &pb.CreateDatasetResponse{DatasetID: id}, nil
}

func (d *datasetsService) prepareForUploadFunc(ctx context.Context, datasetID int64) func(tx pgx.Tx) error {
	return func(tx pgx.Tx) error {
		txDB := d.commonDB.WithTx(tx)

		st, err := txDB.GetDatasetStatus(ctx, datasetID)
		if err != nil {
			return fmt.Errorf("txDB.GetDatasetStatus: %w", err)
		} else if st != db.DatasetStatusInitializing {
			return status.Errorf(codes.InvalidArgument, "invalid dataset status to upload: expected initializing, got: %v", st)
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
