package builder

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"

	"github.com/hse-experiments-platform/datasets/internal/pkg/models"
	"github.com/hse-experiments-platform/datasets/internal/pkg/storage/datasets"
	"github.com/rs/zerolog/log"
)

func (b *Builder) initSchema(columns []string) {
	for _, c := range columns {
		b.Schema = append(b.Schema, models.Column{
			Name: c,
			Type: models.ColumnTypeUndefined,
		})
	}

	log.Debug().Any("schema", b.Schema).Msgf("initialized schema")
}

func (b *Builder) ProcessChunk(data []byte) error {
	var (
		parsingData       = append(b.prevRowPrefix, data...)
		r                 = csv.NewReader(bytes.NewReader(parsingData))
		prevOffset        = int64(0)
		firstRow, nextRow = b.curRow, b.curRow
		prefixLen         = -1
	)

	r.ReuseRecord = true
	b.prevRowPrefix = nil

	for {
		values, err := r.Read()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil && r.InputOffset() == int64(len(parsingData)) ||
			(err == nil && r.InputOffset() > 0 && parsingData[r.InputOffset()-1] != '\n') {
			//log.Info().Msg("got non full row in the end of the chunk")
			if prefixLen == -1 {
				prefixLen = len(data)
			}
			b.prevRowPrefix = parsingData[prevOffset:]
			break
		} else if err != nil {
			log.Error().Str("data", string(data[:1000])).Err(err).Msg("cannot read csv row")
			return err
		}

		if b.Schema == nil {
			b.initSchema(values)
		}
		if prefixLen == -1 {
			prefixLen = int(r.InputOffset()) - len(parsingData) + len(data)
		}
		prevOffset = r.InputOffset()
		nextRow++
	}

	lastChunkRow := nextRow
	// в случае если мы прочитали ровно без остатка, то nextRow будет номером следующей строки, поэтому в текущей
	// строке номер последней строки на один меньше, а если прочитали не всю строку, то следующая строка это та же,
	// которая последняя в чанке
	if len(b.prevRowPrefix) == 0 {
		lastChunkRow--
	}

	//log.Debug().Msg("got chunk to write in db")
	if err := b.DatasetsDB.UploadDatasetChunk(b.Ctx, datasets.UploadDatasetChunkParams{
		DatasetID:    b.DatasetID,
		ChunkNumber:  b.chunkNumber,
		ChunkLen:     int32(len(data)),
		MinRowNumber: firstRow,
		MaxRowNumber: lastChunkRow,
		PrefixLen:    int32(prefixLen),
	}); err != nil {
		return fmt.Errorf("b.DatasetsDB.UploadDatasetChunk: %w", err)
	}

	b.chunkNumber++
	b.curRow = nextRow
	return nil
}

func (b *Builder) GetRowsCount() int64 {
	if b.Schema == nil || len(b.prevRowPrefix) > 0 {
		return -1
	}

	return b.curRow - 1 // curRow будет номером следующей строки после окончания обработки
}
