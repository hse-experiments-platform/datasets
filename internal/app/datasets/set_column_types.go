package datasets

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/hse-experiments-platform/datasets/internal/pkg/domain/dataset"
	"github.com/hse-experiments-platform/datasets/internal/pkg/models"
	"github.com/hse-experiments-platform/datasets/internal/pkg/storage/datasetsdb"
	"github.com/hse-experiments-platform/datasets/internal/pkg/storage/db"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	DatasetConverterCMDPathKey = "DATASET_CONVERTER_CMD_PATH"
	filePath                   = "./converted.csv"
)

func extractToken(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Errorf(codes.Unauthenticated, "metadata is not provided")
	}

	values := md["authorization"]
	if len(values) == 0 {
		return "", status.Errorf(codes.Unauthenticated, "authorization token is not provided")
	}

	return strings.TrimPrefix(values[0], "Bearer "), nil
}

func constructPythonListLiteral(params db.SetDatasetSchemaParams) string {
	b := strings.Builder{}
	b.WriteString("[")
	for i, name := range params.ColumnNames {
		b.WriteString("('")
		b.WriteString(name)
		b.WriteString("','")
		b.WriteString(params.ColumnTypes[i])
		b.WriteString("')")
		if i != len(params.ColumnNames)-1 {
			b.WriteRune(',')
		}
	}
	b.WriteString("]")

	// desired format
	//[('abc','int'),('def','float'),('gfj','dropped')]
	return b.String()
}

func (d *datasetsService) setColumnTypes(userCtx context.Context, params db.SetDatasetSchemaParams, rowsCount int64) error {
	path := os.Getenv(DatasetConverterCMDPathKey)
	if len(path) == 0 {
		return fmt.Errorf("no converter path in environment")
	}
	ctx := context.Background()
	done := make(chan bool)

	token, err := extractToken(userCtx)
	if err != nil {
		return err
	}

	args := []string{path, fmt.Sprint(params.DatasetID), fmt.Sprint(rowsCount), constructPythonListLiteral(params),
		token, filePath, "localhost:8084"}

	cmd := exec.CommandContext(ctx, "python3", args...)

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("cannot start datasest converter command: %w", err)
	}

	go func() {
		var (
			err         error
			chunksCount int
		)
		defer func() {
			if err != nil {
				log.Error().Err(err).Msg("failed set column types")
				d.onConvartationFail(ctx, params)
			} else {
				d.onConvertationSuccess(ctx, params, chunksCount)
			}
		}()

		go func() {
			select {
			case <-time.NewTimer(time.Second * 60 * 5).C:
				err = cmd.Cancel()
				if err != nil {
					log.Error().Err(err).Msg("cannot cancel dataset converter on timeout")
				}

				err = fmt.Errorf("command cancelled due to timeout")
			case <-done:
				return
			}
		}()

		err = cmd.Wait()
		done <- true
		if err != nil {
			log.Error().Err(err).Str("out", string(out.Bytes())).Msg("command finished with error")
			return
		}

		file, err := os.Open(filePath)
		if err != nil {
			err = fmt.Errorf("cannot open converted file: %w", err)
			return
		}
		defer file.Close()

		builder := dataset.NewBuilder(ctx, d.datasetsDB, params.DatasetID)
		buffer := make([]byte, bufferSize)
		for {
			var n int
			n, err = file.Read(buffer)
			if err == io.EOF {
				if err := builder.ProcessChunk(append(buffer[:n], '\n')); err != nil {
					err = fmt.Errorf("builder.ProcessChunk: %w", err)
					return
				}
				chunksCount++
				err = nil
				break
			} else if err != nil {
				err = fmt.Errorf("file.Read: %w", err)
				return
			}

			err = builder.ProcessChunk(buffer[:n])
			if err != nil {
				err = fmt.Errorf("builder.ProcessChunk: %w", err)
				return
			}
			chunksCount++
		}

	}()

	return nil
}

func (d *datasetsService) onConvertationSuccess(ctx context.Context, params db.SetDatasetSchemaParams, count int) {
	err := d.datasetsDB.DeleteOldDatasetData(ctx, datasetsdb.DeleteOldDatasetDataParams{
		DatasetID:   params.DatasetID,
		ChunkNumber: int64(count),
	})
	if err != nil {
		log.Error().Err(err).Msg("d.datasetsDB.DeleteOldDatasetData")
	}

	err = d.commonDB.DropColumnsByType(ctx, db.DropColumnsByTypeParams{
		DatasetID:  params.DatasetID,
		ColumnType: models.TypeToString[models.ColumnTypeDropped],
	})
	if err != nil {
		log.Error().Err(err).Msg("d.commonDB.DropColumnsByType")
	}

	err = d.commonDB.SetStatus(ctx, db.SetStatusParams{
		ID:     params.DatasetID,
		Status: db.DatasetStatusReady,
	})
	if err != nil {
		log.Error().Err(err).Msg("d.commonDB.SetStatus")
	}
}

func (d *datasetsService) onConvartationFail(ctx context.Context, params db.SetDatasetSchemaParams) {
	err := d.commonDB.SetStatus(ctx, db.SetStatusParams{
		ID:     params.DatasetID,
		Status: db.DatasetStatusConvertationError,
	})
	if err != nil {
		log.Error().Err(err).Msg("d.commonDB.SetStatus")
	}
}
