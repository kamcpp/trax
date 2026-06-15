package trax

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestInMemStore_UpdateSagaTemplate verifies updating an existing saga template.
func TestInMemStore_UpdateSagaTemplate(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	// Create template
	tmpl := &SagaTemplate{
		TemplateId:          "test-template-1",
		DisplayName:         "Original Name",
		Description:         "Original Desc",
		SagaStepTemplateIds: []string{"step1"},
	}
	isNew, err := store.SaveSagaTemplateIdempotently(ctx, tmpl)
	require.NoError(t, err)
	require.True(t, isNew)

	// Update
	updated := &SagaTemplate{
		TemplateId:          "test-template-1",
		DisplayName:         "Updated Name",
		Description:         "Updated Desc",
		SagaStepTemplateIds: []string{"step1", "step2"},
	}
	err = store.UpdateSagaTemplate(ctx, updated)
	require.NoError(t, err)

	// Verify
	got, err := store.GetSagaTemplate(ctx, "test-template-1")
	require.NoError(t, err)
	require.Equal(t, "Updated Name", got.DisplayName)
	require.Equal(t, "Updated Desc", got.Description)
	require.Len(t, got.SagaStepTemplateIds, 2)
}

// TestInMemStore_UpdateSagaTemplate_NotFound verifies error on non-existent template.
func TestInMemStore_UpdateSagaTemplate_NotFound(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	err := store.UpdateSagaTemplate(ctx, &SagaTemplate{TemplateId: "nonexistent"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

// TestInMemStore_DeleteSagaTemplate verifies deletion with cascade to step templates.
func TestInMemStore_DeleteSagaTemplate(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	// Create template + step
	tmpl := &SagaTemplate{
		TemplateId:          "test-template-del",
		SagaStepTemplateIds: []string{"step-del-1"},
	}
	_, err := store.SaveSagaTemplateIdempotently(ctx, tmpl)
	require.NoError(t, err)

	step := &SagaStepTemplate{
		TemplateId:     "step-del-1",
		SagaTemplateId: "test-template-del",
	}
	_, err = store.SaveSagaStepTemplateIdempotently(ctx, step)
	require.NoError(t, err)

	// Delete template (should cascade to steps)
	err = store.DeleteSagaTemplate(ctx, "test-template-del")
	require.NoError(t, err)

	// Verify template gone
	_, err = store.GetSagaTemplate(ctx, "test-template-del")
	require.Error(t, err)

	// Verify step template also gone
	_, err = store.GetSagaStepTemplate(ctx, "step-del-1")
	require.Error(t, err)
}

// TestInMemStore_DeleteSagaTemplate_NotFound verifies error on non-existent template.
func TestInMemStore_DeleteSagaTemplate_NotFound(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	err := store.DeleteSagaTemplate(ctx, "nonexistent")
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

// TestInMemStore_UpdateSagaStepTemplate verifies updating a step template.
func TestInMemStore_UpdateSagaStepTemplate(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	// Create parent + step
	_, _ = store.SaveSagaTemplateIdempotently(ctx, &SagaTemplate{TemplateId: "parent"})
	_, _ = store.SaveSagaStepTemplateIdempotently(ctx, &SagaStepTemplate{
		TemplateId:     "step1",
		SagaTemplateId: "parent",
		DisplayName:    "Original",
	})

	// Update
	err := store.UpdateSagaStepTemplate(ctx, &SagaStepTemplate{
		TemplateId:     "step1",
		SagaTemplateId: "parent",
		DisplayName:    "Updated",
	})
	require.NoError(t, err)

	got, err := store.GetSagaStepTemplate(ctx, "step1")
	require.NoError(t, err)
	require.Equal(t, "Updated", got.DisplayName)
}

// TestInMemStore_UpdateSagaStepTemplate_NotFound verifies error on non-existent step.
func TestInMemStore_UpdateSagaStepTemplate_NotFound(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	err := store.UpdateSagaStepTemplate(ctx, &SagaStepTemplate{TemplateId: "nonexistent"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

// TestInMemStore_DeleteSagaStepTemplate verifies deleting a step template.
func TestInMemStore_DeleteSagaStepTemplate(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	_, _ = store.SaveSagaStepTemplateIdempotently(ctx, &SagaStepTemplate{
		TemplateId:     "step-to-delete",
		SagaTemplateId: "parent",
	})

	err := store.DeleteSagaStepTemplate(ctx, "step-to-delete")
	require.NoError(t, err)

	_, err = store.GetSagaStepTemplate(ctx, "step-to-delete")
	require.Error(t, err)
}

// TestInMemStore_DeleteSagaStepTemplate_NotFound verifies error on non-existent step.
func TestInMemStore_DeleteSagaStepTemplate_NotFound(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	err := store.DeleteSagaStepTemplate(ctx, "nonexistent")
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}
