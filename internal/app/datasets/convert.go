package datasets

import (
	"github.com/hse-experiments-platform/datasets/internal/pkg/storage/db"
	pb "github.com/hse-experiments-platform/datasets/pkg/datasets"
)

func convertStatus(s db.DatasetStatus) pb.DatasetStatus {
	switch s {
	case db.DatasetStatusInitializing:
		return pb.DatasetStatus_Initializing
	case db.DatasetStatusLoading:
		return pb.DatasetStatus_Loading
	case db.DatasetStatusReady:
		return pb.DatasetStatus_Ready
	case db.DatasetStatusError:
		return pb.DatasetStatus_Error
	default:
		return pb.DatasetStatus_Unknown
	}
}
