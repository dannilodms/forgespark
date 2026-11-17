package application

import (
	"context"
	"errors"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/domain"
	filesapplication "github.com/dannilodms/software-factory-manager/apps/api/internal/files/application"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/kernel"
)

// confirmImage confirms an image chosen in the setup, when one was chosen,
// through the files area, which reads the storage. It tells apart a rule the
// image breaks, which the person corrects, from a failure of the system.
func confirmImage(
	ctx context.Context,
	confirmUpload *filesapplication.ConfirmUpload,
	id domain.FileID,
	chosen bool,
	purpose FilePurpose,
) (refused, failed error) {
	if !chosen {
		return nil, nil
	}
	err := confirmUpload.Execute(ctx, filesapplication.ConfirmUploadInput{FileID: filesapplication.FileID(id), Purpose: purpose})
	var broken kernel.BrokenRule
	if errors.As(err, &broken) {
		return err, nil
	}
	return nil, err
}
