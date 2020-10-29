package ldap

import (
	"fmt"
	"strings"

	"github.com/doncicuto/glim/models"
	ber "github.com/go-asn1-ber/asn1-ber"
	"github.com/jinzhu/gorm"
)

func getManager(db *gorm.DB, id int64) (*ber.Packet, *ServerError) {
	// Find group
	u := new(models.User)
	err := db.Model(&models.User{}).Where("username = ?", "manager").First(&u).Error
	if err != nil {
		// Does user exist?
		if !gorm.IsRecordNotFoundError(err) {
			return nil, &ServerError{
				Msg:  "could not retrieve users from database",
				Code: Other,
			}
		}
		return nil, nil
	}

	values := map[string][]string{
		"objectClass": []string{"simpleSecurityObject", "organizationalRole"},
		"cn":          []string{*u.Username},
		"description": []string{strings.Join([]string{*u.GivenName, *u.Surname}, " ")},
	}

	e := encodeSearchResultEntry(id, values, fmt.Sprintf("cn=manager,%s", Domain()))
	return e, nil
}

func getUsers(db *gorm.DB, attributes string, id int64) ([]*ber.Packet, *ServerError) {
	attrs := make(map[string]string)
	for _, a := range strings.Split(attributes, " ") {
		attrs[a] = a
	}
	fmt.Printf("%v", attrs)

	var r []*ber.Packet

	users := []models.User{}
	if err := db.
		Preload("MemberOf").
		Model(&models.User{}).
		Where("username <> ?", "manager").
		Find(&users).Error; err != nil {
		return r, &ServerError{
			Msg:  "could not retrieve users from database",
			Code: Other,
		}
	}

	for _, user := range users {
		dn := fmt.Sprintf("uid=%s,ou=Users,%s", *user.Username, Domain())
		values := map[string][]string{}

		_, ok := attrs["objectClass"]
		if ok {
			values["objectClass"] = []string{"top", "person", "inetOrgPerson", "organizationalPerson"}
		}

		_, ok = attrs["uid"]
		if ok {
			values["uid"] = []string{*user.Username}
		}

		_, ok = attrs["cn"]
		if ok {
			values["cn"] = []string{strings.Join([]string{*user.GivenName, *user.Surname}, " ")}
		}

		_, ok = attrs["sn"]
		if ok {
			values["sn"] = []string{*user.Surname}
		}

		_, ok = attrs["givenName"]
		if ok {
			values["givenName"] = []string{*user.GivenName}
		}

		_, ok = attrs["mail"]
		if ok {
			values["mail"] = []string{*user.Email}
		}

		e := encodeSearchResultEntry(id, values, dn)
		r = append(r, e)
	}

	return r, nil
}

func getUser(db *gorm.DB, username string, attributes string, id int64) (*ber.Packet, *ServerError) {
	attrs := make(map[string]string)
	for _, a := range strings.Split(attributes, " ") {
		attrs[a] = a
	}

	user := models.User{}
	if err := db.
		Preload("MemberOf").
		Model(&models.User{}).
		Where("username = ?", username).
		Take(&user).Error; err != nil {
		return nil, &ServerError{
			Msg:  "could not retrieve user from database",
			Code: Other,
		}
	}

	dn := fmt.Sprintf("uid=%s,ou=Users,%s", *user.Username, Domain())
	values := map[string][]string{}

	_, ok := attrs["objectClass"]
	if ok {
		values["objectClass"] = []string{"top", "person", "inetOrgPerson", "organizationalPerson"}
	}

	_, ok = attrs["uid"]
	if ok {
		values["uid"] = []string{*user.Username}
	}

	_, ok = attrs["cn"]
	if ok {
		values["cn"] = []string{strings.Join([]string{*user.GivenName, *user.Surname}, " ")}
	}

	_, ok = attrs["sn"]
	if ok {
		values["sn"] = []string{*user.Surname}
	}

	_, ok = attrs["givenName"]
	if ok {
		values["givenName"] = []string{*user.GivenName}
	}

	_, ok = attrs["mail"]
	if ok {
		values["mail"] = []string{*user.Email}
	}

	e := encodeSearchResultEntry(id, values, dn)

	return e, nil
}
