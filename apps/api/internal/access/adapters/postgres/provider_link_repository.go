package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/access/domain"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database"
)

// ProviderLinkRepository keeps the provider links in the provider_links
// table, which is unique on the account and on the user at each provider. A
// link is never changed, so nothing here locks.
type ProviderLinkRepository struct {
	pool *pgxpool.Pool
}

// NewProviderLinkRepository returns the repository on this pool.
func NewProviderLinkRepository(pool *pgxpool.Pool) *ProviderLinkRepository {
	return &ProviderLinkRepository{pool: pool}
}

func (r *ProviderLinkRepository) queries(ctx context.Context) *Queries {
	return New(database.Within(ctx, r.pool))
}

// Find returns the link of the account at the provider, and false when the
// account was never linked.
func (r *ProviderLinkRepository) Find(ctx context.Context, provider domain.LoginProviderKey, subject domain.ProviderSubject) (*domain.ProviderLink, bool, error) {
	row, err := r.queries(ctx).FindProviderLink(ctx, FindProviderLinkParams{ProviderKey: provider.String(), Subject: subject.String()})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("find a provider link: %w", err)
	}
	link, err := domain.RestoreProviderLink(domain.StoredProviderLink{
		Provider: row.ProviderKey,
		Subject:  row.Subject,
		User:     row.UserID,
		LinkedAt: row.LinkedAt.UTC(),
	})
	if err != nil {
		return nil, false, fmt.Errorf("read a provider link: %w", err)
	}
	return link, true, nil
}

// HasLink tells whether the user has an account of the provider linked.
func (r *ProviderLinkRepository) HasLink(ctx context.Context, provider domain.LoginProviderKey, user domain.UserID) (bool, error) {
	has, err := r.queries(ctx).HasProviderLink(ctx, HasProviderLinkParams{ProviderKey: provider.String(), UserID: int64(user)})
	if err != nil {
		return false, fmt.Errorf("find the provider link of a user: %w", err)
	}
	return has, nil
}

// Save inserts a new link. A second link of the account, or of the user at
// the provider, breaks a unique constraint of the table.
func (r *ProviderLinkRepository) Save(ctx context.Context, link *domain.ProviderLink) error {
	err := r.queries(ctx).InsertProviderLink(ctx, InsertProviderLinkParams{
		ProviderKey: link.Provider().String(),
		Subject:     link.Subject().String(),
		UserID:      int64(link.User()),
		LinkedAt:    link.LinkedAt(),
	})
	if err != nil {
		return fmt.Errorf("save a provider link: %w", err)
	}
	return nil
}
