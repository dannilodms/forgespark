package rest

import (
	"time"

	"github.com/danielgtaylor/huma/v2"

	"github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/application"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/consultancy/domain"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/kernel"
	"github.com/dannilodms/software-factory-manager/apps/api/internal/platform/identifier"
)

// The bodies of the setup. Each declares in the contract what a client can
// check before it sends: a required text has at least one character, and the
// lengths, the form of an e-mail, of a telephone and of a website. The limits
// and the forms come from the domain, in TransformSchema, so the contract
// never drifts from the rules. Whether a country, a currency or a time zone
// is on its list is checked by the domain alone, and answered with the code
// of its rule.

// websitePattern is the form of a website a client checks: an http or https
// address with a host. The domain checks the rest.
const websitePattern = `^https?://[^\s/?#@]+([/?#]\S*)?$`

// SetupConsultancy is the consultancy as typed in the setup.
type SetupConsultancy struct {
	Name       string        `json:"name" minLength:"1" doc:"The name the consultancy is known by."`
	LegalName  string        `json:"legal_name" minLength:"1" doc:"The name documents carry."`
	Country    string        `json:"country" minLength:"1" doc:"The country, by its ISO 3166-1 alpha-2 code: BR."`
	TaxID      string        `json:"tax_id" minLength:"1" doc:"The identifier the tax authority of the country gives the consultancy, as the country writes it."`
	TimeZone   string        `json:"time_zone" minLength:"1" doc:"The time zone, by its name in the IANA time zone database: America/Sao_Paulo."`
	Currency   string        `json:"currency" minLength:"1" doc:"The currency, by its ISO 4217 code: BRL."`
	Email      string        `json:"email,omitempty" format:"email" doc:"The e-mail of the consultancy. It is stored in lowercase."`
	Telephone  string        `json:"telephone,omitempty" doc:"The telephone, in international form: a plus sign and 7 to 15 digits. Spaces are dropped."`
	Website    string        `json:"website,omitempty" doc:"The website: an http or https address."`
	Address    *SetupAddress `json:"address,omitempty" doc:"The address. Every part may be left out."`
	LogoFileID string        `json:"logo_file_id,omitempty" doc:"The file of the logo, as start_setup_upload answered it, once the image was sent to the storage."`
}

// TransformSchema sets the limits and the forms of the domain.
func (SetupConsultancy) TransformSchema(_ huma.Registry, schema *huma.Schema) *huma.Schema {
	limit(schema, "name", domain.NameMaxLength)
	limit(schema, "legal_name", domain.LegalNameMaxLength)
	limit(schema, "tax_id", domain.TaxIDMaxLength)
	limit(schema, "email", kernel.EmailMaxLength)
	limit(schema, "website", domain.WebsiteMaxLength)
	shape(schema, "telephone", domain.TelephonePattern)
	shape(schema, "website", websitePattern)
	shape(schema, "logo_file_id", identifier.File.Pattern())
	return schema
}

// typed gives the consultancy as typed, with the file of the logo read from
// its identifier.
func (c SetupConsultancy) typed() domain.TypedConsultancy {
	return domain.TypedConsultancy{
		Name:       c.Name,
		LegalName:  c.LegalName,
		Country:    c.Country,
		TaxID:      c.TaxID,
		TimeZone:   c.TimeZone,
		Currency:   c.Currency,
		Email:      c.Email,
		Telephone:  c.Telephone,
		Website:    c.Website,
		Address:    c.Address.typed(),
		LogoFileID: fileID(c.LogoFileID),
	}
}

// fileID reads the identifier of a file on the way in. The contract already
// refused one with another prefix. One that decodes to no number names no
// file: the use case answers it as a file that does not exist, with the other
// fields it refuses. No identifier is no file chosen.
func fileID(encoded string) domain.FileID {
	if encoded == "" {
		return 0
	}
	decoded, err := identifier.File.Decode(encoded)
	if err != nil {
		return domain.NoSuchFile
	}
	return domain.FileID(decoded)
}

// SetupAddress is an address as typed in the setup: two lines, the city, the
// region and the postal code, so that an address of any country fits.
type SetupAddress struct {
	Line1      string `json:"line1,omitempty" doc:"The first line: the street and the number."`
	Line2      string `json:"line2,omitempty" doc:"The second line: the suite, the floor, the building."`
	City       string `json:"city,omitempty" doc:"The city."`
	Region     string `json:"region,omitempty" doc:"The region: a state, a province, a county."`
	PostalCode string `json:"postal_code,omitempty" doc:"The postal code, as the country writes it."`
}

// TransformSchema sets the limits of the domain.
func (SetupAddress) TransformSchema(_ huma.Registry, schema *huma.Schema) *huma.Schema {
	limit(schema, "line1", domain.AddressLineMaxLength)
	limit(schema, "line2", domain.AddressLineMaxLength)
	limit(schema, "city", domain.CityMaxLength)
	limit(schema, "region", domain.RegionMaxLength)
	limit(schema, "postal_code", domain.PostalCodeMaxLength)
	return schema
}

// typed gives the address as typed, and no part of it when it was left out.
func (a *SetupAddress) typed() domain.TypedAddress {
	if a == nil {
		return domain.TypedAddress{}
	}
	return domain.TypedAddress{
		Line1:      a.Line1,
		Line2:      a.Line2,
		City:       a.City,
		Region:     a.Region,
		PostalCode: a.PostalCode,
	}
}

// SetupAdministrator is an administrator as typed in the setup.
type SetupAdministrator struct {
	Name              string   `json:"name" minLength:"1" doc:"The name of the person."`
	Email             string   `json:"email" minLength:"1" format:"email" doc:"The e-mail the administrator logs in with. It is stored in lowercase, and belongs to one user."`
	Country           string   `json:"country" minLength:"1" doc:"The country that gave the national ID, by its ISO 3166-1 alpha-2 code: BR."`
	NationalID        string   `json:"national_id" minLength:"1" doc:"The identifier the country gives the person, as the country writes it."`
	Telephone         string   `json:"telephone,omitempty" doc:"The telephone, in international form: a plus sign and 7 to 15 digits. Spaces are dropped."`
	PreferredLanguage language `json:"preferred_language" doc:"The supported language the administrator reads the system in: the language the setup screen was in."`
	PhotoFileID       string   `json:"photo_file_id,omitempty" doc:"The file of the photo, as start_setup_upload answered it, once the image was sent to the storage."`
}

// TransformSchema sets the limits and the forms of the domain.
func (SetupAdministrator) TransformSchema(_ huma.Registry, schema *huma.Schema) *huma.Schema {
	limit(schema, "name", domain.NameMaxLength)
	limit(schema, "email", kernel.EmailMaxLength)
	limit(schema, "national_id", domain.NationalIDMaxLength)
	shape(schema, "telephone", domain.TelephonePattern)
	shape(schema, "photo_file_id", identifier.File.Pattern())
	return schema
}

// typed gives the administrator as typed, with the file of the photo read
// from its identifier.
func (a SetupAdministrator) typed() domain.TypedUser {
	return domain.TypedUser{
		Name:              a.Name,
		Email:             a.Email,
		Country:           a.Country,
		NationalID:        a.NationalID,
		Telephone:         a.Telephone,
		PreferredLanguage: string(a.PreferredLanguage),
		PhotoFileID:       fileID(a.PhotoFileID),
	}
}

// StartSetupUploadRequest carries the setup code and what the sender
// declares of an image before sending a byte of it. No name of a file is
// sent: the storage keeps the image under a random key.
type StartSetupUploadRequest struct {
	Code      string `json:"code" minLength:"1" maxLength:"64" doc:"The setup code, as typed, as in open_setup."`
	Purpose   string `json:"purpose" doc:"What the image is for: the logo of the consultancy or the photo of its administrator."`
	MediaType string `json:"media_type" doc:"The type of the image. The browser sends it as the Content-Type of its PUT."`
	Size      int64  `json:"size" minimum:"1" doc:"The size of the image, in bytes. The browser sends exactly this many."`
}

// TransformSchema lists the purposes and the types an image may have, and
// sets the largest size, from the files area.
func (StartSetupUploadRequest) TransformSchema(_ huma.Registry, schema *huma.Schema) *huma.Schema {
	choose(schema, "purpose", []string{string(application.ConsultancyLogo), string(application.UserPhoto)})
	choose(schema, "media_type", application.ImageTypes())
	size := schema.Properties["size"]
	largest := float64(application.MaxImageSize)
	size.Maximum = &largest
	size.PrecomputeMessages()
	return schema
}

// SetupUpload is the file of an image of the setup and where to send it.
type SetupUpload struct {
	FileID    string    `json:"file_id" doc:"The identifier of the file, which concluding the setup takes as logo_file_id or as photo_file_id."`
	UploadURL string    `json:"upload_url" format:"uri" doc:"Where the browser sends the image, with a PUT from the origin of the app: the Content-Type is the declared type and the body has the declared size. The URL is signed for the host name of the file storage."`
	ExpiresAt time.Time `json:"expires_at" doc:"When the URL stops working: five minutes after it was signed."`
}

// TransformSchema sets the form of the identifier of a file.
func (SetupUpload) TransformSchema(_ huma.Registry, schema *huma.Schema) *huma.Schema {
	shape(schema, "file_id", identifier.File.Pattern())
	return schema
}

// language is a supported language. Its schema lists the supported languages
// of kernel, so the generated client knows them. It asks at least one
// character, so an empty one is answered as a required field left empty.
type language string

// Schema lists the supported languages.
func (language) Schema(huma.Registry) *huma.Schema {
	tags := kernel.SupportedLanguages()
	enum := make([]any, 0, len(tags))
	for _, tag := range tags {
		enum = append(enum, tag)
	}
	atLeastOne := 1
	return &huma.Schema{Type: huma.TypeString, MinLength: &atLeastOne, Enum: enum}
}

// limit sets the longest text a property takes, in characters.
func limit(schema *huma.Schema, property string, maxLength int) {
	field := schema.Properties[property]
	field.MaxLength = &maxLength
	field.PrecomputeMessages()
}

// shape sets the regular expression a property follows.
func shape(schema *huma.Schema, property, pattern string) {
	field := schema.Properties[property]
	field.Pattern = pattern
	field.PrecomputeMessages()
}

// choose sets the values a property may take.
func choose(schema *huma.Schema, property string, values []string) {
	field := schema.Properties[property]
	field.Enum = make([]any, 0, len(values))
	for _, value := range values {
		field.Enum = append(field.Enum, value)
	}
	field.PrecomputeMessages()
}
