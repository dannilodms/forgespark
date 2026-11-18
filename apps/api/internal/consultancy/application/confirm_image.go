package application

import (
	"context"
	"errors"

	filesapplication "github.com/dannilodms/software-factory-manager/apps/api/internal/files/application"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/kernel"
)

// confirmImage confirms an image chosen in the setup, when one was chosen,
// through the files area, which reads the storage. It tells apart a rule the
// image breaks, which the person corrects, from a failure of the system.
func confirmImage(ctx context.Context, confirmUpload *filesapplication.ConfirmUpload, image chosenImage) (refused, failed error) {
	if !image.chosen {
		return nil, nil
	}
	err := confirmUpload.Execute(ctx, filesapplication.ConfirmUploadInput{
		FileID:  filesapplication.FileID(image.id),
		Purpose: image.purpose,
	})
	var broken kernel.BrokenRule
	if errors.As(err, &broken) {
		return err, nil
	}
	return nil, err
}
