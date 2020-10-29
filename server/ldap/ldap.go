/*
Copyright © 2020 Miguel Ángel Álvarez Cabrerizo <mcabrerizo@arrakis.ovh>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ldap

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/doncicuto/glim/models"
	ber "github.com/go-asn1-ber/asn1-ber"
	"github.com/jinzhu/gorm"
)

// Message - TODO comment
type Message struct {
	ID      int64
	Op      int64
	Request []*ber.Packet
}

func messageID(p *ber.Packet) (int64, error) {
	if p.ClassType != ber.ClassUniversal ||
		p.TagType != ber.TypePrimitive ||
		p.Tag != ber.TagInteger ||
		len(p.Children) != 0 {
		return 0, errors.New("wrong message id definition")
	}

	id, err := ber.ParseInt64(p.ByteValue)
	if err != nil {
		return 0, errors.New("could not parse message id")
	}

	return id, nil
}

func protocolOp(p *ber.Packet) (int64, error) {
	if p.ClassType != ber.ClassApplication {
		return 0, errors.New("wrong protocol operation definition")
	}

	return int64(p.Tag), nil
}

func protocolVersion(p *ber.Packet) (int64, *ServerError) {
	if p.ClassType != ber.ClassUniversal ||
		p.TagType != ber.TypePrimitive ||
		p.Tag != ber.TagInteger ||
		len(p.Children) != 0 {
		return 0, &ServerError{
			Msg:  "wrong protocol version definition",
			Code: ProtocolError,
		}
	}

	v, err := ber.ParseInt64(p.ByteValue)
	if err != nil {
		return 0, &ServerError{
			Msg:  "could not parse protocol version",
			Code: Other,
		}
	}

	if v != Version3 {
		return 0, &ServerError{
			Msg:  "historical protocol version requested, use LDAPv3 instead",
			Code: ProtocolError,
		}
	}
	return v, nil
}

func bindName(p *ber.Packet) (string, *ServerError) {
	if p.ClassType != ber.ClassUniversal ||
		p.TagType != ber.TypePrimitive ||
		p.Tag != ber.TagOctetString ||
		len(p.Children) != 0 {
		return "", &ServerError{
			Msg:  "wrong bind name definition",
			Code: ProtocolError,
		}
	}

	n := ber.DecodeString(p.ByteValue)
	return n, nil
}

func bindPassword(p *ber.Packet) (string, *ServerError) {
	if p.ClassType != ber.ClassContext ||
		p.TagType != ber.TypePrimitive ||
		len(p.Children) != 0 {
		return "", &ServerError{
			Msg:  "wrong authentication choice definition",
			Code: ProtocolError,
		}
	}

	if p.Tag == Sasl {
		return "", &ServerError{
			Msg:  "SASL authentication not supported",
			Code: AuthMethodNotSupported,
		}
	}

	return p.Data.String(), nil
}

func requestName(p *ber.Packet) (string, *ServerError) {
	if p.ClassType != ber.ClassContext ||
		p.TagType != ber.TypePrimitive ||
		p.Tag != ber.TagEOC {
		return "", &ServerError{
			Msg:  "wrong extended request name definition",
			Code: ProtocolError,
		}
	}
	return p.Data.String(), nil
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

func decodeFilters(p *ber.Packet) (string, *ServerError) {
	filter := ""
	if p.ClassType != ber.ClassContext ||
		((p.TagType != ber.TypeConstructed && p.Tag != ber.TagEOC) &&
			(p.TagType != ber.TypePrimitive && p.Tag != ber.TagObjectDescriptor)) {
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

	case FilterPresent:
		filter = fmt.Sprintf("(%v=*)", p.Data)
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

func showWholeLDAPTree(base string, scope string) bool {
	return base == Domain() && scope == "wholeSubtree"
}

func showManagerEntry(base string, scope string) bool {
	return base == Domain() && (scope == "wholeSubtree" || scope == "singleLevel")
}

func showBaseOUEntry(base string, scope string, ou string) bool {
	return base == Domain() && (scope == "wholeSubtree" || scope == "singleLevel") ||
		(base == fmt.Sprintf("ou=%s,%s", ou, Domain()) && (scope == "wholeSubtree" || scope == "base"))
}

func showWholeUsersTree(base string, scope string) bool {
	return base == Domain() && scope == "wholeSubtree" ||
		base == fmt.Sprintf("ou=Users,%s", Domain()) && scope == "wholeSubtree"
}

func showUserInfo(base string, filter string) (string, bool) {
	regBase, _ := regexp.Compile(fmt.Sprintf("^uid=([A-Za-z]+),ou=Users,%s$", Domain()))
	if regBase.MatchString(base) {
		matches := regBase.FindStringSubmatch(base)
		if matches != nil {
			return matches[1], true
		}
	}
	regFilter, _ := regexp.Compile("^[(][&][(]objectClass=inetOrgPerson[)][(]uid=([A-Za-z]+)[)][)]$")
	if regFilter.MatchString(filter) {
		matches := regFilter.FindStringSubmatch(filter)
		if matches != nil {
			return matches[1], true
		}
	}
	return "", false
}

// HandleBind - TODO comment
func HandleBind(message *Message, db *gorm.DB, remoteAddr string) (*ber.Packet, string, error) {
	username := ""
	id := message.ID
	p := message.Request

	v, err := protocolVersion(p[0])
	if err != nil {
		return encodeBindResponse(id, err.Code, err.Msg), "", errors.New(err.Msg)
	}

	n, err := bindName(p[1])
	if err != nil {
		return encodeBindResponse(id, err.Code, err.Msg), n, errors.New(err.Msg)
	}

	dn := strings.Split(n, ",")
	if strings.HasPrefix(dn[0], "cn=") {
		username = strings.TrimPrefix(dn[0], "cn=")
		domain := strings.TrimPrefix(n, dn[0])
		domain = strings.TrimPrefix(domain, ",")
		if domain != Domain() {
			return encodeBindResponse(id, InvalidCredentials, ""), n, fmt.Errorf("wrong domain: %s", domain)
		}
	}

	if strings.HasPrefix(dn[0], "uid=") {
		username = strings.TrimPrefix(dn[0], "uid=")
		if dn[1] != "ou=Users" {
			return encodeBindResponse(id, InvalidCredentials, ""), n, fmt.Errorf("wrong ou: %s", dn[1])
		}
		domain := strings.TrimPrefix(n, dn[0])
		domain = strings.TrimPrefix(domain, ",")
		domain = strings.TrimPrefix(domain, dn[1])
		domain = strings.TrimPrefix(domain, ",")

		if domain != Domain() {
			return encodeBindResponse(id, InvalidCredentials, ""), n, fmt.Errorf("wrong domain: %s", domain)
		}
	}

	pass, err := bindPassword(p[2])
	if err != nil {
		return encodeBindResponse(id, err.Code, err.Msg), n, errors.New(err.Msg)
	}

	// DEBUG - TODO
	printLog(fmt.Sprintf("bind protocol version: %d client %s", v, remoteAddr))
	printLog(fmt.Sprintf("bind name: %s client %s", n, remoteAddr))
	printLog(fmt.Sprintf("bind password: %s client %s", "**********", remoteAddr))

	// Check credentials in database
	var dbUser models.User

	// Check if user exists
	if db.Where("username = ?", username).First(&dbUser).RecordNotFound() {
		return encodeBindResponse(id, InsufficientAccessRights, ""), n, fmt.Errorf("wrong username or password client %s", remoteAddr)
	}

	// Check if passwords match
	if err := models.VerifyPassword(*dbUser.Password, pass); err != nil {
		return encodeBindResponse(id, InvalidCredentials, ""), n, fmt.Errorf("wrong username or password client %s", remoteAddr)
	}

	// Successful bind
	printLog("success: valid credentials provided")
	r := encodeBindResponse(id, Success, "")
	return r, n, nil
}

// HandleSearchRequest - TODO comment
func HandleSearchRequest(message *Message, db *gorm.DB) ([]*ber.Packet, error) {

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
	reg, _ := regexp.Compile(fmt.Sprintf("%s$", Domain()))
	if !reg.MatchString(b) {
		p := encodeSearchResultDone(id, NoSuchObject, "")
		r = append(r, p)
		return r, errors.New("Wrong domain")
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

	// TODO parse the following items p[6] filter p[7] attributes
	f, err := searchFilter(p[6])
	if err != nil {
		p := encodeSearchResultDone(id, err.Code, err.Msg)
		r = append(r, p)
		return r, errors.New(err.Msg)
	}
	printLog(fmt.Sprintf("search filter: %s", f))

	// &{{0 32 16} <nil> []  [] }
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

	// TODO - Valid search results
	/* RFC 4511 - The results of the Search operation are returned as zero or more
	    SearchResultEntry and/or SearchResultReference messages, followed by
		a single SearchResultDone message */

	// Show only user info?
	username, onlyUser := showUserInfo(b, f)

	// Domain entry
	if !onlyUser && showWholeLDAPTree(b, scopes[s]) {
		dcs := strings.Split(Domain(), ",")
		dc := strings.TrimPrefix(dcs[0], "dc=")
		values := map[string][]string{
			"objectClass": []string{"top", "dcObject", "organization"},
			"o":           []string{Domain()},
			"dc":          []string{dc},
		}
		e := encodeSearchResultEntry(id, values, Domain())
		r = append(r, e)
	}

	// Manager entry -- Hardcoded TODO
	if !onlyUser && showManagerEntry(b, scopes[s]) {
		manager, err := getManager(db, id)
		if err != nil {
			return r, errors.New(err.Msg)
		}
		if manager != nil {
			r = append(r, manager)
		}
	}

	// ou=Users entry
	if !onlyUser && showBaseOUEntry(b, scopes[s], "Users") {
		ouUsers := fmt.Sprintf("ou=Users,%s", Domain())
		values := map[string][]string{}

		_, ok := attrs["objectClass"]
		if ok {
			values["objectClass"] = []string{"organizationalUnit", "top"}
		}

		_, ok = attrs["ou"]
		if ok {
			values["ou"] = []string{"Users"}
		}

		e := encodeSearchResultEntry(id, values, ouUsers)
		r = append(r, e)
	}

	// ou=Groups entry
	if !onlyUser && showBaseOUEntry(b, scopes[s], "Groups") {
		ouGroups := fmt.Sprintf("ou=Groups,%s", Domain())
		values := map[string][]string{
			"objectClass": []string{"organizationalUnit", "top"},
			"ou":          []string{"Groups"},
		}
		e := encodeSearchResultEntry(id, values, ouGroups)
		r = append(r, e)
	}

	// Users entries
	if onlyUser {
		user, err := getUser(db, username, a, id)
		if err != nil {
			return r, errors.New(err.Msg)
		}
		r = append(r, user)
	}

	if !onlyUser {
		if showWholeUsersTree(b, scopes[s]) {
			users, err := getUsers(db, a, id)
			if err != nil {
				return r, errors.New(err.Msg)
			}
			r = append(r, users...)
		}
	}

	d := encodeSearchResultDone(id, Success, "")
	r = append(r, d)
	return r, nil
}

// HandleExtRequest - TODO comment
func HandleExtRequest(message *Message, username string) (*ber.Packet, error) {

	id := message.ID
	p := message.Request
	n, err := requestName(p[0])
	if err != nil {
		return encodeExtendedResponse(id, err.Code, "", ""), errors.New(err.Msg)
	}

	switch n {
	case WhoamIOID:
		printLog("whoami requested by client")
		response := fmt.Sprintf("dn:%s", username)
		printLog(fmt.Sprintf("whoami response: %s", response))
		r := encodeExtendedResponse(id, Success, "", response)
		return r, nil
	default:
		printLog("unsupported extended request")
		r := encodeExtendedResponse(id, ProtocolError, "", "")
		return r, nil
	}
}

// HandleUnsupportedOperation - TODO comment
func HandleUnsupportedOperation(message *Message) (*ber.Packet, error) {
	id := message.ID
	r := encodeExtendedResponse(id, UnwillingToPerform, "1.3.6.1.4.1.1466.20036", "")
	return r, nil
}

//DecodeMessage - TODO comment
func DecodeMessage(p *ber.Packet) (*Message, error) {
	message := new(Message)

	if p.ClassType != ber.ClassUniversal ||
		p.TagType != ber.TypeConstructed ||
		p.Tag != ber.TagSequence ||
		len(p.Children) < 2 {
		return nil, errors.New("wrong ASN.1 Envelope")
	}

	id, err := messageID(p.Children[0])
	if err != nil {
		return nil, err
	}
	message.ID = id

	op, err := protocolOp(p.Children[1])
	if err != nil {
		return nil, err
	}
	message.Op = op

	message.Request = p.Children[1].Children

	return message, nil
}
