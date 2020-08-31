package models

import (
	"time"
)

//Group - TODO comment
type Group struct {
	ID          uint32    `gorm:"primary_key;auto_increment" json:"id"`
	Name        *string   `gorm:"size:100;unique;not null" json:"name"`
	Description *string   `gorm:"size:255" json:"description"`
	CreatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
	Members     []*User   `gorm:"many2many:group_members"`
}

//GroupInfo - TODO comment
type GroupInfo struct {
	ID          uint32 `json:"gid"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

//GetGroupInfo - TODO comment
func GetGroupInfo(g *Group) *GroupInfo {
	var i GroupInfo
	i.ID = g.ID
	i.Name = *g.Name

	if g.Description != nil {
		i.Description = *g.Description
	}

	return &i
}
