package application

import "context"

// TellSetupState tells anyone, with no session, whether the setup is
// concluded and whether a setup code is waiting. Someone outside learns that
// much and nothing else (ADR-0024).
type TellSetupState struct {
	reader SetupStateReader
	clock  Clock
}

// NewTellSetupState builds the use case.
func NewTellSetupState(reader SetupStateReader, clock Clock) *TellSetupState {
	return &TellSetupState{reader: reader, clock: clock}
}

// SetupState is the state of the setup. SetupCodeWaiting is true while a
// setup code is issued and still works, before the setup is concluded and
// after it.
type SetupState struct {
	Concluded        bool
	SetupCodeWaiting bool
}

// Execute reads the state of the setup at this moment.
func (u *TellSetupState) Execute(ctx context.Context) (SetupState, error) {
	return u.reader.Read(ctx, u.clock.Now())
}
