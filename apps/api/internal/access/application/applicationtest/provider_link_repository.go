package applicationtest

import (
	"context"
	"errors"
	"slices"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/access/domain"
)

// ProviderLinkRepository keeps the provider links in memory, unique on the
// account and on the user at each provider, as the table is.
type ProviderLinkRepository struct {
	links   []domain.ProviderLink
	failure error
}

// NewProviderLinkRepository returns a repository with no link.
func NewProviderLinkRepository() *ProviderLinkRepository {
	return &ProviderLinkRepository{}
}

// errLinkedAlready is the unique constraint of the table, broken.
var errLinkedAlready = errors.New("the account, or the user at this provider, is linked already")

// FailWith makes every later Save fail with this error, as a database that
// refuses the write.
func (r *ProviderLinkRepository) FailWith(failure error) {
	r.failure = failure
}

// Find returns a copy of the link of the account.
func (r *ProviderLinkRepository) Find(_ context.Context, provider domain.LoginProviderKey, subject domain.ProviderSubject) (*domain.ProviderLink, bool, error) {
	position := slices.IndexFunc(r.links, func(stored domain.ProviderLink) bool {
		return stored.Provider() == provider && stored.Subject() == subject
	})
	if position < 0 {
		return nil, false, nil
	}
	found := r.links[position]
	return &found, true, nil
}

// HasLink tells whether the user has an account of the provider linked.
func (r *ProviderLinkRepository) HasLink(_ context.Context, provider domain.LoginProviderKey, user domain.UserID) (bool, error) {
	return slices.ContainsFunc(r.links, func(stored domain.ProviderLink) bool {
		return stored.Provider() == provider && stored.User() == user
	}), nil
}

// Save stores a copy of a new link.
func (r *ProviderLinkRepository) Save(ctx context.Context, link *domain.ProviderLink) error {
	if r.failure != nil {
		return r.failure
	}
	_, linked, _ := r.Find(ctx, link.Provider(), link.Subject())
	userLinked, _ := r.HasLink(ctx, link.Provider(), link.User())
	if linked || userLinked {
		return errLinkedAlready
	}
	r.links = append(r.links, *link)
	return nil
}

// All returns a copy of every link, in the order they were saved.
func (r *ProviderLinkRepository) All() []domain.ProviderLink {
	return slices.Clone(r.links)
}

// Snapshot returns the function that puts the links back as they are now.
func (r *ProviderLinkRepository) Snapshot() func() {
	kept := slices.Clone(r.links)
	return func() { r.links = kept }
}
