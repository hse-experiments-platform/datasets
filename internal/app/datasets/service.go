package datasets

import (
	"context"

	"github.com/hse-experiments-platform/datasets/internal/pkg/storage/datasetsdb"
	"github.com/hse-experiments-platform/datasets/internal/pkg/storage/db"
	pb "github.com/hse-experiments-platform/datasets/pkg/datasets"
	"github.com/hse-experiments-platform/library/pkg/utils/token"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ pb.DatasetsServiceServer = (*datasetsService)(nil)

type datasetsService struct {
	pb.UnimplementedDatasetsServiceServer
	datasetsDBConn *pgx.Conn
	commonDBConn   *pgx.Conn
	maker          token.Maker

	commonDB   *db.Queries
	datasetsDB *datasetsdb.Queries
}

func NewService(commonDBConn, datasetsDBConn *pgx.Conn, maker token.Maker) *datasetsService {
	return &datasetsService{
		commonDBConn:   commonDBConn,
		datasetsDBConn: datasetsDBConn,
		maker:          maker,

		commonDB:   db.New(commonDBConn),
		datasetsDB: datasetsdb.New(datasetsDBConn),
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
