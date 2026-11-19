package postgres_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/access/adapters/postgres"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/access/domain"
)

func linkOf(t *testing.T, provider, subject string, user domain.UserID) *domain.ProviderLink {
	t.Helper()
	key, err := domain.NewLoginProviderKey(provider)
	require.NoError(t, err)
	account, err := domain.NewProviderSubject(subject)
	require.NoError(t, err)
	return domain.LinkProvider(key, account, user, noon)
}

// link saves the link, and tells whether it was saved.
func (s stored) link(t *testing.T, link *domain.ProviderLink) bool {
	t.Helper()
	saved, err := postgres.NewProviderLinkRepository(s.pool).Save(context.Background(), link)
	require.NoError(t, err)
	return saved
}

func (s stored) findLink(t *testing.T, provider, subject string) (*domain.ProviderLink, bool) {
	t.Helper()
	found, ok, err := postgres.NewProviderLinkRepository(s.pool).Find(context.Background(), domain.LoginProviderKey(provider), domain.ProviderSubject(subject))
	require.NoError(t, err)
	return found, ok
}

func (s stored) hasLink(t *testing.T, provider string, user domain.UserID) bool {
	t.Helper()
	has, err := postgres.NewProviderLinkRepository(s.pool).HasLink(context.Background(), domain.LoginProviderKey(provider), user)
	require.NoError(t, err)
	return has
}

func TestAProviderLinkIsFoundAsItWasSaved(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	ana := temporary.aUser(t, "ana@fabrica.example")
	saved := linkOf(t, "local", "the-account-of-ana", ana)
	require.True(t, temporary.link(t, saved))

	found, ok := temporary.findLink(t, "local", "the-account-of-ana")

	require.True(t, ok)
	assert.Equal(t, saved, found)
	assert.True(t, temporary.hasLink(t, "local", ana))
}

func TestNoLinkIsFoundForAnAccountNeverLinked(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	ana := temporary.aUser(t, "ana@fabrica.example")
	require.True(t, temporary.link(t, linkOf(t, "local", "the-account-of-ana", ana)))
	bruno := temporary.aUser(t, "bruno@fabrica.example")

	_, otherSubject := temporary.findLink(t, "local", "another-account")
	_, otherProvider := temporary.findLink(t, "google", "the-account-of-ana")

	assert.False(t, otherSubject)
	assert.False(t, otherProvider)
	assert.False(t, temporary.hasLink(t, "google", ana))
	assert.False(t, temporary.hasLink(t, "local", bruno))
}

// The pair of provider and subject is unique: an account is linked to one
// user, and a second link of it is not saved.
func TestASecondLinkOfAnAccountIsNotSaved(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	ana := temporary.aUser(t, "ana@fabrica.example")
	bruno := temporary.aUser(t, "bruno@fabrica.example")
	require.True(t, temporary.link(t, linkOf(t, "local", "one-account", ana)))

	saved := temporary.link(t, linkOf(t, "local", "one-account", bruno))

	assert.False(t, saved)
	found, _ := temporary.findLink(t, "local", "one-account")
	assert.Equal(t, ana, found.User())
}

// A user has one account of each provider linked, and may have one at each
// provider; the same subject at two providers names two accounts.
func TestASecondAccountOfOneProviderForAUserIsNotSaved(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	ana := temporary.aUser(t, "ana@fabrica.example")
	require.True(t, temporary.link(t, linkOf(t, "local", "one-account", ana)))

	secondAtLocal := temporary.link(t, linkOf(t, "local", "another-account", ana))
	atGoogle := temporary.link(t, linkOf(t, "google", "one-account", ana))

	assert.False(t, secondAtLocal)
	assert.True(t, atGoogle)
	_, kept := temporary.findLink(t, "local", "another-account")
	assert.False(t, kept)
}

// Two first logins of one account at the same moment: the second waits for
// the transaction of the first, saves nothing once it commits, and reads the
// link the first saved in its own transaction, which goes on (ticket 07).
func TestALinkSavedMeanwhileByAnotherLoginIsReadAgainInTheSameTransaction(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	ana := temporary.aUser(t, "ana@fabrica.example")
	links := postgres.NewProviderLinkRepository(temporary.pool)
	type secondLogin struct {
		saved bool
		found *domain.ProviderLink
		err   error
	}
	second := make(chan secondLogin, 1)

	err := temporary.transactions.Run(context.Background(), func(ctx context.Context) error {
		saved, err := links.Save(ctx, linkOf(t, "local", "one-account", ana))
		require.True(t, saved)
		go func() {
			var outcome secondLogin
			outcome.err = temporary.transactions.Run(context.Background(), func(ctx context.Context) error {
				var err error
				outcome.saved, err = links.Save(ctx, linkOf(t, "local", "one-account", ana))
				if err != nil {
					return err
				}
				outcome.found, _, err = links.Find(ctx, "local", "one-account")
				return err
			})
			second <- outcome
		}()
		temporary.waitForATransactionThatWaits(t)
		return err
	})

	require.NoError(t, err)
	outcome := <-second
	require.NoError(t, outcome.err)
	assert.False(t, outcome.saved, "the second login saved the link again")
	require.NotNil(t, outcome.found, "the second login did not read the link of the first")
	assert.Equal(t, ana, outcome.found.User())
}

func TestALinkSavedInATransactionThatFailsIsNotKept(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	ana := temporary.aUser(t, "ana@fabrica.example")
	failed := errors.New("a later step failed")

	err := temporary.transactions.Run(context.Background(), func(ctx context.Context) error {
		if _, err := postgres.NewProviderLinkRepository(temporary.pool).Save(ctx, linkOf(t, "local", "one-account", ana)); err != nil {
			return err
		}
		return failed
	})

	require.ErrorIs(t, err, failed)
	_, kept := temporary.findLink(t, "local", "one-account")
	assert.False(t, kept)
}

// The login never changes or removes a link, and the API role cannot either.
func TestTheAPIRoleCannotChangeOrRemoveALink(t *testing.T) {
	t.Parallel()
	temporary := newStored(t)
	ana := temporary.aUser(t, "ana@fabrica.example")
	require.True(t, temporary.link(t, linkOf(t, "local", "one-account", ana)))
	const insufficientPrivilege = "42501"

	for _, statement := range []string{
		`UPDATE provider_links SET subject = 'another-account'`,
		`DELETE FROM provider_links`,
	} {
		_, err := temporary.pool.Exec(context.Background(), statement)

		var failure *pgconn.PgError
		require.ErrorAs(t, err, &failure, statement)
		assert.Equal(t, insufficientPrivilege, failure.Code, statement)
	}
}
