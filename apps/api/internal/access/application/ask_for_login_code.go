package application

import (
	"cmp"
	"context"
	"errors"
	"io"
	"time"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/access/domain"
	consultancyapplication "github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/application"
	emailapplication "github.com/dannilodms/software-factory-manager/apps/api/internal/email/application"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/kernel"
)

// LoginCodeDependencies are what the use cases of the login code share: the
// transactions, the requests, the session of the browser, the key the codes
// are computed with, which comes from configuration, and the clock.
type LoginCodeDependencies struct {
	Transactions Transactions
	Requests     LoginCodeRequestRepository
	Session      Session
	Key          domain.LoginCodeKey
	Clock        Clock
}

// AskForLoginCode sends a login code to the e-mail of an active user, and
// answers the same way for every e-mail, so that nobody learns from it who is
// a user (ADR-0011). The e-mail leaves through a job, and every ask runs one
// transaction that counts it, so the time of the answer tells little either.
type AskForLoginCode struct {
	login          LoginCodeDependencies
	asks           AskedAddressRepository
	findActiveUser *consultancyapplication.FindActiveUser
	queueEmail     *emailapplication.QueueEmail
	random         io.Reader
}

// NewAskForLoginCode builds the use case. The random source is the
// cryptographic one of the standard library everywhere but in a test.
func NewAskForLoginCode(
	login LoginCodeDependencies,
	asks AskedAddressRepository,
	findActiveUser *consultancyapplication.FindActiveUser,
	queueEmail *emailapplication.QueueEmail,
	random io.Reader,
) *AskForLoginCode {
	return &AskForLoginCode{login: login, asks: asks, findActiveUser: findActiveUser, queueEmail: queueEmail, random: random}
}

// AskForLoginCodeInput is the e-mail as the person typed it.
type AskForLoginCodeInput struct {
	Email string
}

// Execute asks for a login code. In one transaction it counts the ask for the
// typed address, whoever it belongs to; then, for the e-mail of an active
// user, it ends their earlier requests, saves a new one and queues its e-mail
// with the job that delivers it, and the session of the browser keeps the
// request. For any other e-mail nothing else is saved, and the session keeps
// domain.NoLoginCodeRequest, so the answer is the same.
//
// It answers a kernel.InvalidFieldsError that names `email` when the text is
// not an e-mail address, and domain.LoginCodeLimitReachedError past the limits
// of the address, for every address alike; neither changes anything.
func (u *AskForLoginCode) Execute(ctx context.Context, input AskForLoginCodeInput) error {
	email, err := kernel.NewEmail(input.Email)
	if err != nil {
		var invalid kernel.InvalidFieldsError
		invalid.AddField("email", err)
		return invalid.Err()
	}

	var asked domain.LoginCodeRequestID
	err = u.login.Transactions.Run(ctx, func(ctx context.Context) error {
		asked, err = u.ask(ctx, email)
		return err
	})
	if err != nil {
		return err
	}
	return u.login.Session.RememberLoginCodeRequest(ctx, asked)
}

// ask counts the ask for the address, and asks for a code when the address
// is the e-mail of an active user. It returns the request, or
// domain.NoLoginCodeRequest for any other address.
func (u *AskForLoginCode) ask(ctx context.Context, email kernel.Email) (domain.LoginCodeRequestID, error) {
	now := u.login.Clock.Now()
	address := u.login.Key.AskedAddress(email)

	// Held first, so that two asks for one address at the same moment are
	// counted one after the other: the second meets the limit of a minute.
	if err := u.asks.Hold(ctx, address); err != nil {
		return 0, err
	}
	if err := u.withinLimits(ctx, address, now); err != nil {
		return 0, err
	}
	if err := u.asks.Add(ctx, address, now); err != nil {
		return 0, err
	}
	user, found, err := u.findActiveUser.Execute(ctx, email)
	if err != nil || !found {
		return domain.NoLoginCodeRequest, err
	}
	return u.request(ctx, user, now)
}

// request ends the earlier requests of the user, saves a new one and queues
// the e-mail of its code.
func (u *AskForLoginCode) request(ctx context.Context, user consultancyapplication.ActiveUser, now time.Time) (domain.LoginCodeRequestID, error) {
	owner := domain.UserID(user.ID)
	earlier, err := u.login.Requests.LoadNotEndedOf(ctx, owner)
	if err != nil {
		return 0, err
	}
	if err := u.end(ctx, earlier, now); err != nil {
		return 0, err
	}
	id, err := u.login.Requests.NextIdentifier(ctx)
	if err != nil {
		return 0, err
	}
	request, err := domain.AskForLoginCode(id, owner, now, u.random)
	if err != nil {
		return 0, err
	}
	if err := u.login.Requests.Save(ctx, request); err != nil {
		return 0, err
	}
	err = u.queueEmail.Execute(ctx, emailapplication.QueueEmailInput{
		Recipient: user.Email.String(),
		Template:  emailapplication.LoginCode,
		Language:  user.PreferredLanguage,
		Reference: emailapplication.LoginCodeRequest(int64(id)),
	})
	if err != nil {
		return 0, err
	}
	return id, nil
}

// withinLimits counts the asks for the address in the window of each limit,
// and answers domain.LoginCodeLimitReachedError past one: the first failure
// is the one returned.
func (u *AskForLoginCode) withinLimits(ctx context.Context, address domain.AskedAddress, now time.Time) error {
	var err error
	for _, limit := range domain.LoginCodeLimits() {
		err = cmp.Or(err, u.withinLimit(ctx, address, now, limit))
	}
	return err
}

func (u *AskForLoginCode) withinLimit(ctx context.Context, address domain.AskedAddress, now time.Time, limit domain.LoginCodeLimit) error {
	asked, err := u.asks.CountAskedSince(ctx, address, limit.Since(now))
	if err != nil {
		return err
	}
	return limit.Allows(asked)
}

// end ends the earlier requests of the user: only the last code they received
// works.
func (u *AskForLoginCode) end(ctx context.Context, earlier []*domain.LoginCodeRequest, now time.Time) error {
	var err error
	for _, request := range earlier {
		request.End(now)
		err = errors.Join(err, u.login.Requests.Save(ctx, request))
	}
	return err
}
