package datasets

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/hse-experiments-platform/datasets/internal/pkg/storage/datasetsdb"
	pb "github.com/hse-experiments-platform/datasets/pkg/datasets"
	"github.com/jackc/pgx/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	chunksBatchSize = 2
)

func (d *datasetsService) GetDatasetRows(ctx context.Context, request *pb.GetDatasetRowsRequest) (*pb.GetDatasetRowsResponse, error) {
	if request.GetDatasetID() == 0 {
		return nil, status.Error(codes.InvalidArgument, "datasetID must be not 0")
	}
	if request.GetLimit() == 0 {
		return nil, status.Error(codes.InvalidArgument, "limit must be greater than 0")
	}

	userID, err := getUserID(ctx)
	if err != nil {
		return nil, err
	}

	creatorID, err := d.commonDB.GetDatasetCreator(ctx, request.GetDatasetID())
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, status.Error(codes.NotFound, "dataset not found")
	} else if err != nil {
		return nil, fmt.Errorf("d.commonDB.GetDatasetCreator: %w", err)
	}

	if creatorID != userID {
		return nil, status.Error(codes.PermissionDenied, "cannot get other user's dataset rows")
	}

	borders, err := d.datasetsDB.GetDatasetChunkBorders(ctx, request.GetDatasetID())
	if err != nil {
		return nil, fmt.Errorf("d.datasetsDB.GetDatasetChunkBorders: %w", err)
	} else if len(borders.MaxRowNumber) == 0 || len(borders.MinRowNumber) == 0 {
		return &pb.GetDatasetRowsResponse{
			PageInfo: &pb.PageInfo{
				Offset: request.GetOffset(),
				Limit:  request.GetLimit(),
				Total:  0,
			},
		}, nil
	}
	resp := &pb.GetDatasetRowsResponse{
		PageInfo: &pb.PageInfo{
			Offset: request.GetOffset(),
			Limit:  request.GetLimit(),
			Total:  uint64(borders.MaxRowNumber[len(borders.MaxRowNumber)-1]),
		},
	}

	if borders.MaxRowNumber[len(borders.MaxRowNumber)-1] < int64(request.GetOffset()) {
		return resp, nil
	}

	l, _ := slices.BinarySearch(borders.MaxRowNumber, int64(request.GetOffset()))
	r, _ := slices.BinarySearch(borders.MaxRowNumber, int64(request.GetOffset()+request.GetLimit()))

	// TODO: do in tx
	for i := l; i < r; i++ {
		chunks, err := d.datasetsDB.GetDatasetChunks(ctx, datasetsdb.GetDatasetChunksParams{
			DatasetID: request.GetDatasetID(),
			Column2:   int32(l),
			Column3:   int32(l + chunksBatchSize - 1),
		})
		if err != nil {
			return nil, fmt.Errorf("d.datasetsDB.GetDatasetChunks: %w", err)
		}

		_ = chunks
		//for _, chunk := range chunks {
		//    // TODO: split chunks
		//    //resp.Rows = append(resp.Rows, pb.DatasetRow{
		//    //    RowNumber: i,
		//    //    Columns:   nil,
		//    //})
		//}
	}

	return resp, nil
}
