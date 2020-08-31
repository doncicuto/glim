package db

import (
	"fmt"
	"os"

	"github.com/jinzhu/gorm"
	"github.com/muultipla/glim/models"
	"github.com/sethvargo/go-password/password"
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

	username := "manager"
	fullname := "Glim Manager"
	hashed := string(hash)
	manager := true
	readonly := false

	if err := db.Create(&models.User{
		Username: &username,
		Fullname: &fullname,
		Password: &hashed,
		Manager:  &manager,
		Readonly: &readonly,
	}).Error; err != nil {
		return err
	}
	fmt.Println("")
	fmt.Println("------------------------------------- WARNING -------------------------------------")
	fmt.Println("A new user with manager permissions has been created:")
	fmt.Println("- Username: manager") // TODO - Allow username with env
	fmt.Printf("- Password %s\n", randomPass)
	fmt.Println("Please store or write down this password to manage Glim.")
	fmt.Println("You can delete this user once you assign manager permissions to another user")
	fmt.Println("-----------------------------------------------------------------------------------")

	return nil
}

//Initialize - TODO common
func Initialize() (*gorm.DB, error) {
	db, err := gorm.Open(os.Getenv("DB_DRIVER"), os.Getenv("DB_NAME"))
	if err != nil {
		return nil, err
	}

	// Migrate the schema
	db.AutoMigrate(&models.User{})
	db.AutoMigrate(&models.Group{})

	// Do we have a manager? if not create one
	var manager models.User
	if db.Where("manager = ?", true).Take(&manager).RecordNotFound() {
		if err := createManager(db); err != nil {
			return nil, err
		}
	}

	return db, nil
}
