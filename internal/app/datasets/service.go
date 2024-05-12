package datasets

import (
	"github.com/hse-experiments-platform/datasets/internal/pkg/storage/chunks/impl/s3"
	"github.com/hse-experiments-platform/datasets/internal/pkg/storage/common"
	db2 "github.com/hse-experiments-platform/datasets/internal/pkg/storage/datasets"
	pb "github.com/hse-experiments-platform/datasets/pkg/datasets"
	launcherpb "github.com/hse-experiments-platform/launcher/pkg/launcher"
	osinit "github.com/hse-experiments-platform/library/pkg/utils/init"
	"github.com/hse-experiments-platform/library/pkg/utils/token"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ pb.DatasetsServiceServer = (*datasetsService)(nil)

type datasetsService struct {
	pb.UnimplementedDatasetsServiceServer
	datasetsDBConn *pgxpool.Pool
	commonDBConn   *pgxpool.Pool
	maker          token.Maker
	minio          *s3.MinioStorage

	commonDB    *common.Queries
	datasetsDB  *db2.Queries
	storageType StorageType
	launcher    launcherpb.LauncherServiceClient
}

type StorageType string

const (
	StorageTypePG    = "postgres"
	StorageTypeMinio = "minio"
)

func NewService(commonDBConn, datasetsDBConn *pgxpool.Pool, minio *s3.MinioStorage, launcher launcherpb.LauncherServiceClient, maker token.Maker) *datasetsService {
	var storageType = StorageType(osinit.MustLoadEnv("DATASETS_STORAGE"))

	return &datasetsService{
		commonDBConn:   commonDBConn,
		datasetsDBConn: datasetsDBConn,
		maker:          maker,
		minio:          minio,
		storageType:    storageType,
		launcher:       launcher,

		commonDB:   common.New(commonDBConn),
		datasetsDB: db2.New(datasetsDBConn),
	}
}
