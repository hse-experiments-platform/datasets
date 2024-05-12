package datasets

import (
	"github.com/hse-experiments-platform/datasets/internal/pkg/models"
	"github.com/hse-experiments-platform/datasets/internal/pkg/storage/common"
	pb "github.com/hse-experiments-platform/datasets/pkg/datasets"
	launcherpb "github.com/hse-experiments-platform/launcher/pkg/launcher"
)

func convertStatus(s common.DatasetStatus) pb.DatasetStatus {
	switch s {
	case common.DatasetStatusInitializing:
		return pb.DatasetStatus_Initializing
	case common.DatasetStatusLoading:
		return pb.DatasetStatus_Loading
	case common.DatasetStatusWaitsConvertation:
		return pb.DatasetStatus_WaitsConvertation
	case common.DatasetStatusLoadingError:
		return pb.DatasetStatus_LoadingError
	case common.DatasetStatusConvertationInProgress:
		return pb.DatasetStatus_ConvertationInProgress
	case common.DatasetStatusConvertationError:
		return pb.DatasetStatus_ConvertationError
	case common.DatasetStatusReady:
		return pb.DatasetStatus_Ready
	case common.DatasetStatusValue0:
		return pb.DatasetStatus_Unknown
	default:
		return pb.DatasetStatus_Unknown
	}
}

func convertColumnTypePB(t pb.ColumnType) models.ColumnType {
	switch t {
	case pb.ColumnType_ColumnTypeUndefined:
		return models.ColumnTypeUndefined
	case pb.ColumnType_ColumnTypeString:
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

func convertSetTypesSettings(m *pb.SetTypeSettings) *launcherpb.SetTypeSettings {
	return &launcherpb.SetTypeSettings{
		Type:              convertColumnType(m.Type),
		FillingTechnique:  convertFillingTechnique(m.FillingTechnique),
		AggregateFunction: convertAggregateFunction(m.AggregateFunction),
		FillingValue:      m.FillingValue,
	}
}

func convertFillingTechnique(t pb.FillingTechnique) launcherpb.FillingTechnique {
	switch t {
	case pb.FillingTechnique_FillingTechniqueConstant:
		return launcherpb.FillingTechnique_FillingTechniqueConstant
	case pb.FillingTechnique_FillingTechniqueTypeDefault:
		return launcherpb.FillingTechnique_FillingTechniqueTypeDefault
	case pb.FillingTechnique_FillingTechniqueAggregateFunction:
		return launcherpb.FillingTechnique_FillingTechniqueAggregateFunction
	case pb.FillingTechnique_FillingTechniqueDeleteRow:
		return launcherpb.FillingTechnique_FillingTechniqueDeleteRow
	default:
		return launcherpb.FillingTechnique_FillingTechniqueUndefined
	}
}

func convertColumnType(t pb.ColumnType) launcherpb.ColumnType {
	switch t {
	case pb.ColumnType_ColumnTypeInteger:
		return launcherpb.ColumnType_ColumnTypeInteger
	case pb.ColumnType_ColumnTypeFloat:
		return launcherpb.ColumnType_ColumnTypeFloat
	case pb.ColumnType_ColumnTypeCategorial:
		return launcherpb.ColumnType_ColumnTypeCategorial
	case pb.ColumnType_ColumnTypeString:
		return launcherpb.ColumnType_ColumnTypeString
	case pb.ColumnType_ColumnTypeDropped:
		return launcherpb.ColumnType_ColumnTypeDropped
	default:
		return launcherpb.ColumnType_ColumnTypeUndefined
	}
}

func convertAggregateFunction(t pb.AggregateFunction) launcherpb.AggregateFunction {
	switch t {
	case pb.AggregateFunction_AggregateFunctionMean:
		return launcherpb.AggregateFunction_AggregateFunctionMean
	case pb.AggregateFunction_AggregateFunctionMedian:
		return launcherpb.AggregateFunction_AggregateFunctionMedian
	case pb.AggregateFunction_AggregateFunctionMin:
		return launcherpb.AggregateFunction_AggregateFunctionMin
	case pb.AggregateFunction_AggregateFunctionMax:
		return launcherpb.AggregateFunction_AggregateFunctionMax
	case pb.AggregateFunction_AggregateFunctionMostFrequent:
		return launcherpb.AggregateFunction_AggregateFunctionMostFrequent
	default:
		return launcherpb.AggregateFunction_AggregateFunctionUndefined
	}
}
