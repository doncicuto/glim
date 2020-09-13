package ldap

import (
	"errors"
	"fmt"
	"strings"

	ber "github.com/go-asn1-ber/asn1-ber"
	"github.com/jinzhu/gorm"
	"github.com/muultipla/glim/models"
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

// HandleBind - TODO comment
func HandleBind(message *Message, db *gorm.DB, remoteAddr string) (*ber.Packet, error) {

	id := message.ID
	p := message.Request

	v, err := protocolVersion(p[0])
	if err != nil {
		return encodeBindResponse(id, err.Code, err.Msg), errors.New(err.Msg)
	}

	n, err := bindName(p[1])
	if err != nil {
		return encodeBindResponse(id, err.Code, err.Msg), errors.New(err.Msg)
	}

	rdn := strings.Split(n, ",")
	n = strings.TrimPrefix(rdn[0], "cn=")

	pass, err := bindPassword(p[2])
	if err != nil {
		return encodeBindResponse(id, err.Code, err.Msg), errors.New(err.Msg)
	}

	// DEBUG - TODO
	printLog(fmt.Sprintf("bind protocol version: %d client %s", v, remoteAddr))
	printLog(fmt.Sprintf("bind name: %s client %s", n, remoteAddr))
	printLog(fmt.Sprintf("bind password: %s client %s", "**********", remoteAddr))

	// Check credentials in database
	var dbUser models.User

	// Check if user exists
	if db.Where("username = ?", n).First(&dbUser).RecordNotFound() {
		return encodeBindResponse(id, InsufficientAccessRights, ""), fmt.Errorf("wrong username or password client %s", remoteAddr)
	}

	// Check if passwords match
	if err := models.VerifyPassword(*dbUser.Password, pass); err != nil {
		return encodeBindResponse(id, InsufficientAccessRights, ""), fmt.Errorf("wrong username or password client %s", remoteAddr)
	}

	// Successful bind
	printLog("success: valid credentials provided")
	r := encodeBindResponse(id, Success, "")
	return r, nil
}

// HandleSearchRequest - TODO comment
func HandleSearchRequest(message *Message) ([]*ber.Packet, error) {

	var r []*ber.Packet
	id := message.ID
	p := message.Request

	b, err := baseObject(p[0])
	if err != nil {
		p := encodeSearchResultDone(id, err.Code, err.Msg)
		r = append(r, p)
		return r, errors.New(err.Msg)
	}

	s, err := searchScope(p[1])
	if err != nil {
		p := encodeSearchResultDone(id, err.Code, err.Msg)
		r = append(r, p)
		return r, errors.New(err.Msg)
	}

	// p[2] represents derefAliases which are not currently supported by Glim

	n, err := searchSize(p[3])
	if err != nil {
		p := encodeSearchResultDone(id, err.Code, err.Msg)
		r = append(r, p)
		return r, errors.New(err.Msg)
	}

	l, err := searchTimeLimit(p[4])
	if err != nil {
		p := encodeSearchResultDone(id, err.Code, err.Msg)
		r = append(r, p)
		return r, errors.New(err.Msg)
	}

	t, err := searchTypesOnly(p[5])
	if err != nil {
		p := encodeSearchResultDone(id, err.Code, err.Msg)
		r = append(r, p)
		return r, errors.New(err.Msg)
	}

	// TODO parse the following items p[6] filter p[7] attributes
	// &{{128 0 7} <nil> [] objectclass [] } filter: (objectclass=*)
	// &{{0 32 16} <nil> []  [] }

	// DEBUG - TODO
	printLog(fmt.Sprintf("search base object: %s", b))
	printLog(fmt.Sprintf("search scope: %s", scopes[s]))
	printLog(fmt.Sprintf("search maximum number of entries to be returned (0 - No limit restriction): %d", n))
	printLog(fmt.Sprintf("search maximum time limit (0 - No limit restriction): %d", l))
	printLog(fmt.Sprintf("search show types only: %t", t))

	// TODO - Valid search results
	/* RFC 4511 - The results of the Search operation are returned as zero or more
	    SearchResultEntry and/or SearchResultReference messages, followed by
		a single SearchResultDone message */

	values := map[string][]string{
		"objectClass": []string{"top", "dcObject", "organization"},
		"o":           []string{Domain()},
		"dc":          []string{"example"}, //TODO hardcoded top o
	}

	e := encodeSearchResultEntry(id, values)
	r = append(r, e)

	values = map[string][]string{
		"objectClass": []string{"simpleSecurityObject", "organizationalRole"},
		"cn":          []string{"admin"},
		"description": []string{"LDAP administrator"},
	}

	e = encodeSearchResultEntry(id, values)
	r = append(r, e)

	d := encodeSearchResultDone(id, Success, "")
	r = append(r, d)
	return r, nil
}

// HandleExtRequest - TODO comment
func HandleExtRequest(message *Message) (*ber.Packet, error) {

	id := message.ID
	p := message.Request

	n, err := requestName(p[0])
	if err != nil {
		return encodeExtendedResponse(id, err.Code, "", ""), errors.New(err.Msg)
	}

	switch n {
	case WhoamIOID: // TODO - WHOAMI
		printLog("whoami requested by client")
		printLog(fmt.Sprintf("whoami response: dn:cn=admin,%s", Domain()))
		r := encodeExtendedResponse(id, Success, "", fmt.Sprintf("dn:cn=admin,%s", Domain()))
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
