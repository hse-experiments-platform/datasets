package datasets

import (
	"context"
	"errors"
	"fmt"

	"github.com/hse-experiments-platform/library/pkg/utils/token"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func getUserID(ctx context.Context) (int64, error) {
	var userID int64
	userID, ok := ctx.Value(token.UserIDContextKey).(int64)
	if !ok {
		log.Error().Msgf("invalid userID context key type: %T", ctx.Value(token.UserIDContextKey))
		return 0, status.New(codes.Internal, "internal error").Err()
	}

	return userID, nil
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
