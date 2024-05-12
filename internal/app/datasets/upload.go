package datasets

import (
	"context"
	xerrors "errors"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/hse-experiments-platform/datasets/internal/pkg/domain"
	"github.com/hse-experiments-platform/datasets/internal/pkg/domain/errors"
	"github.com/hse-experiments-platform/datasets/internal/pkg/models"
	"golang.org/x/sync/errgroup"

	"github.com/hse-experiments-platform/datasets/internal/pkg/domain/builder"
	"github.com/hse-experiments-platform/datasets/internal/pkg/storage/common"
	"github.com/hse-experiments-platform/datasets/internal/pkg/utils"
	"github.com/hse-experiments-platform/datasets/internal/pkg/utils/pipe"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opensaucerer/grab/v3"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const bufferSize = 10 * 1024 * 1024 // 10mb
//const bufferSize = 12 // 10mb

func printProgress(ctx context.Context, r *grab.Response) {
	t := time.NewTicker(time.Second * 5)

	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Debug().Msg("printProgress: context cancelled")
			return
		case <-t.C:
		}
		log.Debug().Msgf("transferred %v / %v bytes (%.2f%%)", r.BytesComplete(), r.Size(), 100*r.Progress())
		if r.IsComplete() {
			return
		}
	}
}

func getS3WriteFunc(s3Buf io.ReadWriter) func([]byte) error {
	return func(data []byte) error {
		n, err := s3Buf.Write(data)
		if err != nil {
			return fmt.Errorf("s3Buf.Write: %w", err)
		} else if n != len(data) {
			return fmt.Errorf("buffered writer didn't write full data")
		}

		return nil
	}
}

func (d *datasetsService) uploadFromURL(baseCtx context.Context, addr string, userID, datasetID int64) error {
	client := grab.NewClient()
	client.BufferSize = bufferSize

	if _, err := url.Parse(addr); err != nil {
		return status.Errorf(codes.InvalidArgument, "cannot parse url, value=%s, err=%s", addr, err.Error())
	}

	g, ctx := errgroup.WithContext(baseCtx)
	ctx, c := context.WithTimeout(ctx, time.Minute*10)

	builder := builder.NewBuilder(ctx, d.datasetsDB, datasetID)
	s3Buffer := pipe.NewPipeReadWriter()

	writer := utils.NewChunkedFunctionCaller([]func(data []byte) error{builder.ProcessChunk, getS3WriteFunc(s3Buffer)}, bufferSize)

	req, err := grab.NewRequestToWriter(writer, addr)
	req.WithContext(ctx)
	if err != nil {
		c()
		return fmt.Errorf("grab.NewRequestToWriter: %w", err)
	}

	resp := client.Do(req)

	// start s3 upload
	g.Go(func() error {
		if err := d.uploadToS3(ctx, userID, datasetID, s3Buffer); err != nil {
			return errors.NewRevertable(fmt.Errorf("s3 upload failed: %w", err), "s3 upload failed")
		}

		return nil
	})

	// start progress printer
	g.Go(func() error {
		printProgress(ctx, resp)
		return nil
	})

	// process results
	g.Go(func() error {
		return d.processUploadResult(ctx, datasetID, writer, builder, resp, s3Buffer)
	})

	go func() {
		defer c()
		defer s3Buffer.Close()

		d.completeUpload(baseCtx, resp, g, builder, datasetID)
	}()

	return nil
}

func (d *datasetsService) completeUpload(ctx context.Context, r *grab.Response, g *errgroup.Group, builder *builder.Builder, datasetID int64) {
	err := g.Wait()
	// если завершились без ошибок, завершаем загрузку и отлавливаем новые ошибки, которые могли возникнуть
	if err == nil {
		log.Debug().Str("filename", r.Filename).Msg("file downloaded successfully")
		params := common.SetDatasetSchemaParams{DatasetID: datasetID}
		for i, v := range builder.Schema {
			params.Indexes = append(params.Indexes, int32(i))
			params.ColumnNames = append(params.ColumnNames, v.Name)
			params.ColumnTypes = append(params.ColumnTypes, models.TypeToString[v.Type])
		}
		if err = d.commonDB.SetDatasetSchema(ctx, params); err != nil {
			log.Error().Err(err).Msg("d.commonDB.SetDatasetSchema")
			err = errors.NewRevertable(fmt.Errorf("d.commonDB.SetDatasetSchema: %w", err), "cannot set dataset schema")
		}

		if err = d.commonDB.UpdateAfterUpload(ctx, common.UpdateAfterUploadParams{
			ID:        datasetID,
			Status:    common.DatasetStatusWaitsConvertation,
			Version:   "1.1",
			RowsCount: builder.GetRowsCount(),
		}); err != nil {
			log.Error().Err(err).Msg("cannot update meta for uploaded dataset")
			err = errors.NewRevertable(fmt.Errorf("cannot update meta for uploaded dataset: %w", err), "cannot update meta for uploaded dataset")
		}

		log.Debug().Msg("upload completed successfully")
		return
	}
	// если дошли до сюда, значит у нас в какой то момент возникла ошибка, откатываем
	log.Error().Err(err).Msg("error in upload")
	var revErr *errors.RevertableError
	if xerrors.As(err, &revErr) {
		if err = d.setFailedUpload(ctx, datasetID, revErr.GetReason()); err != nil {
			log.Error().Err(err).Msg("cannot set failed upload status")
			return
		}
	}
}

func (d *datasetsService) uploadToS3(ctx context.Context, userID int64, datasetID int64, buffer *pipe.PipeReadWriter) error {
	if err := d.minio.UploadObjectMultipart(
		ctx, domain.GetBucketName(userID), domain.GetObjectName(datasetID), buffer, -1,
	); err != nil {
		return fmt.Errorf("d.minio.UploadObjectMultipart: %w", err)
	}

	return nil
}

func (d *datasetsService) processUploadResult(ctx context.Context, datasetID int64, writer utils.ChunkedFunctionCaller, builder *builder.Builder, r *grab.Response, buffer *pipe.PipeReadWriter) error {
	// check for errors
	if err := r.Err(); err != nil {
		log.Error().Err(err).Msg("file download failed")
		return errors.NewRevertable(fmt.Errorf("file download failed: %w", err), "file download failed")
	}

	if err := writer.Flush(); err != nil {
		log.Error().Err(err).Msg("file download failed: flush error")
		return errors.NewRevertable(fmt.Errorf("file download failed: flush error: %w", err), "file download failed: flush error")
	}

	buffer.Close()

	c := builder.GetRowsCount()
	if c == -1 {
		log.Error().Msg("dataset wasn't parsed properly")
		return errors.NewRevertable(fmt.Errorf("dataset wasn't parsed properly: invalid rows count"), "dataset wasn't parsed properly")
	}

	return nil
}

func (d *datasetsService) setFailedUpload(ctx context.Context, datasetID int64, reason string) error {
	log.Error().Int64("datasetID", datasetID).Str("reason", reason).Msg("dataset upload attempt failed")
	if err := d.commonDB.SetErrorStatus(ctx, common.SetErrorStatusParams{
		ID:          datasetID,
		Status:      common.DatasetStatusLoadingError,
		UploadError: pgtype.Text{String: reason, Valid: true},
	}); err != nil {
		return fmt.Errorf("d.commonDB.SetStatus: %w", err)
	}
	if err := d.datasetsDB.DeleteDatasetData(ctx, datasetID); err != nil {
		return fmt.Errorf("d.datasetsDB.DeleteDatasetData: %w", err)
	}

	return nil
}
