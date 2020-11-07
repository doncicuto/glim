package ldap

import (
	"fmt"
	"strings"

	"github.com/doncicuto/glim/models"
	ber "github.com/go-asn1-ber/asn1-ber"
	"github.com/jinzhu/gorm"
)

func groupEntry(group models.Group, attributes string) map[string][]string {
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
			createdBy = fmt.Sprintf("cn=admin,%s", Domain())
		} else {
			createdBy = fmt.Sprintf("uid=%s,ou=Users,%s", creator, Domain())
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
			updatedBy = fmt.Sprintf("cn=admin,%s", Domain())
		} else {
			updatedBy = fmt.Sprintf("uid=%s,ou=Users,%s", modifier, Domain())
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
		values["entryDN"] = []string{fmt.Sprintf("cn=%s,ou=Groups,%s", *group.Name, Domain())}
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
	if attributes == "ALL" || ok || operational {
		members := []string{}
		for _, member := range group.Members {
			members = append(members, fmt.Sprintf("uid=%s,ou=Users,%s", *member.Username, Domain()))
		}
		values["member"] = members
	}

	return values
}

func getGroups(db *gorm.DB, name string, attributes string, id int64) ([]*ber.Packet, *ServerError) {

	var r []*ber.Packet
	groups := []models.Group{}
	if name != "" {
		if err := db.
			Preload("Members").
			Model(&models.Group{}).
			Where("name = ?", name).
			Find(&groups).Error; err != nil {
			return nil, &ServerError{
				Msg:  "could not retrieve groups from database",
				Code: Other,
			}
		}
	} else {
		if err := db.
			Preload("Members").
			Model(&models.Group{}).
			Find(&groups).Error; err != nil {
			return nil, &ServerError{
				Msg:  "could not retrieve groups from database",
				Code: Other,
			}
		}
	}

	for _, group := range groups {
		dn := fmt.Sprintf("cn=%s,ou=Groups,%s", *group.Name, Domain())
		values := groupEntry(group, attributes)
		e := encodeSearchResultEntry(id, values, dn)
		r = append(r, e)
	}

	return r, nil
}

func getGroupsByUser(db *gorm.DB, username string, attributes string, id int64) ([]*ber.Packet, *ServerError) {

	var r []*ber.Packet
	groups := []models.Group{}
	if err := db.
		Preload("Members").
		Model(&models.Group{}).
		Find(&groups).Error; err != nil {
		return nil, &ServerError{
			Msg:  "could not retrieve groups from database",
			Code: Other,
		}
	}

	for _, group := range groups {
		for _, member := range group.Members {
			if *member.Username == username {
				dn := fmt.Sprintf("cn=%s,ou=Groups,%s", *group.Name, Domain())
				values := groupEntry(group, attributes)
				e := encodeSearchResultEntry(id, values, dn)
				r = append(r, e)
			}
		}
	}

	return r, nil
}
