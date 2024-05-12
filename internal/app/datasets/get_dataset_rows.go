package datasets

//
import (
	"context"
	"fmt"

	pb "github.com/hse-experiments-platform/datasets/pkg/datasets"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

func (d *datasetsService) GetDatasetRows(ctx context.Context, request *pb.GetDatasetRowsRequest) (*pb.GetDatasetRowsResponse, error) {
	if request.GetDatasetID() == 0 {
		return nil, status.Error(codes.InvalidArgument, "datasetID must be not 0")
	}
	if request.GetLimit() == 0 {
		return nil, status.Error(codes.InvalidArgument, "limit must be greater than 0")
	}

	if err := d.checkDatasetAccess(ctx, request.GetDatasetID()); err != nil {
		return nil, err
	}

	total := 0
	err := d.datasetsDBConn.QueryRow(ctx, fmt.Sprintf("SELECT count(1) FROM dataset_%v", request.DatasetID)).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("d.datasetsDBConn.QueryRow: %w", err)
	}

	rows, err := d.datasetsDBConn.Query(ctx, fmt.Sprintf("SELECT * FROM dataset_%v LIMIT $1 OFFSET $2", request.DatasetID), request.GetLimit(), request.GetOffset())
	if err != nil {
		return nil, fmt.Errorf("d.datasetsDBConn.Query: %w", err)
	}
	defer rows.Close()

	resp := &pb.GetDatasetRowsResponse{
		PageInfo: &pb.PageInfo{
			Offset: request.GetOffset(),
			Limit:  request.GetLimit(),
		},
	}
	i := 0
	for rows.Next() {
		vals, err := rows.Values()
		if err != nil {
			return nil, fmt.Errorf("rows.Values: %w", err)
		}

		cols := make([]*structpb.Value, len(vals))
		for i, v := range vals {
			cols[i], err = structpb.NewValue(v)
			if err != nil {
				return nil, fmt.Errorf("structpb.NewValue: %w", err)
			}
		}

		resp.Rows = append(resp.Rows, &pb.DatasetRow{
			RowNumber: request.Offset + uint64(i),
			Columns:   cols,
		})
		i++

	}

	resp.PageInfo.Total = uint64(total)

	return resp, nil
}
