package datasets

import (
	"context"
	"errors"
	"fmt"

	"github.com/hse-experiments-platform/datasets/internal/pkg/storage/common"
	pb "github.com/hse-experiments-platform/datasets/pkg/datasets"
	"github.com/jackc/pgx/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var statusesWithoutSchema = map[common.DatasetStatus]bool{
	common.DatasetStatusValue0:       true,
	common.DatasetStatusInitializing: true,
	common.DatasetStatusLoading:      true,
	common.DatasetStatusLoadingError: true,
}

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

	columns, err := d.commonDB.GetDatasetSchema(ctx, request.GetDatasetID())
	if err != nil {
		return nil, fmt.Errorf("d.commonDB.GetDatasetSchema: %w", err)
	}
	if !statusesWithoutSchema[dataset.Status] && len(columns) == 0 {
		return nil, status.Errorf(codes.Internal, "empty schema for dataset %v", dataset.ID)
	}

	resp := &pb.GetDatasetResponse{
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
		Schema: &pb.DatasetSchema{
			Columns: make([]*pb.DatasetSchema_SchemaColumn, 0, len(columns)),
		},
	}

	for _, c := range columns {
		resp.Schema.Columns = append(resp.Schema.Columns, &pb.DatasetSchema_SchemaColumn{
			Name: c.ColumnName,
			Type: c.ColumnType,
		})
	}

	return resp, nil
}
