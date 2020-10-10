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
	Members     []UserInfo `json:"members,omitempty"`
}

//GroupMembers - TODO comment
type GroupMembers struct {
	Members string `json:"members"`
}

// JSONGroupBody - TODO comment
type JSONGroupBody struct {
	Name           string `json:"name"`
	Description    string `json:"description"`
	Members        string `json:"members,omitempty"`
	ReplaceMembers bool   `json:"replace"`
}

//GetGroupInfo - TODO comment
func GetGroupInfo(g *Group, showMembers bool) *GroupInfo {
	var i GroupInfo
	i.ID = g.ID
	i.Name = *g.Name

	if g.Description != nil {
		i.Description = *g.Description
	}

	if showMembers {
		members := []UserInfo{}
		for _, member := range g.Members {
			members = append(members, GetUserInfo(*member, !showMembers))
		}
		i.Members = members
	}

	return &i
}
