package ldap

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/doncicuto/glim/models"
	ber "github.com/go-asn1-ber/asn1-ber"
	"gorm.io/gorm"
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
		values["objectClass"] = []string{"top", "person", "inetOrgPerson", "organizationalPerson", "ldapPublicKey"}
	}

	if attributes == "ALL" || attrs["uid"] != "" || attrs["inetOrgPerson"] != "" || operational {
		if user.Username != nil {
			values["uid"] = []string{*user.Username}
		}
	}

	if attributes == "ALL" || attrs["cn"] != "" || attrs["inetOrgPerson"] != "" || operational {
		if user.GivenName != nil && user.Surname != nil {
			values["cn"] = []string{strings.Join([]string{*user.GivenName, *user.Surname}, " ")}
		}
	}

	_, ok = attrs["sn"]
	if attributes == "ALL" || ok || attrs["inetOrgPerson"] != "" || operational {
		if user.Surname != nil {
			values["sn"] = []string{*user.Surname}
		}
	}

	_, ok = attrs["givenName"]
	if attributes == "ALL" || ok || attrs["inetOrgPerson"] != "" || operational {
		if user.GivenName != nil {
			values["givenName"] = []string{*user.GivenName}
		}
	}

	_, ok = attrs["mail"]
	if attributes == "ALL" || ok || attrs["inetOrgPerson"] != "" || operational {
		if user.Email != nil {
			values["mail"] = []string{*user.Email}
		}
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
		if user.Username != nil {
			values["entryDN"] = []string{fmt.Sprintf("uid=%s,ou=Users,%s", *user.Username, Domain())}
		}
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

// func getManager(db *gorm.DB, id int64) (*ber.Packet, *ServerError) {
// 	// Find group
// 	u := new(models.User)
// 	err := db.Model(&models.User{}).Where("username = ?", "admin").First(&u).Error
// 	if err != nil {
// 		// Does user exist?

// 		if !errors.Is(err, gorm.ErrRecordNotFound) {
// 			return nil, &ServerError{
// 				Msg:  "could not retrieve users from database",
// 				Code: Other,
// 			}
// 		}
// 		return nil, nil
// 	}

// 	values := map[string][]string{
// 		"objectClass": {"simpleSecurityObject", "organizationalRole"},
// 		"cn":          {*u.Username},
// 		"description": {strings.Join([]string{*u.GivenName, *u.Surname}, " ")},
// 	}

// 	e := encodeSearchResultEntry(id, values, fmt.Sprintf("cn=admin,%s", Domain()))
// 	return e, nil
// }

// func getUsers(db *gorm.DB, username string, groupName string, attributes string, id int64) ([]*ber.Packet, *ServerError) {

// 	var r []*ber.Packet
// 	users := []models.User{}

// 	if username != "" {
// 		if username != "admin" {
// 			if err := db.
// 				Preload("MemberOf").
// 				Model(&models.User{}).
// 				Where("username = ? ", username).
// 				Find(&users).Error; err != nil {
// 				return nil, &ServerError{
// 					Msg:  "could not retrieve users from database",
// 					Code: Other,
// 				}
// 			}
// 		} else {
// 			return nil, &ServerError{
// 				Msg:  "could not retrieve users from database",
// 				Code: Other,
// 			}
// 		}
// 	} else {
// 		if err := db.
// 			Preload("MemberOf").
// 			Model(&models.User{}).
// 			Find(&users).Error; err != nil {
// 			return nil, &ServerError{
// 				Msg:  "could not retrieve users from database",
// 				Code: Other,
// 			}
// 		}
// 	}

// 	if groupName != "" {
// 		for _, user := range users {
// 			for _, member := range user.MemberOf {
// 				if *member.Name == groupName {
// 					dn := fmt.Sprintf("uid=%s,ou=Users,%s", *user.Username, Domain())
// 					values := userEntry(user, attributes)
// 					e := encodeSearchResultEntry(id, values, dn)
// 					r = append(r, e)
// 				}
// 			}
// 		}
// 	} else {
// 		for _, user := range users {
// 			if *user.Username != "admin" {
// 				dn := fmt.Sprintf("uid=%s,ou=Users,%s", *user.Username, Domain())
// 				values := userEntry(user, attributes)
// 				e := encodeSearchResultEntry(id, values, dn)
// 				r = append(r, e)
// 			}
// 		}
// 	}

// 	return r, nil
// }

func getUsers(db *gorm.DB, filter string, originalFilter string, attributes string, id int64) ([]*ber.Packet, *ServerError) {
	var r []*ber.Packet
	users := []models.User{}

	db = db.Preload("MemberOf").Model(&models.User{})
	analyzeUsersCriteria(db, filter, false, "", 0)
	err := db.Find(&users).Error
	if err != nil {
		return nil, &ServerError{
			Msg:  "could not retrieve information from database",
			Code: Other,
		}
	}

	filterGroup, _ := regexp.Compile(fmt.Sprintf("memberOf=cn=([A-Za-z0-9-]+),ou=Groups,%s", Domain()))

	for _, user := range users {
		if *user.Username != "admin" && !*user.Readonly {
			if filterGroup.MatchString(originalFilter) {
				matches := filterGroup.FindStringSubmatch(originalFilter)
				if matches != nil {
					for _, group := range user.MemberOf {
						if *group.Name == matches[1] {
							dn := fmt.Sprintf("uid=%s,ou=Users,%s", *user.Username, Domain())
							values := userEntry(user, attributes)
							e := encodeSearchResultEntry(id, values, dn)
							r = append(r, e)
							break
						}
					}
				}

			} else {
				dn := fmt.Sprintf("uid=%s,ou=Users,%s", *user.Username, Domain())
				values := userEntry(user, attributes)
				e := encodeSearchResultEntry(id, values, dn)
				r = append(r, e)
			}

		}

	}

	return r, nil
}

func analyzeUsersCriteria(db *gorm.DB, filter string, boolean bool, booleanOperator string, index int) {
	if boolean {
		re := regexp.MustCompile(`\(\|(.*)\)|\(\&(.*)\)|\(\!(.*)\)|\(([a-zA-Z=*]*)\)`)
		submatchall := re.FindAllString(filter, -1)

		for index, element := range submatchall {
			element = strings.TrimPrefix(element, "(")
			element = strings.TrimSuffix(element, ")")
			analyzeUsersCriteria(db, element, false, booleanOperator, index)
		}

	} else {
		switch {
		case strings.HasPrefix(filter, "(") && strings.HasSuffix(filter, ")"):
			element := strings.TrimPrefix(filter, "(")
			element = strings.TrimSuffix(element, ")")
			analyzeUsersCriteria(db, element, false, "", 0)
		case strings.HasPrefix(filter, "&"):
			element := strings.TrimPrefix(filter, "&")
			analyzeUsersCriteria(db, element, true, "and", 0)
		case strings.HasPrefix(filter, "|"):
			element := strings.TrimPrefix(filter, "|")
			analyzeUsersCriteria(db, element, true, "or", 0)
		case strings.HasPrefix(filter, "!"):
			element := strings.TrimPrefix(filter, "!")
			analyzeUsersCriteria(db, element, true, "not", 0)
		case strings.HasPrefix(filter, "uid="):
			element := strings.TrimPrefix(filter, "uid=")
			if strings.Contains(element, "*") {
				element = strings.Replace(element, "*", "%", -1)
				if index == 0 {
					db.Where("username LIKE ?", element)
				} else {
					db.Or("username LIKE ?", element)
				}
			} else {
				element = strings.Replace(element, "*", "%", -1)
				if index == 0 {
					db.Where("username = ?", element)
				} else {
					db.Or("username = ?", element)
				}
			}

		case strings.HasPrefix(filter, "sn="):
			element := strings.TrimPrefix(filter, "sn=")
			if strings.Contains(element, "*") {
				element = strings.Replace(element, "*", "%", -1)
				if index == 0 {
					db.Where("surname LIKE ?", element)
				} else {
					db.Or("surname LIKE ?", element)
				}
			} else {
				element = strings.Replace(element, "*", "%", -1)
				if index == 0 {
					db.Where("surname = ?", element)
				} else {
					db.Or("surname = ?", element)
				}
			}
		case strings.HasPrefix(filter, "givenName="):
			element := strings.TrimPrefix(filter, "givenName=")
			if strings.Contains(element, "*") {
				element = strings.Replace(element, "*", "%", -1)
				if index == 0 {
					db.Where("given_name LIKE ?", element)
				} else {
					db.Or("given_name LIKE ?", element)
				}
			} else {
				element = strings.Replace(element, "*", "%", -1)
				if index == 0 {
					db.Where("given_name = ?", element)
				} else {
					db.Or("given_name = ?", element)
				}
			}

		}
	}
}
