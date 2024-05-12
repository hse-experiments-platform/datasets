package datasets

import (
	"context"
	"fmt"

	"github.com/hse-experiments-platform/datasets/internal/pkg/models"
	"github.com/hse-experiments-platform/datasets/internal/pkg/storage/common"
	pb "github.com/hse-experiments-platform/datasets/pkg/datasets"
	launcherpb "github.com/hse-experiments-platform/launcher/pkg/launcher"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func (d *datasetsService) constructSchema(rows []common.GetDatasetSchemaRow) map[string]string {
	res := make(map[string]string, len(rows))

	for _, row := range rows {
		res[row.ColumnName] = row.ColumnType
	}

	return res
}

func (d *datasetsService) SetDatasetColumnTypes(ctx context.Context, req *pb.SetDatasetColumnTypesRequest) (*pb.SetDatasetColumnTypesResponse, error) {
	if req.GetDatasetID() == 0 {
		return nil, status.Error(codes.InvalidArgument, "datasetID must be not 0")
	}
	if err := d.checkDatasetAccess(ctx, req.GetDatasetID()); err != nil {
		return nil, err
	}

	log.Debug().Any("req", req).Msg("req")

	dataset, err := d.commonDB.GetDataset(ctx, req.GetDatasetID())
	if err != nil {
		return nil, fmt.Errorf("d.commonDB.GetDataset: %w", err)
	} else if dataset.Status != common.DatasetStatusWaitsConvertation && dataset.Status != common.DatasetStatusConvertationError {
		return nil, status.Errorf(codes.InvalidArgument, "dataset must be in status %v, got %v", common.DatasetStatusWaitsConvertation, dataset.Status)
	}

	schemaRows, err := d.commonDB.GetDatasetSchema(ctx, req.GetDatasetID())
	if err != nil {
		return nil, fmt.Errorf("d.commonDB.GetDatasetSchema: %w", err)
	} else if len(schemaRows) == 0 {
		return nil, status.Error(codes.Internal, "dataset schema empty")
	}

	dbReq := common.SetDatasetSchemaParams{
		DatasetID:   req.GetDatasetID(),
		Indexes:     make([]int32, len(schemaRows)),
		ColumnNames: make([]string, len(schemaRows)),
		ColumnTypes: make([]string, len(schemaRows)),
	}
	settings := make(map[string]*launcherpb.SetTypeSettings, len(schemaRows))
	columnsMap := req.GetColumnTypes()

	for i, row := range schemaRows {
		t, ok := columnsMap[row.ColumnName]
		if !ok {
			return nil, status.Errorf(codes.InvalidArgument, "column type for column %s not specified", row.ColumnName)
		} else if t.Type == pb.ColumnType_ColumnTypeUndefined {
			return nil, status.Errorf(codes.InvalidArgument, "column type for column %s not selected", row.ColumnName)
		}

		dbReq.Indexes[i] = int32(i)
		dbReq.ColumnNames[i] = row.ColumnName
		dbReq.ColumnTypes[i] = models.TypeToString[convertColumnTypePB(t.Type)]
		settings[row.ColumnName] = convertSetTypesSettings(t)
	}

	log.Debug().Any("settings", settings).Msg("set column types")

	md, _ := metadata.FromIncomingContext(ctx)
	resp, err := d.launcher.LaunchDatasetSetTypes(metadata.NewOutgoingContext(ctx, md), &launcherpb.LaunchDatasetSetTypesRequest{
		LaunchInfo: &launcherpb.CommonLaunchInfo{
			Name:        "Dataset set types and fill blanks",
			Description: fmt.Sprintf("Dataset set types and fill blanks for datasetID: %v", req.GetDatasetID()),
		},
		DatasetID:    req.GetDatasetID(),
		NewDatasetID: req.GetDatasetID(),
		ColumnTypes:  settings,
	})
	if err != nil {
		return nil, fmt.Errorf("d.launcher.LaunchDatasetSetTypes: %w", err)
	}

	log.Debug().Int64("dataset_id", req.GetDatasetID()).Int64("launch_id", resp.GetLaunchID()).Msg("dataset set types launch started")

	go func() {
		err := d.setColumnTypes(ctx, dataset.ID, resp.LaunchID, dbReq)
		if err != nil {
			log.Error().Err(err).Msg("failed set column types")
		}
	}()

	return &pb.SetDatasetColumnTypesResponse{}, nil
}
