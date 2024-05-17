package datasets

import (
	"context"

	xerrors "errors"
	"fmt"

	"github.com/hse-experiments-platform/datasets/internal/pkg/domain/errors"
	"github.com/hse-experiments-platform/datasets/internal/pkg/models"
	"github.com/hse-experiments-platform/datasets/internal/pkg/storage/common"
	launcherpb "github.com/hse-experiments-platform/launcher/pkg/launcher"
	"github.com/rs/zerolog/log"
)

func (d *datasetsService) uploadWithLauncher(baseCtx context.Context, addr string, userID, datasetID int64) error {
	ctx := context.Background()
	ctx, err := d.maker.TransferToAnotherContext(baseCtx, ctx)
	if err != nil {
		return fmt.Errorf("d.maker.TransferToAnotherContext: %w", err)
	}

	log.Debug().Int64("dataset_id", datasetID).Msg("staring dataset upload launch")
	resp, err := d.launcher.LaunchDatasetUpload(ctx, &launcherpb.LaunchDatasetUploadRequest{
		LaunchInfo: &launcherpb.CommonLaunchInfo{
			Name:        "Dataset upload",
			Description: fmt.Sprintf("Dataset upload for datasetID: %v", datasetID),
		},
		Url:       addr,
		DatasetID: datasetID,
	})
	if err != nil {
		return fmt.Errorf("d.launcher.LaunchDatasetUpload: %w", err)
	}
	log.Debug().Int64("dataset_id", datasetID).Int64("launch_id", resp.GetLaunchID()).Msg("dataset upload launch started")

	go func() {
		err := d.processLaunch(ctx, resp.GetLaunchID(), datasetID)
		if err != nil {
			log.Error().Err(err).Msg("processLaunch")
		}
	}()

	return nil
}

func (d *datasetsService) getRowsCount(ctx context.Context, datasetID int64) (int64, error) {
	var rowsCount int64
	err := d.datasetsDBConn.QueryRow(ctx, fmt.Sprintf("select count(1) from dataset_%v", datasetID)).Scan(&rowsCount)
	if err != nil {
		log.Error().Err(err).Msg("d.datasetsDBConn.QueryRow")
		return 0, errors.NewRevertable(fmt.Errorf("d.datasetsDBConn.QueryRow: %w", err), "cannot get rows count")
	}

	return rowsCount, nil
}

func (d *datasetsService) processLaunch(ctx context.Context, launchID, datasetID int64) (err error) {
	var cols []string
	defer func() {
		if err == nil && cols != nil {
			params := common.SetDatasetSchemaParams{DatasetID: datasetID}
			for i, v := range cols {
				params.Indexes = append(params.Indexes, int32(i))
				params.ColumnNames = append(params.ColumnNames, v)
				params.ColumnTypes = append(params.ColumnTypes, models.TypeToString[models.ColumnTypeUndefined])
			}
			if err = d.commonDB.SetDatasetSchema(ctx, params); err != nil {
				log.Error().Err(err).Msg("d.commonDB.SetDatasetSchema")
				err = errors.NewRevertable(fmt.Errorf("d.commonDB.SetDatasetSchema: %w", err), "cannot set dataset schema")
			}

			var rowsCount int64
			err = d.datasetsDBConn.QueryRow(ctx, fmt.Sprintf("select count(1) from dataset_%v", datasetID)).Scan(&rowsCount)
			if err != nil {
				log.Error().Err(err).Msg("d.datasetsDBConn.QueryRow")
				err = errors.NewRevertable(fmt.Errorf("d.datasetsDBConn.QueryRow: %w", err), "cannot get rows count")
			}

			if err = d.commonDB.UpdateAfterUpload(ctx, common.UpdateAfterUploadParams{
				ID:        datasetID,
				Status:    common.DatasetStatusWaitsConvertation,
				Version:   "1.1",
				RowsCount: rowsCount, // TODO: select rows from postgres table
			}); err != nil {
				log.Error().Err(err).Msg("cannot update meta for uploaded dataset")
				err = errors.NewRevertable(fmt.Errorf("cannot update meta for uploaded dataset: %w", err), "cannot update meta for uploaded dataset")
			}

			log.Debug().Msg("upload completed successfully")
			if err == nil {
				return
			}
		}
		// если дошли до сюда, значит у нас в какой то момент возникла ошибка, откатываем
		log.Error().Err(err).Msg("error in upload")
		var revErr *errors.RevertableError
		if !xerrors.As(err, &revErr) {
			xerrors.As(errors.NewRevertable(err, "Ошибка при загрузке датасета"), &revErr)
		}
		if err = d.setFailedUpload(ctx, datasetID, revErr.GetReason()); err != nil {
			log.Error().Err(err).Msg("cannot set failed upload status")
			return
		}
	}()

	_, err = d.launcher.WaitForLaunchFinish(ctx, &launcherpb.WaitForLaunchFinishRequest{
		LaunchID: launchID,
	})
	if err != nil {
		return errors.NewRevertable(fmt.Errorf("d.launcher.WaitForLaunchFinish: %w", err), "cannot wait for launch finish")
	} else {
		log.Debug().Int64("launch_id", launchID).Msg("launch finished")
	}

	launch, err := d.launcher.GetDatasetUploadLaunch(ctx, &launcherpb.GetDatasetUploadLaunchRequest{
		LaunchID: launchID,
	})
	if err != nil {
		return errors.NewRevertable(fmt.Errorf("d.launcher.GetDatasetUploadLaunch: %w", err), "cannot get launch info")
	}

	colsMap := launch.GetColumnTypes()
	for k := range colsMap {
		cols = append(cols, k)
	}

	return nil
}
