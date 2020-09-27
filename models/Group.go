package models

import (
	"time"
)

//Group - TODO comment
type Group struct {
	ID          uint32    `gorm:"primary_key;auto_increment" json:"gid"`
	Name        *string   `gorm:"size:100;unique;not null" json:"name"`
	Description *string   `gorm:"size:255" json:"description"`
	CreatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
	Members     []*User   `gorm:"many2many:group_members"`
}

//GroupInfo - TODO comment
type GroupInfo struct {
	ID          uint32     `json:"gid"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Members     []UserInfo `json:"members"`
}

//GroupMembers - TODO comment
type GroupMembers struct {
	Members string `json:"members"`
}

// JSONGroupBody - TODO comment
type JSONGroupBody struct {
	Name           string `json:"name"`
	Description    string `json:"description"`
	Members        string `json:"members"`
	ReplaceMembers bool   `json:"replace"`
}

//GetGroupInfo - TODO comment
func GetGroupInfo(g *Group) *GroupInfo {
	var i GroupInfo
	i.ID = g.ID
	i.Name = *g.Name

	if g.Description != nil {
		i.Description = *g.Description
	}

	members := []UserInfo{}
	for _, member := range g.Members {
		members = append(members, GetUserInfo(*member))
	}
	i.Members = members

	return &i
}
