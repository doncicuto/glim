package ldap

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/doncicuto/glim/models"
	ber "github.com/go-asn1-ber/asn1-ber"
	"gorm.io/gorm"
)

func groupEntry(group models.Group, attributes string, domain string) map[string][]string {
	attrs := make(map[string]string)
	for _, a := range strings.Split(attributes, " ") {
		attrs[a] = a
	}

	_, operational := attrs["+"]

	values := map[string][]string{}

	_, ok := attrs["structuralObjectClass"]
	if ok || operational {
		values["structuralObjectClass"] = []string{"groupOfNames"}
	}

	_, ok = attrs["entryUUID"]
	if ok || operational {
		values["entryUUID"] = []string{*group.UUID}
	}

	_, ok = attrs["creatorsName"]
	if ok || operational {
		creator := *group.CreatedBy
		createdBy := ""
		if creator == "admin" {
			createdBy = fmt.Sprintf("cn=admin,%s", domain)
		} else {
			createdBy = fmt.Sprintf("uid=%s,ou=Users,%s", creator, domain)
		}
		values["creatorsName"] = []string{createdBy}
	}

	_, ok = attrs["createTimestamp"]
	if ok || operational {
		values["createTimestamp"] = []string{group.CreatedAt.Format("20060102150405Z")}
	}

	_, ok = attrs["creatorsName"]
	if ok || operational {
		modifier := *group.UpdatedBy
		updatedBy := ""
		if modifier == "admin" {
			updatedBy = fmt.Sprintf("cn=admin,%s", domain)
		} else {
			updatedBy = fmt.Sprintf("uid=%s,ou=Users,%s", modifier, domain)
		}
		values["modifiersName"] = []string{updatedBy}
	}

	_, ok = attrs["modifyTimestamp"]
	if ok || operational {
		values["modifyTimestamp"] = []string{group.UpdatedAt.Format("20060102150405Z")}
	}

	if attributes == "ALL" || attrs["cn"] != "" || attrs["groupOfNames"] != "" || operational {
		values["cn"] = []string{*group.Name}
	}

	_, ok = attrs["objectClass"]
	if attributes == "ALL" || ok || operational {
		values["objectClass"] = []string{"groupOfNames"}
	}

	_, ok = attrs["entryDN"]
	if ok || operational {
		values["entryDN"] = []string{fmt.Sprintf("cn=%s,ou=Groups,%s", *group.Name, domain)}
	}

	_, ok = attrs["subschemaSubentry"]
	if ok || operational {
		values["subschemaSubentry"] = []string{"cn=Subschema"}
	}

	_, ok = attrs["hasSubordinates"]
	if ok || operational {
		values["hasSubordinates"] = []string{"FALSE"}
	}

	_, ok = attrs["member"]
	if attributes == "ALL" || ok || attrs["groupOfNames"] != "" || operational {
		members := []string{}
		for _, member := range group.Members {
			members = append(members, fmt.Sprintf("uid=%s,ou=Users,%s", *member.Username, domain))
		}
		values["member"] = members
	}

	_, ok = attrs["uid"]
	if attributes == "ALL" || ok || attrs["groupOfNames"] != "" || operational {
		uids := []string{}
		for _, member := range group.Members {
			uids = append(uids, *member.Username)
		}
		values["uid"] = uids
	}

	return values
}

func getGroups(db *gorm.DB, filter string, originalFilter string, attributes string, id int64, domain string) ([]*ber.Packet, *ServerError) {
	var r []*ber.Packet
	groups := []models.Group{}

	db = db.Preload("Members").Model(&models.Group{})
	analyzeGroupsCriteria(db, filter, false, "", 0, domain)
	err := db.Find(&groups).Error
	if err != nil {
		return nil, &ServerError{
			Msg:  "could not retrieve information from database",
			Code: Other,
		}
	}

	filterUser, _ := regexp.Compile("uid=([A-Za-z.0-9-]+)")
	if filterUser.MatchString(originalFilter) {
		matches := filterUser.FindStringSubmatch(originalFilter)
		if matches != nil {
			for _, group := range groups {
				filteredMembers := []*models.User{}
				for _, member := range group.Members {
					if *member.Username == matches[1] {
						filteredMembers = append(filteredMembers, member)
					}
				}
				group.Members = filteredMembers
				dn := fmt.Sprintf("cn=%s,ou=Groups,%s", *group.Name, domain)
				values := groupEntry(group, attributes, domain)
				e := encodeSearchResultEntry(id, values, dn)
				r = append(r, e)
			}
		}
	} else {

		for _, group := range groups {
			dn := fmt.Sprintf("cn=%s,ou=Groups,%s", *group.Name, domain)
			values := groupEntry(group, attributes, domain)
			e := encodeSearchResultEntry(id, values, dn)
			r = append(r, e)
		}
	}

	return r, nil
}

func analyzeGroupsCriteria(db *gorm.DB, filter string, boolean bool, booleanOperator string, index int, domain string) {

	if boolean {

		re := regexp.MustCompile(`\(\|(.*)\)|\(\&(.*)\)|\(\!(.*)\)|\(([a-zA-Z=\ \.*]*)\)*`)
		submatchall := re.FindAllString(filter, -1)

		for index, element := range submatchall {
			element = strings.TrimPrefix(element, "(")
			element = strings.TrimSuffix(element, ")")
			analyzeGroupsCriteria(db, element, false, booleanOperator, index, domain)
		}

	} else {
		switch {
		case strings.HasPrefix(filter, "(") && strings.HasSuffix(filter, ")"):
			element := strings.TrimPrefix(filter, "(")
			element = strings.TrimSuffix(element, ")")
			analyzeGroupsCriteria(db, element, false, "", 0, domain)
		case strings.HasPrefix(filter, "&"):
			element := strings.TrimPrefix(filter, "&")
			analyzeGroupsCriteria(db, element, true, "and", 0, domain)
		case strings.HasPrefix(filter, "|"):
			element := strings.TrimPrefix(filter, "|")
			analyzeGroupsCriteria(db, element, true, "or", 0, domain)
		case strings.HasPrefix(filter, "!"):
			element := strings.TrimPrefix(filter, "!")
			analyzeGroupsCriteria(db, element, true, "not", 0, domain)
		case strings.HasPrefix(filter, "entryDN=cn=") && strings.HasSuffix(filter, fmt.Sprintf(",ou=Groups,%s", domain)):
			element := strings.TrimPrefix(filter, "entryDN=cn=")
			element = strings.TrimSuffix(element, fmt.Sprintf(",ou=Groups,%s", domain))
			if index == 0 {
				db.Where("name = ?", element)
			} else {
				db.Or("name = ?", element)
			}
		case strings.HasPrefix(filter, "cn="):
			element := strings.TrimPrefix(filter, "cn=")
			if strings.Contains(element, "*") {
				element = strings.Replace(element, "*", "%", -1)
				if index == 0 {
					db.Where("name LIKE ?", element)
				} else {
					db.Or("name LIKE ?", element)
				}
			} else {
				element = strings.Replace(element, "*", "%", -1)
				if index == 0 {
					db.Where("name = ?", element)
				} else {
					db.Or("name = ?", element)
				}
			}
		}
	}
}
