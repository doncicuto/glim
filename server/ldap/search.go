package ldap

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	ber "github.com/go-asn1-ber/asn1-ber"
	"gorm.io/gorm"
)

func searchSize(p *ber.Packet) (int64, *ServerError) {
	if p.ClassType != ber.ClassUniversal ||
		p.TagType != ber.TypePrimitive ||
		p.Tag != ber.TagInteger {
		return 0, &ServerError{
			Msg:  "wrong search size definition",
			Code: ProtocolError,
		}
	}

	size, err := ber.ParseInt64(p.ByteValue)
	if err != nil {
		return 0, &ServerError{
			Msg:  "could not parse search size",
			Code: Other,
		}
	}

	return size, nil
}

func searchTimeLimit(p *ber.Packet) (int64, *ServerError) {
	if p.ClassType != ber.ClassUniversal ||
		p.TagType != ber.TypePrimitive ||
		p.Tag != ber.TagInteger {
		return 0, &ServerError{
			Msg:  "wrong search time limit definition",
			Code: ProtocolError,
		}
	}

	timeLimit, err := ber.ParseInt64(p.ByteValue)
	if err != nil {
		return 0, &ServerError{
			Msg:  "could not parse search time limit",
			Code: Other,
		}
	}

	return timeLimit, nil
}

func searchTypesOnly(p *ber.Packet) (bool, *ServerError) {
	if p.ClassType != ber.ClassUniversal ||
		p.TagType != ber.TypePrimitive ||
		p.Tag != ber.TagBoolean {
		return false, &ServerError{
			Msg:  "wrong types only definition",
			Code: ProtocolError,
		}
	}

	t := p.Value.(bool)
	return t, nil
}

func searchFilter(p *ber.Packet) (string, *ServerError) {
	filter, err := decodeFilters(p)
	if err != nil {
		return "", &ServerError{
			Msg:  "wrong search filter definition",
			Code: ProtocolError,
		}
	}
	return filter, nil
}

func decodeAssertionValue(p *ber.Packet) (string, *ServerError) {
	filter := ""
	if p.Tag == ber.TagOctetString {
		filter += fmt.Sprintf("%v", p.Value)
	}
	if p.Tag == ber.TagSequence && len(p.Children) == 1 {
		filter += fmt.Sprintf("=%v*", p.Children[0].Data)
	}
	return filter, nil
}

func decodeSubstringFilter(p *ber.Packet) (string, *ServerError) {
	filter := "("

	if p.ClassType != ber.ClassContext ||
		p.TagType != ber.TypeConstructed && p.Tag != ber.TagOctetString {
		return "", &ServerError{
			Msg:  "wrong search filter definition",
			Code: ProtocolError,
		}
	}

	for _, f := range p.Children {
		df, err := decodeAssertionValue(f)

		if err != nil {
			return "", &ServerError{
				Msg:  "wrong search filter definition",
				Code: ProtocolError,
			}
		}
		filter += df
	}

	filter += ")"
	return filter, nil
}

func decodeFilters(p *ber.Packet) (string, *ServerError) {
	filter := ""

	if p.ClassType != ber.ClassContext ||
		((p.TagType != ber.TypeConstructed && p.Tag != ber.TagEOC) &&
			(p.TagType != ber.TypePrimitive && p.Tag != ber.TagObjectDescriptor && p.Tag != ber.TagSequence)) {
		return "", &ServerError{
			Msg:  "wrong search filter definition",
			Code: ProtocolError,
		}
	}

	switch p.Tag {
	case FilterAnd:
		filter += "(&"
		for _, f := range p.Children {
			df, err := decodeFilters(f)
			if err != nil {
				return "", &ServerError{
					Msg:  "wrong search filter definition",
					Code: ProtocolError,
				}
			}
			filter += df
		}
		filter += ")"

	case FilterOr:
		filter += "(|"

		for _, f := range p.Children {
			df, err := decodeFilters(f)
			if err != nil {
				return "", &ServerError{
					Msg:  "wrong search filter definition",
					Code: ProtocolError,
				}
			}
			filter += df
		}
		filter += ")"

	case FilterEquality:
		if len(p.Children) != 2 {
			return "", &ServerError{
				Msg:  "wrong search filter definition",
				Code: ProtocolError,
			}
		}
		filter = fmt.Sprintf("(%v=%v)", p.Children[0].Value, p.Children[1].Value)

	case FilterSubstrings:
		df, err := decodeSubstringFilter(p)
		if err != nil {
			return "", &ServerError{
				Msg:  "wrong search filter definition",
				Code: ProtocolError,
			}
		}
		filter += df

	case FilterPresent:
		filter = fmt.Sprintf("(%v=*)", p.Data)
	default:
		printLog(fmt.Sprintf("substring %v, %v", p.Tag, p.TagType))
	}

	return filter, nil
}

func searchAttributes(p *ber.Packet) (string, *ServerError) {
	var attributes []string

	// &{{0 32 16} <nil> []  [] }
	if p.ClassType != ber.ClassUniversal ||
		p.TagType != ber.TypeConstructed ||
		p.Tag != ber.TagSequence {
		return "", &ServerError{
			Msg:  "wrong attributes definition",
			Code: ProtocolError,
		}
	}

	if len(p.Children) == 0 {
		return "ALL", nil
	}

	for _, att := range p.Children {
		attributes = append(attributes, att.Value.(string))
	}

	return strings.Join(attributes, " "), nil
}

func baseObject(p *ber.Packet) (string, *ServerError) {
	if p.ClassType != ber.ClassUniversal ||
		p.TagType != ber.TypePrimitive ||
		p.Tag != ber.TagOctetString {
		return "", &ServerError{
			Msg:  "wrong search base object definition",
			Code: ProtocolError,
		}
	}
	return p.Data.String(), nil
}

func searchScope(p *ber.Packet) (int64, *ServerError) {
	if p.ClassType != ber.ClassUniversal ||
		p.TagType != ber.TypePrimitive ||
		p.Tag != ber.TagEnumerated {
		return 0, &ServerError{
			Msg:  "wrong search scope definition",
			Code: ProtocolError,
		}
	}

	scope, err := ber.ParseInt64(p.ByteValue)
	if err != nil {
		return 0, &ServerError{
			Msg:  "could not parse search scope",
			Code: Other,
		}
	}

	if scope != BaseObject && scope != SingleLevel && scope != WholeSubtree {
		return 0, &ServerError{
			Msg:  "wrong search scope option",
			Code: ProtocolError,
		}
	}

	return scope, nil
}

// HandleSearchRequest - TODO comment
func HandleSearchRequest(message *Message, db *gorm.DB, domain string) ([]*ber.Packet, error) {

	var r []*ber.Packet
	id := message.ID
	p := message.Request
	b, err := baseObject(p[0])
	if err != nil {
		p := encodeSearchResultDone(id, err.Code, err.Msg)
		r = append(r, p)
		return r, errors.New(err.Msg)
	}
	printLog(fmt.Sprintf("search base object: %s", b))

	//Check if base object is valid
	reg, _ := regexp.Compile(fmt.Sprintf("%s$", domain))
	if !reg.MatchString(b) {
		p := encodeSearchResultDone(id, NoSuchObject, "")
		r = append(r, p)
		return r, errors.New("wrong domain")
	}

	s, err := searchScope(p[1])
	if err != nil {
		p := encodeSearchResultDone(id, err.Code, err.Msg)
		r = append(r, p)
		return r, errors.New(err.Msg)
	}
	printLog(fmt.Sprintf("search scope: %s", scopes[s]))

	// p[2] represents derefAliases which are not currently supported by Glim

	n, err := searchSize(p[3])
	if err != nil {
		p := encodeSearchResultDone(id, err.Code, err.Msg)
		r = append(r, p)
		return r, errors.New(err.Msg)
	}
	printLog(fmt.Sprintf("search maximum number of entries to be returned (0 - No limit restriction): %d", n))

	l, err := searchTimeLimit(p[4])
	if err != nil {
		p := encodeSearchResultDone(id, err.Code, err.Msg)
		r = append(r, p)
		return r, errors.New(err.Msg)
	}
	printLog(fmt.Sprintf("search maximum time limit (0 - No limit restriction): %d", l))

	t, err := searchTypesOnly(p[5])
	if err != nil {
		p := encodeSearchResultDone(id, err.Code, err.Msg)
		r = append(r, p)
		return r, errors.New(err.Msg)
	}
	printLog(fmt.Sprintf("search show types only: %t", t))

	f, err := searchFilter(p[6])
	if err != nil {
		p := encodeSearchResultDone(id, err.Code, err.Msg)
		r = append(r, p)
		return r, errors.New(err.Msg)
	}
	printLog(fmt.Sprintf("search filter: %s", f))

	a, err := searchAttributes(p[7])
	if err != nil {
		p := encodeSearchResultDone(id, err.Code, err.Msg)
		r = append(r, p)
		return r, errors.New(err.Msg)
	}
	attrs := make(map[string]string)
	for _, a := range strings.Split(a, " ") {
		attrs[a] = a
	}
	printLog(fmt.Sprintf("search attributes: %s", a))

	/* RFC 4511 - The results of the Search operation are returned as zero or more
	    SearchResultEntry and/or SearchResultReference messages, followed by
		a single SearchResultDone message */

	regBase, _ := regexp.Compile(fmt.Sprintf("^ou=users,%s$", domain))

	if (b == domain && strings.Contains(f, "objectClass=*")) || regBase.MatchString(strings.ToLower(b)) {
		if f == "(objectclass=*)" {
			ouUsers := fmt.Sprintf("ou=Users,%s", domain)
			values := map[string][]string{
				"objectClass": {"organizationalUnit", "top"},
				"ou":          {"Users"},
			}
			e := encodeSearchResultEntry(id, values, ouUsers)
			r = append(r, e)
		}

		users, err := getUsers(db, f, f, a, id, domain)
		if err != nil {
			return r, errors.New(err.Msg)
		}
		r = append(r, users...)
	}

	regBase, _ = regexp.Compile(fmt.Sprintf("^uid=([A-Za-z0-9-]+),ou=users,%s$", domain))
	if regBase.MatchString(strings.ToLower(b)) {
		matches := regBase.FindStringSubmatch(strings.ToLower(b))
		if matches != nil {

			users, err := getUsers(db, fmt.Sprintf("uid=%s", matches[1]), f, a, id, domain)
			if err != nil {
				return r, errors.New(err.Msg)
			}
			r = append(r, users...)
		}
	}

	regBase, _ = regexp.Compile(fmt.Sprintf("^ou=groups,%s$", domain))
	if (b == domain && strings.Contains(f, "objectClass=*")) || regBase.MatchString(strings.ToLower(b)) {
		if f == "(objectclass=*)" {
			ouGroups := fmt.Sprintf("ou=Groups,%s", domain)
			values := map[string][]string{
				"objectClass": {"organizationalUnit", "top"},
				"ou":          {"Groups"},
			}
			e := encodeSearchResultEntry(id, values, ouGroups)
			r = append(r, e)
		}
		groups, err := getGroups(db, f, a, id, domain)
		if err != nil {
			return r, errors.New(err.Msg)
		}
		r = append(r, groups...)
	}

	// if query.showUsers || query.filterUser != "" || query.filterUsersByGroup != "" {
	// 	// users, err := getUsers(db, query.filterUser, query.filterUsersByGroup, a, id)
	// 	users, err := getUsers(db, f, a, id)
	// 	if err != nil {
	// 		return r, errors.New(err.Msg)
	// 	}
	// 	r = append(r, users...)
	// }

	// Groups entries

	// if query.filterGroupsByUser != "" {
	// 	groups, err := getGroupsByUser(db, query.filterGroupsByUser, a, id)
	// 	if err != nil {
	// 		return r, errors.New(err.Msg)
	// 	}
	// 	r = append(r, groups...)
	// }

	// if query.showGroups || query.filterGroup != "" || query.showGroupOfNames {
	// 	groups, err := getGroupsPro(db, f, a, id)
	// 	if err != nil {
	// 		return r, errors.New(err.Msg)
	// 	}
	// 	r = append(r, groups...)
	// }

	// if query.filterGroup != "" {
	// 	groups, err := getGroups(db, query.filterGroup, a, id)
	// 	if err != nil {
	// 		return r, errors.New(err.Msg)
	// 	}
	// 	r = append(r, groups...)
	// }

	// if (query.showGroups && query.filterGroupsByUser == "") || query.showGroupOfNames {
	// 	groups, err := getGroups(db, "", a, id)
	// 	if err != nil {
	// 		return r, errors.New(err.Msg)
	// 	}
	// 	r = append(r, groups...)
	// }

	d := encodeSearchResultDone(id, Success, "")
	r = append(r, d)
	return r, nil
}
