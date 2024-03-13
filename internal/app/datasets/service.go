package datasets

import (
	"context"

	"github.com/hse-experiments-platform/datasets/internal/pkg/storage/datasetsdb"
	"github.com/hse-experiments-platform/datasets/internal/pkg/storage/db"
	pb "github.com/hse-experiments-platform/datasets/pkg/datasets"
	"github.com/hse-experiments-platform/library/pkg/utils/token"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ pb.DatasetsServiceServer = (*datasetsService)(nil)

type datasetsService struct {
	pb.UnimplementedDatasetsServiceServer
	commonDB   db.Querier
	datasetsDB datasetsdb.Querier
	maker      token.Maker
}

func NewService(commonDB db.Querier, datasetsDB datasetsdb.Querier, maker token.Maker) *datasetsService {
	return &datasetsService{
		commonDB:   commonDB,
		datasetsDB: datasetsDB,
		maker:      maker,
	}
}

func getUserID(ctx context.Context) (int64, error) {
	var userID int64
	userID, ok := ctx.Value(token.UserIDContextKey).(int64)
	if !ok {
		log.Error().Msgf("invalid userID context key type: %T", ctx.Value(token.UserIDContextKey))
		return 0, status.New(codes.Internal, "internal error").Err()
	}

	return userID, nil
}
