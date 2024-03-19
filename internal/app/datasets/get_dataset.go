package datasets

import (
	"context"
	"errors"
	"fmt"

	pb "github.com/hse-experiments-platform/datasets/pkg/datasets"
	"github.com/jackc/pgx/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (d *datasetsService) GetDataset(ctx context.Context, request *pb.GetDatasetRequest) (*pb.GetDatasetResponse, error) {
	if request.GetDatasetID() == 0 {
		return nil, status.Error(codes.InvalidArgument, "datasetID must be not 0")
	}

	userID, err := getUserID(ctx)
	if err != nil {
		return nil, err
	}

	dataset, err := d.commonDB.GetDataset(ctx, request.GetDatasetID())
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, status.Error(codes.NotFound, "dataset not found")
	} else if err != nil {
		return nil, fmt.Errorf("d.commonDB.GetUserDatasets: %w", err)
	}

	if dataset.CreatorID != userID {
		return nil, status.Error(codes.PermissionDenied, "cannot get other user's dataset")
	}

	return &pb.GetDatasetResponse{
		Dataset: &pb.Dataset{
			Id:          dataset.ID,
			Name:        dataset.Name,
			Version:     dataset.Version,
			Status:      convertStatus(dataset.Status),
			RowsCount:   uint64(dataset.RowsCount),
			CreatedAt:   timestamppb.New(dataset.CreatedAt.Time),
			UpdatedAt:   timestamppb.New(dataset.UpdatedAt.Time),
			UploadError: dataset.UploadError.String,
		},
	}, nil
}
