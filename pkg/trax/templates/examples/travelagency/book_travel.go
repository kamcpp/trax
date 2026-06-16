package travelagency

import (
	"context"
	"fmt"

	"github.com/xshyft/trax/pkg/trax"
)

type BookTravelParams struct {
	RefRequestId string `json:"ref_request_id"`
	AuxData      string `json:"aux_data"`
}

func CreateBookTravelSagaTemplates(ctx context.Context, store trax.Store) error {
	err := store.BeginTransaction(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	sagaTemplate := &trax.SagaTemplate{
		TemplateId:  "book_travel",
		DisplayName: "Book Travel",
		Description: "Saga for booking travel",
		Labels: map[string]string{
			"short_id": "bt",
		},
		Tags:     []string{"examples", "saga", "travel", "agency", "booking"},
		Metadata: map[string]string{},
		SagaStepTemplateIds: []string{
			"complete_payment",
			"book_flight",
			"book_hotel_room",
			"book_rental_car",
		},
	}
	_, err = store.SaveSagaTemplateIdempotently(ctx, sagaTemplate)
	if err != nil {
		err2 := store.RollbackTransaction(ctx)
		if err2 != nil {
			return fmt.Errorf("failed to rollback transaction: %v, err: %v", err2, err)
		}
		return fmt.Errorf("failed to save saga template: %v", err)
	}
	firstSagaStepTemplate := &trax.SagaStepTemplate{
		TemplateId:     "complete_payment",
		SagaTemplateId: "book_travel",
		DisplayName:    "Complete Payment",
		Description:    "Step for completing the payment",
		Labels: map[string]string{
			"short_id": "cp",
		},
		Tags: []string{"examples", "saga", "travel", "agency", "booking", "payment"},
		Metadata: map[string]string{
			"index":           "1",
			"payment_gateway": "stripe",
			"stripe_currency": "USD",
			"stripe_key":      "sk_test_4eC39HqLyjWDarjtT1zdp7dc",
			"raw_amount":      "10000",
		},
	}
	_, err = store.SaveSagaStepTemplateIdempotently(ctx, firstSagaStepTemplate)
	if err != nil {
		err2 := store.RollbackTransaction(ctx)
		if err2 != nil {
			return fmt.Errorf("failed to rollback transaction: %v, err: %v", err2, err)
		}
		return fmt.Errorf("failed to save first saga step template: %v", err)
	}
	secondSagaStepTemplate := &trax.SagaStepTemplate{
		TemplateId:     "book_flight",
		SagaTemplateId: "book_travel",
		DisplayName:    "Book Flight",
		Description:    "Step for booking a flight",
		Labels: map[string]string{
			"short_id": "bf",
		},
		Tags: []string{"examples", "saga", "travel", "agency", "booking", "flight"},
		Metadata: map[string]string{
			"index":           "2",
			"default_airline": "Delta",
			"seat_class":      "economy",
		},
	}
	_, err = store.SaveSagaStepTemplateIdempotently(ctx, secondSagaStepTemplate)
	if err != nil {
		err2 := store.RollbackTransaction(ctx)
		if err2 != nil {
			return fmt.Errorf("failed to rollback transaction: %v, err: %v", err2, err)
		}
		return fmt.Errorf("failed to save second saga step template: %v", err)
	}
	thirdSagaStepTemplate := &trax.SagaStepTemplate{
		TemplateId:     "book_hotel_room",
		SagaTemplateId: "book_travel",
		DisplayName:    "Book Hotel Room",
		Description:    "Step for booking a hotel room",
		Labels: map[string]string{
			"short_id": "bhr",
		},
		Tags: []string{"examples", "saga", "travel", "agency", "booking", "hotel"},
		Metadata: map[string]string{
			"index":             "3",
			"default_hotel":     "Marriott",
			"room_type":         "standard",
			"include_breakfast": "true",
		},
	}
	_, err = store.SaveSagaStepTemplateIdempotently(ctx, thirdSagaStepTemplate)
	if err != nil {
		err2 := store.RollbackTransaction(ctx)
		if err2 != nil {
			return fmt.Errorf("failed to rollback transaction: %v, err: %v", err2, err)
		}
		return fmt.Errorf("failed to save third saga step template: %v", err)
	}
	fourthSagaStepTemplate := &trax.SagaStepTemplate{
		TemplateId:     "book_rental_car",
		SagaTemplateId: "book_travel",
		DisplayName:    "Book Rental Car",
		Description:    "Step for booking a rental car",
		Labels: map[string]string{
			"short_id": "brc",
		},
		Tags: []string{"examples", "saga", "travel", "agency", "booking", "car", "rental"},
		Metadata: map[string]string{
			"index":             "4",
			"default_company":   "Hertz",
			"car_type":          "sedan",
			"rental_days":       "3",
			"include_insurance": "true",
		},
	}
	_, err = store.SaveSagaStepTemplateIdempotently(ctx, fourthSagaStepTemplate)
	if err != nil {
		err2 := store.RollbackTransaction(ctx)
		if err2 != nil {
			return fmt.Errorf("failed to rollback transaction: %v, err: %v", err2, err)
		}
		return fmt.Errorf("failed to save fourth saga step template: %v", err)
	}
	err = store.CommitTransaction(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}
	return nil
}
