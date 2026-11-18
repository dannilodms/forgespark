package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	auditapplication "github.com/dannilodms/software-factory-manager/apps/api/internal/audit/application"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/application"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/domain"
)

// concluded is an installation whose setup was concluded with the first code,
// ten minutes after noon, with the clock ten minutes later still. Its users
// are the first administrator, and no setup code works.
func concluded(t *testing.T) installation {
	t.Helper()
	installed := issued(t)
	require.NoError(t, installed.concludeSetup.Execute(context.Background(), typedSetup(firstCode)))
	installed.clock.Advance(10 * time.Minute)
	return installed
}

// backIn is an installation whose setup was concluded, with the second code
// issued at twenty past noon to get back in, and the clock five minutes
// later.
func backIn(t *testing.T) installation {
	t.Helper()
	installed := concluded(t)
	_, err := installed.issueSetupCode.Execute(context.Background())
	require.NoError(t, err)
	installed.clock.Advance(5 * time.Minute)
	return installed
}

// typedAdministrator is what a person types to register a new administrator,
// every field filled in, with the code they were given.
func typedAdministrator(code string) application.RegisterAdministratorInput {
	return application.RegisterAdministratorInput{
		Code: code,
		Administrator: domain.TypedUser{
			Name:              "Bruno Lima",
			Email:             "Bruno.Lima@Fabrica.example",
			Country:           "PT",
			NationalID:        "123 456 789",
			Telephone:         "+351 21 000 1234",
			PreferredLanguage: "es",
		},
	}
}

func TestRegisteringAnAdministratorCreatesAUserAsTypedWithTheAdministratorRole(t *testing.T) {
	installed := backIn(t)

	err := installed.registerAdministrator.Execute(context.Background(), typedAdministrator(secondCode))

	require.NoError(t, err)
	users := installed.users.All()
	require.Len(t, users, 2)
	registered := users[1]
	details := registered.Details()
	assert.Equal(t, "Bruno Lima", details.Name())
	assert.Equal(t, "bruno.lima@fabrica.example", details.Email().String(), "the e-mail is stored in lowercase")
	assert.Equal(t, "PT", details.Country().String())
	assert.Equal(t, "123 456 789", details.NationalID(), "the national ID is stored as typed")
	telephone, hasTelephone := details.Telephone()
	assert.True(t, hasTelephone)
	assert.Equal(t, "+351210001234", telephone.String())
	assert.Equal(t, "es", details.PreferredLanguage().String())
	assert.False(t, registered.Deactivated())
	administratorRole := roleWithKey(t, installed, domain.AdministratorRoleKey)
	assert.Equal(t, []domain.RoleID{administratorRole.ID()}, registered.Roles())
}

func TestRegisteringAnAdministratorEndsTheSetupCodeAndRecordsTheRegistration(t *testing.T) {
	installed := backIn(t)
	entriesBefore := len(installed.auditLog.Entries())

	err := installed.registerAdministrator.Execute(context.Background(), typedAdministrator(secondCode))

	require.NoError(t, err)
	setupCodes := installed.setupCodes.All()
	require.Len(t, setupCodes, 2)
	endedAt, ended := setupCodes[1].EndedAt()
	assert.True(t, ended, "the setup code still works")
	assert.Equal(t, noon.Add(25*time.Minute), endedAt)
	_, err = installed.openSetup.Execute(context.Background(), application.OpenSetupInput{Code: secondCode})
	requireSetupCodeInvalid(t, err)
	entries := installed.auditLog.Entries()
	require.Len(t, entries, entriesBefore+1, "one entry for the registration")
	recorded := entries[len(entries)-1]
	assert.Equal(t, auditapplication.AdministratorRegistered, recorded.Operation())
	assert.Equal(t, auditapplication.SetupCodeUsed(int64(setupCodes[1].ID())), recorded.Actor())
	assert.Equal(t, auditapplication.User(int64(installed.users.All()[1].ID())), recorded.Subject())
	assert.Equal(t, noon.Add(25*time.Minute), recorded.OccurredAt())
}

// The way back in is for an install that has a consultancy: before the setup
// is concluded, a setup code opens the first setup instead.
func TestRegisteringAnAdministratorBeforeTheSetupIsConcludedIsRefusedAndSavesNothing(t *testing.T) {
	installed := issued(t)

	err := installed.registerAdministrator.Execute(context.Background(), typedAdministrator(firstCode))

	var notConcluded domain.SetupNotConcludedError
	require.ErrorAs(t, err, &notConcluded)
	assert.Equal(t, "setup_not_concluded", notConcluded.Code())
	requireNothingSaved(t, installed)
}

// Nor does it reach the storage: the photo chosen is neither confirmed nor
// removed.
func TestRegisteringAnAdministratorBeforeTheSetupIsConcludedConfirmsNoPhoto(t *testing.T) {
	installed := issued(t)
	photo := sentImage(t, installed, firstCode, application.UserPhoto, aPNG(1024))
	tooLarge := sentImage(t, installed, firstCode, application.UserPhoto, aPNG(3*1024*1024))
	for _, chosen := range []domain.FileID{photo, tooLarge} {
		input := typedAdministrator(firstCode)
		input.Administrator.PhotoFileID = chosen

		err := installed.registerAdministrator.Execute(context.Background(), input)

		require.ErrorAs(t, err, &domain.SetupNotConcludedError{})
	}
	_, confirmed := installed.fileOf(t, photo).ConfirmedAt()
	assert.False(t, confirmed, "the photo was confirmed")
	assert.True(t, installed.holdsTheBytesOf(t, tooLarge), "the photo over the limit was removed")
}

// requireNoAdministratorRegistered checks that a registration that was
// refused left the users and the audit log as the setup left them, and the
// setup code working.
func requireNoAdministratorRegistered(t *testing.T, installed installation) {
	t.Helper()
	assert.Len(t, installed.users.All(), 1, "a user was saved")
	assert.Len(t, installed.auditLog.Entries(), 3, "the registration was recorded")
	_, err := installed.openSetup.Execute(context.Background(), application.OpenSetupInput{Code: secondCode})
	assert.NoError(t, err, "the setup code stopped working")
}

// requireEmailInUse checks that the error is the answer to an e-mail that
// belongs to a user.
func requireEmailInUse(t *testing.T, err error) {
	t.Helper()
	var inUse domain.EmailInUseError
	require.ErrorAs(t, err, &inUse)
	assert.Equal(t, "email_in_use", inUse.Code())
}

// The setup changes no user it did not create: the e-mail of one, typed in
// any case, is refused.
func TestTheEmailOfAUserThatExistsIsRefusedAndNothingIsSaved(t *testing.T) {
	installed := backIn(t)
	input := typedAdministrator(secondCode)
	input.Administrator.Email = "ANA.Souza@fabrica.example"

	err := installed.registerAdministrator.Execute(context.Background(), input)

	requireEmailInUse(t, err)
	requireNoAdministratorRegistered(t, installed)
}

func TestTheEmailOfADeactivatedUserIsRefusedToo(t *testing.T) {
	installed := backIn(t)
	existing := installed.users.All()[0]
	existing.Deactivate()
	require.NoError(t, installed.users.Save(context.Background(), &existing))
	input := typedAdministrator(secondCode)
	input.Administrator.Email = existing.Details().Email().String()

	err := installed.registerAdministrator.Execute(context.Background(), input)

	requireEmailInUse(t, err)
	requireNoAdministratorRegistered(t, installed)
	assert.True(t, installed.users.All()[0].Deactivated(), "the user that exists was changed")
}

// The way back in changes nobody else: the consultancy, the users that exist
// and the roles are as the setup left them.
func TestRegisteringAnAdministratorLeavesTheConsultancyTheUsersThatExistAndTheRolesAsTheyWere(t *testing.T) {
	installed := backIn(t)
	consultancies := installed.consultancies.All()
	users := installed.users.All()
	roles := installed.roles.All()

	err := installed.registerAdministrator.Execute(context.Background(), typedAdministrator(secondCode))

	require.NoError(t, err)
	assert.Equal(t, consultancies, installed.consultancies.All())
	assert.Equal(t, users, installed.users.All()[:len(users)])
	assert.Equal(t, roles, installed.roles.All())
}

func TestEachRequiredFieldOfTheAdministratorLeftOutIsNamedAndNothingIsSaved(t *testing.T) {
	leftOut := map[string]func(typed *domain.TypedUser){
		"administrator.name":               func(typed *domain.TypedUser) { typed.Name = "" },
		"administrator.email":              func(typed *domain.TypedUser) { typed.Email = " " },
		"administrator.country":            func(typed *domain.TypedUser) { typed.Country = "" },
		"administrator.national_id":        func(typed *domain.TypedUser) { typed.NationalID = "" },
		"administrator.preferred_language": func(typed *domain.TypedUser) { typed.PreferredLanguage = "" },
	}
	for field, leave := range leftOut {
		t.Run(field, func(t *testing.T) {
			installed := backIn(t)
			input := typedAdministrator(secondCode)
			leave(&input.Administrator)

			err := installed.registerAdministrator.Execute(context.Background(), input)

			assert.Equal(t, map[string]string{field: "required"}, brokenFields(t, err))
			requireNoAdministratorRegistered(t, installed)
		})
	}
}

func TestASetupCodeThatDoesNotWorkRegistersNobody(t *testing.T) {
	codes := map[string]func(installed installation) string{
		"a wrong code":                          func(installation) string { return "ZZZZZ-ZZZZZ-ZZZZZ-ZZZZZ" },
		"the code the setup was concluded with": func(installation) string { return firstCode },
		"a code that expired": func(installed installation) string {
			installed.clock.Advance(time.Hour)
			return secondCode
		},
	}
	for name, codeOf := range codes {
		t.Run(name, func(t *testing.T) {
			installed := backIn(t)

			err := installed.registerAdministrator.Execute(context.Background(), typedAdministrator(codeOf(installed)))

			requireSetupCodeInvalid(t, err)
			assert.Len(t, installed.users.All(), 1, "a user was saved")
			assert.Len(t, installed.auditLog.Entries(), 3, "the registration was recorded")
		})
	}
}

func TestThePhotoOfTheNewAdministratorIsConfirmedAndSavedWithThem(t *testing.T) {
	installed := backIn(t)
	photo := sentImage(t, installed, secondCode, application.UserPhoto, aPNG(1024))
	input := typedAdministrator(secondCode)
	input.Administrator.PhotoFileID = photo

	err := installed.registerAdministrator.Execute(context.Background(), input)

	require.NoError(t, err)
	saved, hasPhoto := installed.users.All()[1].Details().PhotoFileID()
	assert.True(t, hasPhoto)
	assert.Equal(t, photo, saved)
	confirmedAt, confirmed := installed.fileOf(t, photo).ConfirmedAt()
	assert.True(t, confirmed, "the photo was not confirmed")
	assert.Equal(t, noon.Add(25*time.Minute), confirmedAt)
}

func TestAPhotoThatBreaksARuleIsNamedOnItsFieldAndNothingIsSaved(t *testing.T) {
	photos := map[string]struct {
		code  string
		photo func(installed installation) domain.FileID
	}{
		"a photo never sent": {"file_not_uploaded", func(installed installation) domain.FileID {
			return startedImage(t, installed, secondCode, application.UserPhoto)
		}},
		"a file that does not exist": {"file_not_found", func(installation) domain.FileID { return 404 }},
		"a logo taken for the photo": {"file_not_found", func(installed installation) domain.FileID {
			return sentImage(t, installed, secondCode, application.ConsultancyLogo, aPNG(1024))
		}},
		"a photo over the limit": {"file_too_large", func(installed installation) domain.FileID {
			return sentImage(t, installed, secondCode, application.UserPhoto, aPNG(3*1024*1024))
		}},
	}
	for name, refused := range photos {
		t.Run(name, func(t *testing.T) {
			installed := backIn(t)
			input := typedAdministrator(secondCode)
			input.Administrator.PhotoFileID = refused.photo(installed)

			err := installed.registerAdministrator.Execute(context.Background(), input)

			assert.Equal(t, map[string]string{"administrator.photo_file_id": refused.code}, brokenFields(t, err))
			requireNoAdministratorRegistered(t, installed)
		})
	}
}

func TestAPhotoThatNamesNoFileIsNamedWithTheOtherFieldsThatBreakARule(t *testing.T) {
	installed := backIn(t)
	input := typedAdministrator(secondCode)
	input.Administrator.PhotoFileID = domain.NoSuchFile
	input.Administrator.Country = "ZZ"

	err := installed.registerAdministrator.Execute(context.Background(), input)

	assert.Equal(t, map[string]string{
		"administrator.country":       "country_unknown",
		"administrator.photo_file_id": "file_not_found",
	}, brokenFields(t, err))
	requireNoAdministratorRegistered(t, installed)
}

func TestAFailureToRecordTheRegistrationLeavesNoUserAndTheSetupCodeWorking(t *testing.T) {
	installed := backIn(t)
	auditLogIsDown := errors.New("the audit log is down")
	installed.auditLog.FailWith(auditLogIsDown)

	err := installed.registerAdministrator.Execute(context.Background(), typedAdministrator(secondCode))

	require.ErrorIs(t, err, auditLogIsDown)
	installed.auditLog.FailWith(nil)
	requireNoAdministratorRegistered(t, installed)
}
