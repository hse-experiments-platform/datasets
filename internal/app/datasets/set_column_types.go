package datasets

import (
	"context"
	"fmt"

	"github.com/hse-experiments-platform/datasets/internal/pkg/domain/errors"
	"github.com/hse-experiments-platform/datasets/internal/pkg/models"
	"github.com/hse-experiments-platform/datasets/internal/pkg/storage/common"
	"github.com/hse-experiments-platform/launcher/pkg/launcher"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/metadata"
)

type PreprocessRequest struct {
	UserID    string      `json:"user_id"`
	DatasetID string      `json:"dataset_id"`
	Schema    [][2]string `json:"schema"`
}

func (d *datasetsService) setColumnTypes(userCtx context.Context, datasetID, launchID int64, req common.SetDatasetSchemaParams) (err error) {
	ctx := context.Background()
	md, _ := metadata.FromIncomingContext(userCtx)
	ctx = metadata.NewOutgoingContext(ctx, md)

	defer func() {
		if err == nil {
			err = d.onConvertationSuccess(ctx, req)
		}
		if err != nil {
			d.onConvartationFail(ctx, req)
		}
	}()

	if err = d.commonDB.SetStatus(ctx, common.SetStatusParams{
		ID:     datasetID,
		Status: common.DatasetStatusConvertationInProgress,
	}); err != nil {
		return fmt.Errorf("d.commonDB.SetStatus: %w", err)
	}

	if _, err = d.launcher.WaitForLaunchFinish(ctx, &launcher.WaitForLaunchFinishRequest{
		LaunchID: launchID,
	}); err != nil {
		return fmt.Errorf("d.launcher.WaitForLaunchFinish: %w", err)
	}

	var resp *launcher.GetDatasetSetTypesLaunchResponse
	resp, err = d.launcher.GetDatasetSetTypesLaunch(ctx, &launcher.GetDatasetSetTypesLaunchRequest{
		LaunchID: launchID,
	})
	if err != nil {
		return fmt.Errorf("d.launcher.GetDatasetSetTypesLaunch: %w", err)
	} else if resp.LaunchInfo.Status != launcher.LaunchStatus_LaunchStatusSuccess || resp.Error != "" {
		err = fmt.Errorf("launch failed: %s", resp.Error)
		return errors.NewRevertable(err, resp.Error)
	}

	return nil
}

func (d *datasetsService) onConvertationSuccess(ctx context.Context, params common.SetDatasetSchemaParams) error {
	rowsCount, err := d.getRowsCount(ctx, params.DatasetID)
	if err != nil {
		return fmt.Errorf("d.getRowsCount: %w", err)
	}

	if err := pgx.BeginTxFunc(ctx, d.commonDBConn, pgx.TxOptions{
		AccessMode: pgx.ReadWrite,
	}, func(tx pgx.Tx) (err error) {
		defer func() {
			if err != nil {
				errTx := tx.Rollback(ctx)
				if errTx != nil {
					log.Error().Err(errTx).Err(err).Msg("tx.Rollback error")
				}
			}
		}()

		txdb := d.commonDB.WithTx(tx)

		// обновляем количество строк
		if err = txdb.SetRowsCount(ctx, common.SetRowsCountParams{
			ID:        params.DatasetID,
			RowsCount: rowsCount,
		}); err != nil {
			return fmt.Errorf("txdb.SetRowsCount: %w", err)
		}

		// обновляем на новую схему
		if err = txdb.SetDatasetSchema(ctx, params); err != nil {
			return fmt.Errorf("txdb.SetDatasetSchema: %w", err)
		}

		// удаляем все дропаемые колонки
		if err = txdb.DropColumnsByType(ctx, common.DropColumnsByTypeParams{
			DatasetID:  params.DatasetID,
			ColumnType: models.TypeToString[models.ColumnTypeDropped],
		}); err != nil {
			return fmt.Errorf("txdb.DropColumnsByType: %w", err)
		}

		if err = txdb.SetStatus(ctx, common.SetStatusParams{
			ID:     params.DatasetID,
			Status: common.DatasetStatusReady,
		}); err != nil {
			return fmt.Errorf("txdb.SetStatus: %w", err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("pgx.BeginTxFunc: %w", err)
	}

	return nil
}

func (d *datasetsService) onConvartationFail(ctx context.Context, params common.SetDatasetSchemaParams) {
	err := d.commonDB.SetStatus(ctx, common.SetStatusParams{
		ID:     params.DatasetID,
		Status: common.DatasetStatusConvertationError,
	})
	if err != nil {
		log.Error().Err(err).Msg("d.commonDB.SetStatus")
	}
}
