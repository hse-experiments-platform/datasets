package datasets

import (
	"context"
	"fmt"
	"os"

	"github.com/hse-experiments-platform/datasets/internal/pkg/domain"
	pb "github.com/hse-experiments-platform/datasets/pkg/datasets"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (d *datasetsService) GetDatasetDownloadLink(ctx context.Context, request *pb.GetDatasetDownloadLinkRequest) (*pb.GetDatasetDownloadLinkResponse, error) {
	if request.GetDatasetID() == 0 {
		return nil, status.Error(codes.InvalidArgument, "datasetID must be not 0")
	}

	if err := d.checkDatasetAccess(ctx, request.GetDatasetID()); err != nil {
		return nil, err
	}

	userID, _ := getUserID(ctx)

	url, err := d.minio.GetObjectDownloadLink(ctx, domain.GetBucketName(userID), domain.GetObjectName(request.GetDatasetID()))
	if err != nil {
		return nil, fmt.Errorf("d.minio.GetObjectDownloadLink: %w", err)
	}

	url.Host = os.Getenv("HOST")

	return &pb.GetDatasetDownloadLinkResponse{
		Url: url.String(),
	}, nil
}
