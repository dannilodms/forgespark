// Package config reads the configuration of the binary from environment
// variables, once, into one typed struct. No other package reads the
// environment. Rules: Configuration in CODING_STANDARDS.md.
package config

import (
	"errors"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
)

// Command is a command of the binary. Each one requires its own variables.
type Command string

const (
	// ServeAPI is the command that serves the API.
	ServeAPI Command = "api"
	// Migrate is the command that applies the pending migrations.
	Migrate Command = "migrate"
)

// Config holds every value the binary takes from the environment. Each
// variable is listed, with what it is for, in .env.example at the root.
type Config struct {
	ListenAddress                    ListenAddress `env:"API_LISTEN_ADDRESS"`
	DatabaseURL                      PostgresURL   `env:"DATABASE_URL"`
	DatabasePoolMaxSize              int32         `env:"DATABASE_POOL_MAX_SIZE" envDefault:"10"`
	DatabaseStatementTimeout         time.Duration `env:"DATABASE_STATEMENT_TIMEOUT" envDefault:"30s"`
	DatabaseIdleInTransactionTimeout time.Duration `env:"DATABASE_IDLE_IN_TRANSACTION_TIMEOUT" envDefault:"1m"`
	MigrationDatabaseURL             PostgresURL   `env:"MIGRATION_DATABASE_URL"`
}

// requiredBy lists the variables with no default that each command needs. The
// API never receives the connection of the migration role, and `migrate`
// never receives the one of the API role.
var requiredBy = map[Command][]string{
	ServeAPI: {"API_LISTEN_ADDRESS", "DATABASE_URL"},
	Migrate:  {"MIGRATION_DATABASE_URL"},
}

// Load reads the environment into a Config for one command. A variable that
// is missing or not valid is an error whose text names the variable and never
// its value, so the text is safe to print.
func Load(command Command, environment map[string]string) (Config, error) {
	var loaded Config
	found := problems{}

	for _, variable := range requiredBy[command] {
		found.addWhen(environment[variable] == "", variable, "is required")
	}
	found.addUnparsed(env.ParseWithOptions(&loaded, env.Options{Environment: environment}))
	found.addWhen(loaded.DatabasePoolMaxSize < 1, "DATABASE_POOL_MAX_SIZE", "is not valid: expected at least 1")
	found.addWhen(loaded.DatabaseStatementTimeout <= 0, "DATABASE_STATEMENT_TIMEOUT", "is not valid: expected more than zero")
	found.addWhen(loaded.DatabaseIdleInTransactionTimeout <= 0, "DATABASE_IDLE_IN_TRANSACTION_TIMEOUT", "is not valid: expected more than zero")

	if len(found) > 0 {
		return Config{}, errors.New("configuration: " + found.line())
	}
	return loaded, nil
}

// problems maps a variable to what is wrong with it. The first problem found
// for a variable is the one kept.
type problems map[string]string

func (p problems) addWhen(broken bool, variable, problem string) {
	if !broken {
		return
	}
	if _, known := p[variable]; known {
		return
	}
	p[variable] = problem
}

// addUnparsed records the values the library could not read. Its own messages
// quote the value, so each one is replaced by what was expected instead.
func (p problems) addUnparsed(err error) {
	var aggregate env.AggregateError
	if !errors.As(err, &aggregate) {
		p.addWhen(err != nil, "the environment", "could not be read")
		return
	}
	for _, each := range aggregate.Errors {
		var unparsed env.ParseError
		if !errors.As(each, &unparsed) {
			p.addWhen(true, "the environment", "could not be read")
			continue
		}
		p.addWhen(true, variableOf(unparsed.Name), "is not valid: expected "+expectationOf(unparsed.Type))
	}
}

// line lists the problems in the order the variables are declared in Config.
func (p problems) line() string {
	variables := make([]string, 0, len(p))
	for variable := range p {
		variables = append(variables, variable)
	}
	declared := declaredVariables()
	slices.SortFunc(variables, func(a, b string) int {
		return slices.Index(declared, a) - slices.Index(declared, b)
	})

	parts := make([]string, 0, len(variables))
	for _, variable := range variables {
		parts = append(parts, variable+" "+p[variable])
	}
	return strings.Join(parts, "; ")
}

func declaredVariables() []string {
	fields := reflect.VisibleFields(reflect.TypeFor[Config]())
	variables := make([]string, 0, len(fields))
	for _, field := range fields {
		variables = append(variables, field.Tag.Get("env"))
	}
	return variables
}

func variableOf(fieldName string) string {
	field, found := reflect.TypeFor[Config]().FieldByName(fieldName)
	if !found {
		return "the environment"
	}
	return field.Tag.Get("env")
}

func expectationOf(valueType reflect.Type) string {
	switch valueType {
	case reflect.TypeFor[ListenAddress]():
		return "host:port"
	case reflect.TypeFor[PostgresURL]():
		return "a PostgreSQL connection URL"
	case reflect.TypeFor[time.Duration]():
		return "a duration such as 30s"
	case reflect.TypeFor[int32]():
		return "a whole number"
	}
	return "another kind of value"
}
