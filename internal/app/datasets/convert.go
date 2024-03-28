package datasets

import (
	"github.com/hse-experiments-platform/datasets/internal/pkg/models"
	"github.com/hse-experiments-platform/datasets/internal/pkg/storage/db"
	pb "github.com/hse-experiments-platform/datasets/pkg/datasets"
)

func convertStatus(s db.DatasetStatus) pb.DatasetStatus {
	switch s {
	case db.DatasetStatusInitializing:
		return pb.DatasetStatus_Initializing
	case db.DatasetStatusLoading:
		return pb.DatasetStatus_Loading
	case db.DatasetStatusWaitsConvertation:
		return pb.DatasetStatus_WaitsConvertation
	case db.DatasetStatusLoadingError:
		return pb.DatasetStatus_LoadingError
	case db.DatasetStatusConvertationInProgress:
		return pb.DatasetStatus_ConvertationInProgress
	case db.DatasetStatusConvertationError:
		return pb.DatasetStatus_ConvertationError
	case db.DatasetStatusReady:
		return pb.DatasetStatus_Ready
	case db.DatasetStatusValue0:
		return pb.DatasetStatus_Unknown
	default:
		return pb.DatasetStatus_Unknown
	}
}

func convertColumnTypePB(t pb.ColumnType) models.ColumnType {
	switch t {
	case pb.ColumnType_ColumnTypeNotSelected, pb.ColumnType_ColumnTypeString:
		return models.ColumnTypeString
	case pb.ColumnType_ColumnTypeInteger:
		return models.ColumnTypeInteger
	case pb.ColumnType_ColumnTypeFloat:
		return models.ColumnTypeFloat
	case pb.ColumnType_ColumnTypeCategorial:
		return models.ColumnTypeCategorial
	case pb.ColumnType_ColumnTypeDropped:
		return models.ColumnTypeDropped
	default:
		return models.ColumnTypeString
	}
}
