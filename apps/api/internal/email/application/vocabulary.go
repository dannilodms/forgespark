package application

import "github.com/dannilodms/software-factory-manager/apps/api/internal/email/domain"

// Another domain area imports this package and no other of the e-mail area,
// so the words it needs to queue an e-mail and to give its content are named
// here.
type (
	// EmailID identifies an e-mail record.
	EmailID = domain.EmailID
	// Template is the kind of an e-mail.
	Template = domain.Template
	// Reference is what an e-mail refers to.
	Reference = domain.Reference
	// Content is what one e-mail says beyond the texts of its template.
	Content = domain.Content
	// LoginCodeContent is the content of the e-mail of a login code.
	LoginCodeContent = domain.LoginCodeContent
	// RecipientInvalidError is a recipient that is not an e-mail address.
	RecipientInvalidError = domain.RecipientInvalidError
)

// The templates of the e-mails the system sends.
const (
	LoginCode = domain.LoginCode
)

// LoginCodeRequest is the login code request with this identifier, as what
// an e-mail refers to.
func LoginCodeRequest(identifier int64) Reference {
	return domain.LoginCodeRequest(identifier)
}
