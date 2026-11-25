package main

import (
	"github.com/dannilodms/forgespark/apps/api/internal/consultancy/domain"
	"github.com/dannilodms/forgespark/apps/api/internal/platform/config"
)

// The consultancy and the administrator the seed registers, as a person would
// type them in the setup. Nothing here is real (ADR-0021): the people exist
// nowhere, the domain of the default mailbox is reserved for examples, and
// the tax ID and the national ID are placeholders in the form of Brazil that
// no check digit holds for. Both e-mail addresses come from the mailbox of
// the configuration, so whoever owns it receives every e-mail of the seed.

// seededConsultancy is the fictitious software consultancy. Its e-mail is
// the mailbox itself. It has no logo, no telephone and no address.
func seededConsultancy(mailbox config.Mailbox) domain.TypedConsultancy {
	return domain.TypedConsultancy{
		Name:       "Fábrica Exemplo",
		LegalName:  "Fábrica Exemplo Software Ltda.",
		Country:    "BR",
		TaxID:      "00.000.000/0000-00",
		TimeZone:   "America/Sao_Paulo",
		Currency:   "BRL",
		Email:      mailbox.String(),
		Telephone:  "",
		Website:    "https://fabrica.example",
		Address:    domain.TypedAddress{Line1: "", Line2: "", City: "", Region: "", PostalCode: ""},
		LogoFileID: 0,
	}
}

// seededAdministrator is the first administrator, at the address of the
// mailbox tagged `admin`, which the local login provider lists. They read
// the system in Brazilian Portuguese, so the shell shows the preferred
// language of a user at work where the browser asks for another. They have
// no photo and no telephone.
func seededAdministrator(mailbox config.Mailbox) domain.TypedUser {
	return domain.TypedUser{
		Name:              "Ana Exemplo",
		Email:             mailbox.Subaddress("admin"),
		Country:           "BR",
		NationalID:        "000.000.000-00",
		Telephone:         "",
		PreferredLanguage: "pt-BR",
		PhotoFileID:       0,
	}
}
