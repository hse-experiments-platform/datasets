package datasets

import (
	"context"
	"fmt"

	"github.com/hse-experiments-platform/datasets/internal/pkg/storage/common"
	pb "github.com/hse-experiments-platform/datasets/pkg/datasets"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (d *datasetsService) GetDatasets(ctx context.Context, request *pb.GetDatasetsRequest) (*pb.GetDatasetsResponse, error) {
	if request.GetLimit() == 0 {
		return nil, status.Error(codes.InvalidArgument, "limit must be greater than 0")
	}

	userID, err := getUserID(ctx)
	if err != nil {
		return nil, err
	}

	var statuses []common.DatasetStatus
	if filter := request.GetStatuses(); filter != nil {
		if filter.IncludeInitializing {
			statuses = append(statuses, common.DatasetStatusInitializing)
		}
		if filter.IncludeLoading {
			statuses = append(statuses, common.DatasetStatusLoading)
		}
		if filter.IncludeLoadingError {
			statuses = append(statuses, common.DatasetStatusLoadingError)
		}
		if filter.IncludeWaitsConvertation {
			statuses = append(statuses, common.DatasetStatusWaitsConvertation)
		}
		if filter.IncludeConvertationInProgress {
			statuses = append(statuses, common.DatasetStatusConvertationInProgress)
		}
		if filter.IncludeConvertationError {
			statuses = append(statuses, common.DatasetStatusConvertationError)
		}
		if filter.IncludeReady {
			statuses = append(statuses, common.DatasetStatusReady)
		}
	} else {
		statuses = append(statuses, common.DatasetStatusValue0, common.DatasetStatusInitializing, common.DatasetStatusLoading,
			common.DatasetStatusWaitsConvertation, common.DatasetStatusLoadingError, common.DatasetStatusConvertationInProgress,
			common.DatasetStatusConvertationError, common.DatasetStatusReady,
		)
	}

	query := "%" + request.GetQuery() + "%"
	rows, err := d.commonDB.GetUserDatasets(ctx, common.GetUserDatasetsParams{
		CreatorID:       userID,
		Limit:           int64(request.GetLimit()),
		Offset:          int64(request.GetOffset()),
		Name:            query,
		AllowedStatuses: statuses,
	})
	if err != nil {
		return nil, fmt.Errorf("d.commonDB.GetUserDatasets: %w", err)
	}

	total := int64(0)
	if len(rows) != 0 {
		total = rows[0].Count
	}

	resp := &pb.GetDatasetsResponse{
		PageInfo: &pb.PageInfo{
			Limit:  request.GetLimit(),
			Offset: request.GetOffset(),
			Total:  uint64(total),
		},
	}

	for _, row := range rows {
		resp.Datasets = append(resp.Datasets, &pb.ShortDatasetInfo{
			Id:      row.ID,
			Name:    row.Name,
			Version: row.Version,
			Status:  convertStatus(row.Status),
		})
	}

	return resp, nil
}
