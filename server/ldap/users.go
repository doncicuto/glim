package ldap

import (
	"fmt"
	"strings"

	"github.com/doncicuto/glim/models"
	ber "github.com/go-asn1-ber/asn1-ber"
	"github.com/jinzhu/gorm"
)

func userEntry(user models.User, attributes string) map[string][]string {
	attrs := make(map[string]string)
	for _, a := range strings.Split(attributes, " ") {
		attrs[a] = a
	}

	_, operational := attrs["+"]

	values := map[string][]string{}

	_, ok := attrs["structuralObjectClass"]
	if ok || operational {
		values["structuralObjectClass"] = []string{"inetOrgPerson"}
	}

	_, ok = attrs["entryUUID"]
	if ok || operational {
		values["entryUUID"] = []string{*user.UUID}
	}

	_, ok = attrs["creatorsName"]
	if ok || operational {
		creator := *user.CreatedBy
		createdBy := ""
		if creator == "admin" {
			createdBy = fmt.Sprintf("cn=admin,%s", Domain())
		} else {
			createdBy = fmt.Sprintf("uid=%s,ou=Users,%s", creator, Domain())
		}
		values["creatorsName"] = []string{createdBy}
	}

	_, ok = attrs["createTimestamp"]
	if ok || operational {
		values["createTimestamp"] = []string{user.CreatedAt.Format("20060102150405Z")}
	}

	_, ok = attrs["creatorsName"]
	if ok || operational {
		modifier := *user.UpdatedBy
		updatedBy := ""
		if modifier == "admin" {
			updatedBy = fmt.Sprintf("cn=admin,%s", Domain())
		} else {
			updatedBy = fmt.Sprintf("uid=%s,ou=Users,%s", modifier, Domain())
		}
		values["modifiersName"] = []string{updatedBy}
	}

	_, ok = attrs["modifyTimestamp"]
	if ok || operational {
		values["modifyTimestamp"] = []string{user.UpdatedAt.Format("20060102150405Z")}
	}

	_, ok = attrs["objectClass"]
	if attributes == "ALL" || ok || operational {
		values["objectClass"] = []string{"top", "person", "inetOrgPerson", "organizationalPerson"}
	}

	if attributes == "ALL" || attrs["uid"] != "" || attrs["inetOrgPerson"] != "" || operational {
		values["uid"] = []string{*user.Username}
	}

	if attributes == "ALL" || attrs["cn"] != "" || attrs["inetOrgPerson"] != "" || operational {
		values["cn"] = []string{strings.Join([]string{*user.GivenName, *user.Surname}, " ")}
	}

	_, ok = attrs["sn"]
	if attributes == "ALL" || ok || attrs["inetOrgPerson"] != "" || operational {
		values["sn"] = []string{*user.Surname}
	}

	_, ok = attrs["givenName"]
	if attributes == "ALL" || ok || attrs["inetOrgPerson"] != "" || operational {
		values["givenName"] = []string{*user.GivenName}
	}

	_, ok = attrs["mail"]
	if attributes == "ALL" || ok || attrs["inetOrgPerson"] != "" || operational {
		values["mail"] = []string{*user.Email}
	}

	_, ok = attrs["memberOf"]
	if attributes == "ALL" || ok {
		groups := []string{}
		for _, memberOf := range user.MemberOf {
			groups = append(groups, fmt.Sprintf("cn=%s,ou=Groups,dc=example,dc=org", *memberOf.Name))
		}
		values["memberOf"] = groups
	}

	_, ok = attrs["entryDN"]
	if ok || operational {
		values["entryDN"] = []string{fmt.Sprintf("uid=%s,ou=Users,%s", *user.Username, Domain())}
	}

	_, ok = attrs["subschemaSubentry"]
	if ok || operational {
		values["subschemaSubentry"] = []string{"cn=Subschema"}
	}

	_, ok = attrs["hasSubordinates"]
	if ok || operational {
		values["hasSubordinates"] = []string{"FALSE"}
	}

	return values
}

func getManager(db *gorm.DB, id int64) (*ber.Packet, *ServerError) {
	// Find group
	u := new(models.User)
	err := db.Model(&models.User{}).Where("username = ?", "admin").First(&u).Error
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

	e := encodeSearchResultEntry(id, values, fmt.Sprintf("cn=admin,%s", Domain()))
	return e, nil
}

func getUsers(db *gorm.DB, username string, attributes string, id int64) ([]*ber.Packet, *ServerError) {

	var r []*ber.Packet
	users := []models.User{}
	if username != "" {
		if username != "admin" {
			if err := db.
				Preload("MemberOf").
				Model(&models.User{}).
				Where("username = ?", username).
				Find(&users).Error; err != nil {
				return nil, &ServerError{
					Msg:  "could not retrieve users from database",
					Code: Other,
				}
			}
		} else {
			return nil, &ServerError{
				Msg:  "could not retrieve users from database",
				Code: Other,
			}
		}
	} else {
		if err := db.
			Preload("MemberOf").
			Model(&models.User{}).
			Where("username <> ?", "admin").
			Find(&users).Error; err != nil {
			return r, &ServerError{
				Msg:  "could not retrieve users from database",
				Code: Other,
			}
		}
	}

	for _, user := range users {
		dn := fmt.Sprintf("uid=%s,ou=Users,%s", *user.Username, Domain())
		values := userEntry(user, attributes)
		e := encodeSearchResultEntry(id, values, dn)
		r = append(r, e)
	}

	return r, nil
}
