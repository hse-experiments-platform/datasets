package datasets

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/hse-experiments-platform/datasets/internal/pkg/domain/dataset"
	"github.com/hse-experiments-platform/datasets/internal/pkg/storage/db"
	"github.com/hse-experiments-platform/datasets/internal/pkg/utils"
	"github.com/opensaucerer/grab/v3"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const bufferSize = 10 * 1024 * 1024 // 10mb
//const bufferSize = 12 // 10mb

func (d *datasetsService) uploadFromURL(ctx context.Context, addr string, datasetID int64) error {
	client := grab.NewClient()
	client.BufferSize = bufferSize

	if _, err := url.Parse(addr); err != nil {
		return status.Errorf(codes.InvalidArgument, "cannot parse url, value=%s, err=%s", addr, err.Error())
	}

	builder := dataset.NewBuilder(ctx, d.datasetsDB, datasetID)

	writer := utils.NewChunkedFunctionCaller(builder.ProcessChunk, bufferSize)

	req, err := grab.NewRequestToWriter(writer, addr)
	if err != nil {
		return fmt.Errorf("grab.NewRequestToWriter: %w", err)
	}

	resp := client.Do(req)

	// start progress ticker
	t := time.NewTicker(time.Second * 5)

	// print progress
	go func(r *grab.Response) {
		defer t.Stop()
		for {
			<-t.C
			log.Debug().Msgf("transferred %v / %v bytes (%.2f%%)", r.BytesComplete(), r.Size(), 100*r.Progress())
			if resp.IsComplete() {
				return
			}
		}
	}(resp)

	// process results
	go func(r *grab.Response) {
		// check for errors
		if err := r.Err(); err != nil {
			log.Error().Err(err).Msg("file download failed")
			if err := d.setFailedUpload(ctx, datasetID); err != nil {
				log.Error().Err(err).Msg("cannot set failed upload status")
			}
			return
		}

		if err := writer.Flush(); err != nil {
			log.Error().Err(err).Msg("file download failed: flush error")
			if err := d.setFailedUpload(ctx, datasetID); err != nil {
				log.Error().Err(err).Msg("cannot set failed upload status")
			}
			return
		}

		c := builder.GetRowsCount()
		if c == -1 {
			log.Error().Err(err).Msg("dataset wasn't parsed properly")
			if err := d.setFailedUpload(ctx, datasetID); err != nil {
				log.Error().Err(err).Msg("cannot set failed upload status")
			}
			return
		}

		log.Debug().Str("filename", r.Filename).Msg("file downloaded successfully")

		if err := d.commonDB.UpdateAfterUpload(ctx, db.UpdateAfterUploadParams{
			ID:        datasetID,
			Status:    db.DatasetStatusReady,
			Version:   "1.1",
			RowsCount: c,
		}); err != nil {
			log.Error().Err(err).Msg("cannot update meta for uploaded dataset")
			if err := d.setFailedUpload(ctx, datasetID); err != nil {
				log.Error().Err(err).Msg("cannot set failed upload status")
			}
		}
	}(resp)

	// cancel on timeout
	go func(r *grab.Response) {
		select {
		case <-time.NewTimer(time.Second * 60 * 5).C:
			if err := r.Cancel(); err != nil {
				log.Error().Err(err).Msg("error in response.Cancel")
			}
		case <-r.Done:
			return
		}
	}(resp)

	return nil
}

func (d *datasetsService) setFailedUpload(ctx context.Context, datasetID int64) error {
	log.Error().Int64("datasetID", datasetID).Msg("dataset upload attempt failed")
	if err := d.commonDB.SetStatus(ctx, db.SetStatusParams{
		ID:     datasetID,
		Status: db.DatasetStatusError,
	}); err != nil {
		return fmt.Errorf("d.commonDB.SetStatus: %w", err)
	}
	if err := d.datasetsDB.DeleteDatasetData(ctx, datasetID); err != nil {
		return fmt.Errorf("d.datasetsDB.DeleteDatasetData: %w", err)
	}

	return nil
}
