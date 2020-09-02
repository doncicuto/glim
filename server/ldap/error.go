package ldap

// ServerError - TODO comment
type ServerError struct {
	Msg  string // error
	Code int64  // LDAP result codes defined in RFC 4511
}
