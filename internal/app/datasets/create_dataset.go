package datasets

import (
	"context"
	"fmt"

	"github.com/hse-experiments-platform/datasets/internal/pkg/storage/common"
	pb "github.com/hse-experiments-platform/datasets/pkg/datasets"
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

	id, err := d.commonDB.CreateDataset(ctx, common.CreateDatasetParams{
		Name:      request.GetName(),
		CreatorID: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("d.commonDB.CreateDataset: %w", err)
	}

	return &pb.CreateDatasetResponse{DatasetID: id}, nil
}
