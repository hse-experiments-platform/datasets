package datasets

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"slices"

	"github.com/hse-experiments-platform/datasets/internal/pkg/storage/datasetsdb"
	pb "github.com/hse-experiments-platform/datasets/pkg/datasets"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

func (d *datasetsService) getChunks(ctx context.Context, offset, limit int64, datasetID int64, borders datasetsdb.GetDatasetChunkBordersRow, reader *csv.Reader, buff *bytes.Buffer) ([][]string, error) {
	if borders.MaxRowNumber[len(borders.MaxRowNumber)-1] < int64(offset) {
		return nil, nil
	} else if borders.MaxRowNumber[len(borders.MaxRowNumber)-1]+1 < offset+limit {
		limit = borders.MaxRowNumber[len(borders.MaxRowNumber)-1] + 1 - offset
	}

	l, _ := slices.BinarySearch(borders.MaxRowNumber, offset)
	r, _ := slices.BinarySearch(borders.MinRowNumber, offset+limit)

	// get chunks with data
	chunks, err := d.datasetsDB.GetDatasetChunks(ctx, datasetsdb.GetDatasetChunksParams{
		DatasetID: datasetID,
		L:         int64(l),
		R:         int64(r),
	})
	if err != nil {
		return nil, fmt.Errorf("d.datasetsDB.GetDatasetChunks: %w", err)
	} else if len(chunks) == 0 {
		log.Error().Int64("datasetID", datasetID).Int("left_index", l).Int("right_index", r).
			Msg("got no chunks")
		return nil, fmt.Errorf("unexpected result from GetDatasetChunks: no chunks got")
	}

	if err != nil {
		return nil, fmt.Errorf("d.datasetsDB.GetDatasetChunks: %w", err)
	} else if len(chunks) == 0 {
		log.Error().Int64("datasetID", datasetID).Int("left_index", l).Int("right_index", r).
			Msg("got no chunks")
		return nil, fmt.Errorf("unexpected result from GetDatasetChunks: no chunks got")
	}

	if chunks[0].MinRowNumber != offset {
		chunks[0].RawDataChunk = chunks[0].RawDataChunk[chunks[0].PrefixLen:]
		chunks[0].MinRowNumber++
	}

	var data []byte
	for _, chunk := range chunks {
		data = append(data, chunk.RawDataChunk...)
	}

	buff.Write(data)
	var result [][]string
	for i := chunks[0].MinRowNumber; i < offset+limit; i++ {
		log.Debug().Any("aval", buff.Available()).Any("len", buff.Len()).Msg("debug")
		cols, err := reader.Read()
		if err != nil {
			log.Error().Err(err).Int64("index", i).Int64("offset", offset).Int64("index", limit).Msg("cannot read row")
			return nil, fmt.Errorf("reader.Read: %w", err)
		}
		if i < offset {
			continue
		}
		result = append(result, cols)
	}

	_, _ = reader.ReadAll()
	buff.Reset()

	return result, nil
}

func (d *datasetsService) getRows(ctx context.Context, request *pb.GetDatasetRowsRequest, resp *pb.GetDatasetRowsResponse) func(tx pgx.Tx) error {
	return func(tx pgx.Tx) error {
		borders, err := d.datasetsDB.GetDatasetChunkBorders(ctx, request.GetDatasetID())
		if errors.Is(err, pgx.ErrNoRows) {
			return status.Error(codes.InvalidArgument, "dataset is not uploaded yet")
		} else if err != nil {
			return fmt.Errorf("d.datasetsDB.GetDatasetChunkBorders: %w", err)
		} else if len(borders.MaxRowNumber) == 0 || len(borders.MinRowNumber) == 0 {
			resp = &pb.GetDatasetRowsResponse{
				PageInfo: &pb.PageInfo{
					Offset: request.GetOffset(),
					Limit:  request.GetLimit(),
					Total:  0,
				},
			}
			return nil
		}

		*resp = pb.GetDatasetRowsResponse{
			PageInfo: &pb.PageInfo{
				Offset: request.GetOffset(),
				Limit:  request.GetLimit(),
				Total:  uint64(borders.MaxRowNumber[len(borders.MaxRowNumber)-1]),
			},
		}

		buff := bytes.NewBuffer(nil)
		reader := csv.NewReader(buff)

		schema, err := d.getChunks(ctx, 0, 1, request.GetDatasetID(), borders, reader, buff)
		if err != nil {
			return fmt.Errorf("schema d.getChunks: %w", err)
		} else if len(schema) != 1 {
			return fmt.Errorf("wrong result len: d.getChunks, expected 1, got %v", len(schema))
		}
		resp.Schema = &pb.DatasetSchema{Columns: make([]*pb.DatasetSchema_SchemaColumn, 0, len(schema[0]))}
		for _, c := range schema[0] {
			resp.Schema.Columns = append(resp.Schema.Columns, &pb.DatasetSchema_SchemaColumn{
				Name: c,
				Type: "string",
			})
		}

		// переводим в 1-нумерцию так как 0 строчка это названия колонок
		offset := request.GetOffset() + 1
		limit := request.GetLimit()
		rows, err := d.getChunks(ctx, int64(offset), int64(limit), request.GetDatasetID(), borders, reader, buff)
		if err != nil {
			return fmt.Errorf("rows d.getChunks: %w", err)
		}

		for i, row := range rows {
			resp.Rows = append(resp.Rows, &pb.DatasetRow{
				RowNumber: request.GetOffset() + uint64(i),
				Columns:   make([]*structpb.Value, 0, len(row)),
			})
			for _, col := range row {
				resp.Rows[i].Columns = append(resp.Rows[i].Columns, structpb.NewStringValue(col))
			}
		}

		return nil
	}
}

func (d *datasetsService) checkDatasetAccess(ctx context.Context, datasetID int64) error {
	userID, err := getUserID(ctx)
	if err != nil {
		return err
	}

	creatorID, err := d.commonDB.GetDatasetCreator(ctx, datasetID)
	if errors.Is(err, pgx.ErrNoRows) {
		return status.Error(codes.NotFound, "dataset not found")
	} else if err != nil {
		return fmt.Errorf("d.commonDB.GetDatasetCreator: %w", err)
	}

	if creatorID != userID {
		return status.Error(codes.PermissionDenied, "cannot get other user's dataset rows")
	}

	return nil
}

func (d *datasetsService) GetDatasetRows(ctx context.Context, request *pb.GetDatasetRowsRequest) (*pb.GetDatasetRowsResponse, error) {
	if request.GetDatasetID() == 0 {
		return nil, status.Error(codes.InvalidArgument, "datasetID must be not 0")
	}
	if request.GetLimit() == 0 {
		return nil, status.Error(codes.InvalidArgument, "limit must be greater than 0")
	}

	if err := d.checkDatasetAccess(ctx, request.GetDatasetID()); err != nil {
		return nil, err
	}

	resp := &pb.GetDatasetRowsResponse{}
	if err := pgx.BeginTxFunc(ctx, d.datasetsDBConn, pgx.TxOptions{
		AccessMode: pgx.ReadOnly,
	}, d.getRows(ctx, request, resp)); err != nil {
		return nil, fmt.Errorf("pgx.BeginTxFunc: %w", err)
	}

	return resp, nil
}

// "6","Judgment Night","6.527","316","Released","1993-10-15","12136938.0","109","21000000.0","tt0107286","en","Judgment Night","Four young friends, while taking a shortcut en route to a local boxing match, witness a brutal murder which leaves them running for their lives.","11.92","Don't move. Don't whisper. Don't even breathe.","Action, Crime, Thriller","Largo Entertainment, JVC, Universal Pictures","United States of America","English","Everlast, Peter Greene, Michael DeLorenzo, Denis Leary, Stephen Dorff, Galyn Görg, Eugene Williams, Will Zahrn, Emilio Estevez, Cuba Gooding Jr., Michael Wiseman, Jeremy Piven, Relioues Webb, Angela Alvarado, Christine Harnos","Stephen Hopkins","Peter Levy","Jere Cunningham, Lewis Colick","Marilyn Vance, Gene Levy, Lloyd Segan","Alan Silvestri"
