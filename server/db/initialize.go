/*
Copyright © 2022 Miguel Ángel Álvarez Cabrerizo <mcabrerizo@arrakis.ovh>

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

package db

import (
	"errors"
	"fmt"

	"github.com/doncicuto/glim/models"
	"github.com/google/uuid"
	"github.com/sethvargo/go-password/password"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func createManager(db *gorm.DB) error {
	randomPass, err := password.Generate(64, 10, 0, false, true)
	if err != nil {
		return err
	}
	hash, err := models.Hash(randomPass)
	if err != nil {
		return err
	}
	userUUID := uuid.New().String()
	username := "admin"
	firstname := "LDAP"
	lastname := "administrator"
	hashed := string(hash)
	manager := true
	readonly := false

	if err := db.Create(&models.User{
		Username:  &username,
		GivenName: &firstname,
		Surname:   &lastname,
		Password:  &hashed,
		Manager:   &manager,
		Readonly:  &readonly,
		UUID:      &userUUID,
	}).Error; err != nil {
		return err
	}
	fmt.Println("")
	fmt.Println("------------------------------------- WARNING -------------------------------------")
	fmt.Println("A new user with manager permissions has been created:")
	fmt.Println("- Username: admin") // TODO - Allow username with env
	fmt.Printf("- Password %s\n", randomPass)
	fmt.Println("Please store or write down this password to manage Glim.")
	fmt.Println("You can delete this user once you assign manager permissions to another user")
	fmt.Println("-----------------------------------------------------------------------------------")

	return nil
}

//Initialize - TODO common
func Initialize() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	// If we want to log Gorm queries
	// db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})
	if err != nil {
		return nil, err
	}

	// Migrate the schema
	db.AutoMigrate(&models.User{})
	db.AutoMigrate(&models.Group{})

	// Do we have a manager? if not create one
	var manager models.User
	err = db.Where("manager = ?", true).Take(&manager).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if err := createManager(db); err != nil {
			return nil, err
		}
	}

	return db, nil
}
