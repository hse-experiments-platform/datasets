package datasets

import (
	"context"
	"fmt"

	"github.com/hse-experiments-platform/datasets/internal/pkg/storage/db"
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

	statuses := []db.DatasetStatus{}
	if filter := request.GetStatuses(); filter != nil {
		if filter.IncludeInitializing {
			statuses = append(statuses, db.DatasetStatusInitializing)
		}
		if filter.IncludeLoading {
			statuses = append(statuses, db.DatasetStatusLoading)
		}
		if filter.IncludeError {
			statuses = append(statuses, db.DatasetStatusError)
		}
		if filter.IncludeReady {
			statuses = append(statuses, db.DatasetStatusReady)
		}
	} else {
		statuses = append(statuses, db.DatasetStatusInitializing, db.DatasetStatusLoading, db.DatasetStatusError, db.DatasetStatusReady)
	}

	query := "%" + request.GetQuery() + "%"
	rows, err := d.commonDB.GetUserDatasets(ctx, db.GetUserDatasetsParams{
		CreatorID: userID,
		Limit:     int64(request.GetLimit()),
		Offset:    int64(request.GetOffset()),
		Name:      query,
		Column5:   statuses,
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
