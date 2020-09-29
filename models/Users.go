package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

//User - TODO comment
type User struct {
	ID        uint32    `gorm:"primary_key;auto_increment" json:"uid"`
	Username  *string   `gorm:"size:64;not null;unique" json:"username"`
	Fullname  *string   `gorm:"size:300;not null" json:"fullname"`
	Email     *string   `gorm:"size:322" json:"email"`
	Password  *string   `gorm:"size:60;not null;" json:"password"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
	Manager   *bool     `gorm:"default:false" json:"manager"`
	Readonly  *bool     `gorm:"default:true" json:"readonly"`
	MemberOf  []*Group  `gorm:"many2many:group_members"`
}

//UserInfo - TODO comment
type UserInfo struct {
	ID       uint32      `json:"uid"`
	Username string      `json:"username"`
	Fullname string      `json:"fullname"`
	Email    string      `json:"email"`
	Manager  bool        `json:"manager"`
	Readonly bool        `json:"readonly"`
	MemberOf []GroupInfo `json:"memberOf,omitempty"`
}

//Hash - TODO comment
func Hash(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

//VerifyPassword - TODO comment
func VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

//GetUserInfo - TODO comment
func GetUserInfo(u User, showMemberOf bool) UserInfo {
	var i UserInfo
	i.ID = u.ID
	if u.Username != nil {
		i.Username = *u.Username
	}
	if u.Fullname != nil {
		i.Fullname = *u.Fullname
	}
	if u.Email != nil {
		i.Email = *u.Email
	}
	if u.Manager != nil {
		i.Manager = *u.Manager
	}
	if u.Readonly != nil {
		i.Readonly = *u.Readonly
	}

	if showMemberOf {
		members := []GroupInfo{}
		for _, member := range u.MemberOf {
			members = append(members, *GetGroupInfo(member, !showMemberOf))
		}
		i.MemberOf = members
	}

	return i
}
