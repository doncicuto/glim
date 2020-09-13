package ldap

import (
	"fmt"

	ber "github.com/go-asn1-ber/asn1-ber"
)

// Every response has the message ID
func responseHeader(messageID int64) *ber.Packet {
	r := ber.Encode(
		ber.ClassUniversal,
		ber.TypeConstructed,
		ber.TagSequence,
		nil,
		"LDAP Response")

	// Message ID
	r.AppendChild(
		ber.NewInteger(
			ber.ClassUniversal,
			ber.TypePrimitive,
			ber.TagInteger,
			messageID,
			"MessageId"))

	return r
}

func encodeResponseType(t int) *ber.Packet {
	return ber.Encode(
		ber.ClassApplication,
		ber.TypeConstructed,
		ber.Tag(t),
		nil,
		types[t])
}

func encodeResultCode(code int64) *ber.Packet {
	return ber.NewInteger(
		ber.ClassUniversal,
		ber.TypePrimitive,
		ber.TagEnumerated,
		code,
		"Result Code")
}

func encodeOctetString(value string, description string) *ber.Packet {
	return ber.NewString(
		ber.ClassUniversal,
		ber.TypePrimitive,
		ber.TagOctetString,
		value,
		description)
}

func encodeExtendedResponse(messageID int64, resultCode int64, name string, value string) *ber.Packet {
	// LDAP Message envelope
	r := responseHeader(messageID)

	// Response packet
	bp := encodeResponseType(ExtendedResponse)
	bp.AppendChild(encodeResultCode(resultCode))
	bp.AppendChild(encodeOctetString("", "MatchedDN"))
	bp.AppendChild(encodeOctetString("", "DiagnosticMessage"))
	r.AppendChild(bp)
	if name != "" {
		r.AppendChild(ber.NewString(
			ber.ClassContext,
			ber.TypePrimitive,
			ber.TagEnumerated, // responseName    [10] LDAPOID OPTIONAL 10 = TagEnumerated
			name,
			""))
	}

	if value != "" {
		r.AppendChild(ber.NewString(
			ber.ClassContext,
			ber.TypePrimitive,
			ber.TagEmbeddedPDV, // responseValue    [11] OCTET STRING OPTIONAL 11 = TagEmbeddedPDV
			value,
			""))
	}

	return r
}

func encodeBindResponse(messageID int64, resultCode int64, msg string) *ber.Packet {
	// LDAP Message envelope
	r := responseHeader(messageID)

	// Response packet
	bp := encodeResponseType(BindResponse)
	bp.AppendChild(encodeResultCode(resultCode))
	bp.AppendChild(encodeOctetString("", "MatchedDN"))
	bp.AppendChild(encodeOctetString(msg, "DiagnosticMessage"))
	r.AppendChild(bp)
	return r
}

func encodeSearchResultEntry(messageID int64, values map[string][]string) *ber.Packet {
	// LDAP Message envelope
	r := responseHeader(messageID)

	// Attributes
	a := ber.NewSequence("attributes")
	for k, v := range values {
		al := ber.NewSequence("PartialAttributeList")
		al.AppendChild(encodeOctetString(k, "PartialAttributeType"))
		vs := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSet, nil, "PartialAttributeValues")
		for _, value := range v {
			vs.AppendChild(encodeOctetString(value, "PartialAttributeValue"))
		}
		al.AppendChild(vs)
		a.AppendChild(al)
	}

	// Response packet
	bp := encodeResponseType(SearchResultEntry)
	bp.AppendChild(encodeOctetString(fmt.Sprintf("cn=admin,%s", Domain()), "objectName"))
	bp.AppendChild(a)
	r.AppendChild(bp)
	return r
}

func encodeSearchResultDone(messageID int64, resultCode int64, msg string) *ber.Packet {
	// LDAP Message envelope
	r := responseHeader(messageID)

	// Response packet
	bp := encodeResponseType(SearchResultDone)
	bp.AppendChild(encodeResultCode(resultCode))
	bp.AppendChild(encodeOctetString("", "MatchedDN"))
	bp.AppendChild(encodeOctetString(msg, "DiagnosticMessage"))

	// Add response packet to LDAP Message
	r.AppendChild(bp)
	return r
}
