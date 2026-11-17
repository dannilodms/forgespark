package databasetest

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io/fs"

	"github.com/peterldowns/pgtestdb"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/database"
	"github.com/dannilodms/software-factory-manager/apps/api/migrations"
)

// migrator builds the template of pgtestdb with the function the `migrate`
// command calls, on the migrations the binary carries. pgtestdb then knows
// nothing of goose, so a new version of goose cannot break the two apart.
type migrator struct{}

// Hash names the template. It changes when a migration is added or edited,
// and pgtestdb builds a new template then.
func (migrator) Hash() (string, error) {
	files, err := fs.Glob(migrations.Files, "*.sql")
	if err != nil {
		return "", fmt.Errorf("list the migrations: %w", err)
	}
	digest := sha256.New()
	var unread error
	for _, file := range files {
		unread = errors.Join(unread, addMigration(digest, file))
	}
	if unread != nil {
		return "", unread
	}
	return hex.EncodeToString(digest.Sum(nil)), nil
}

func addMigration(digest hash.Hash, file string) error {
	content, err := fs.ReadFile(migrations.Files, file)
	if err != nil {
		return fmt.Errorf("read the migration %s: %w", file, err)
	}
	fmt.Fprintf(digest, "%s %d\n", file, len(content))
	digest.Write(content)
	return nil
}

// Migrate applies every migration to the new template, as the migration role.
func (migrator) Migrate(ctx context.Context, _ *sql.DB, template pgtestdb.Config) error {
	_, err := database.Migrate(ctx, template.URL(), migrations.Files)
	return err
}
