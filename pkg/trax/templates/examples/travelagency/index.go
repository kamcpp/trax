package travelagency

import (
	"context"

	"github.com/xshyft/trax/pkg/trax"
)

func CreateSagaTemplates(ctx context.Context, store trax.Store) error {
	err := CreateBookTravelSagaTemplates(ctx, store)
	if err != nil {
		return err
	}
	return nil
}
