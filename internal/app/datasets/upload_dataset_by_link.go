package datasets

import (
	"context"
	"fmt"

	pb "github.com/hse-experiments-platform/datasets/pkg/datasets"
	"github.com/jackc/pgx/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (d *datasetsService) UploadDatasetByLink(ctx context.Context, request *pb.UploadDatasetByLinkRequest) (*pb.UploadDatasetByLinkResponse, error) {
	if request.GetDatasetID() == 0 {
		return nil, status.Error(codes.InvalidArgument, "datasetID must be not 0")
	}
	if request.GetUrl() == "" {
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
