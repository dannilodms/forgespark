package domain

// currencies are the codes of ISO 4217 of the currencies in circulation, in
// alphabetical order: 155 codes.
//
// Source: data/iso_4217.json of Debian's iso-codes, release v4.20.1
// (https://salsa.debian.org/iso-codes-team/iso-codes), every alpha_3 code
// except the ones that are not a currency a consultancy charges in: the funds
// and units of account of one country (BOV, CHE, CHW, CLF, COU, MXV, USN, UYI,
// UYW), and the precious metals, the bond market units, the units of account
// of institutions, the code for testing and the code for no currency (XAD,
// XAG, XAU, XBA, XBB, XBC, XBD, XDR, XPD, XPT, XSU, XTS, XUA, XXX).
// To update, take that file of a newer release and leave the same kinds out,
// then run `task api:lists`: the app offers this list.
var currencies = []string{
	"AED", "AFN", "ALL", "AMD", "AOA", "ARS", "AUD", "AWG", "AZN", "BAM",
	"BBD", "BDT", "BHD", "BIF", "BMD", "BND", "BOB", "BRL", "BSD", "BTN",
	"BWP", "BYN", "BZD", "CAD", "CDF", "CHF", "CLP", "CNY", "COP", "CRC",
	"CUP", "CVE", "CZK", "DJF", "DKK", "DOP", "DZD", "EGP", "ERN", "ETB",
	"EUR", "FJD", "FKP", "GBP", "GEL", "GHS", "GIP", "GMD", "GNF", "GTQ",
	"GYD", "HKD", "HNL", "HTG", "HUF", "IDR", "ILS", "INR", "IQD", "IRR",
	"ISK", "JMD", "JOD", "JPY", "KES", "KGS", "KHR", "KMF", "KPW", "KRW",
	"KWD", "KYD", "KZT", "LAK", "LBP", "LKR", "LRD", "LSL", "LYD", "MAD",
	"MDL", "MGA", "MKD", "MMK", "MNT", "MOP", "MRU", "MUR", "MVR", "MWK",
	"MXN", "MYR", "MZN", "NAD", "NGN", "NIO", "NOK", "NPR", "NZD", "OMR",
	"PAB", "PEN", "PGK", "PHP", "PKR", "PLN", "PYG", "QAR", "RON", "RSD",
	"RUB", "RWF", "SAR", "SBD", "SCR", "SDG", "SEK", "SGD", "SHP", "SLE",
	"SOS", "SRD", "SSP", "STN", "SVC", "SYP", "SZL", "THB", "TJS", "TMT",
	"TND", "TOP", "TRY", "TTD", "TWD", "TZS", "UAH", "UGX", "USD", "UYU",
	"UZS", "VED", "VES", "VND", "VUV", "WST", "XAF", "XCD", "XCG", "XOF",
	"XPF", "YER", "ZAR", "ZMW", "ZWG",
}
