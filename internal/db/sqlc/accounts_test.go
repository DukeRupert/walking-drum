package sqlc_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/dukerupert/walking-drum/internal/db/sqlc"
	"github.com/dukerupert/walking-drum/internal/testdb"
)

func TestAccountsCRUD(t *testing.T) {
	q, _ := testdb.WithTx(t)
	ctx := context.Background()

	id, err := uuid.NewV7()
	if err != nil {
		t.Fatalf("uuid: %v", err)
	}
	pgID := pgtype.UUID{Bytes: id, Valid: true}

	acc, err := q.CreateAccount(ctx, sqlc.CreateAccountParams{
		ID:           pgID,
		Email:        "alice@example.com",
		DisplayName:  "Alice",
		PasswordHash: "not-a-real-hash",
	})
	if err != nil {
		t.Fatalf("CreateAccount: %v", err)
	}
	if acc.Status != "active" {
		t.Errorf("default status: got %q, want active", acc.Status)
	}
	if !acc.CreatedAt.Valid {
		t.Error("created_at should be set by default")
	}

	byID, err := q.GetAccountByID(ctx, pgID)
	if err != nil {
		t.Fatalf("GetAccountByID: %v", err)
	}
	if byID.Email != "alice@example.com" {
		t.Errorf("GetAccountByID email: got %q, want alice@example.com", byID.Email)
	}

	// CITEXT means email lookups are case-insensitive.
	byEmail, err := q.GetAccountByEmail(ctx, "ALICE@example.com")
	if err != nil {
		t.Fatalf("GetAccountByEmail (mixed case): %v", err)
	}
	if byEmail.ID != pgID {
		t.Error("case-insensitive email lookup returned wrong account")
	}

	updated, err := q.UpdateAccountStatus(ctx, sqlc.UpdateAccountStatusParams{
		ID:     pgID,
		Status: "suspended",
	})
	if err != nil {
		t.Fatalf("UpdateAccountStatus: %v", err)
	}
	if updated.Status != "suspended" {
		t.Errorf("status after update: got %q, want suspended", updated.Status)
	}

	if err := q.SoftDeleteAccount(ctx, pgID); err != nil {
		t.Fatalf("SoftDeleteAccount: %v", err)
	}
	if _, err := q.GetAccountByID(ctx, pgID); err == nil {
		t.Error("GetAccountByID should fail after soft-delete (deleted_at IS NOT NULL)")
	}
}

func TestAccountStatusCheckConstraint(t *testing.T) {
	q, _ := testdb.WithTx(t)
	ctx := context.Background()

	id, _ := uuid.NewV7()
	acc, err := q.CreateAccount(ctx, sqlc.CreateAccountParams{
		ID:           pgtype.UUID{Bytes: id, Valid: true},
		Email:        "bob@example.com",
		DisplayName:  "Bob",
		PasswordHash: "x",
	})
	if err != nil {
		t.Fatalf("CreateAccount: %v", err)
	}

	if _, err := q.UpdateAccountStatus(ctx, sqlc.UpdateAccountStatusParams{
		ID:     acc.ID,
		Status: "made-up-status",
	}); err == nil {
		t.Fatal("expected CHECK constraint violation, got nil")
	}
}
