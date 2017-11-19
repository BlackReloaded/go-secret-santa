// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"log"
	"strconv"

	"github.com/jinzhu/gorm"
	"github.com/spf13/cobra"
)

type User struct {
	gorm.Model
	Firstname string
	Lastname  string
	Enabled   bool
	Email     string `gorm:"type:varchar(255);unique;not null"`
}

var firstname string
var lastname string
var email string

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "user command to manage users",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		RootCmd.PersistentPreRun(cmd, args)
		db.AutoMigrate(&User{})
	},
}

var userAddCmd = &cobra.Command{
	Use:   "add",
	Short: "add a user to the user table",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 || len(args) > 1 {
			log.Fatal("Need email address")
		}
		user := User{
			Firstname: firstname,
			Lastname:  lastname,
			Email:     args[0],
			Enabled:   true,
		}
		db.Create(&user)
		log.Println("User created")
	},
}

var userLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "list alls users",
	Run: func(cmd *cobra.Command, args []string) {
		users := []User{}
		db.Find(&users)
		fmt.Printf("%2s|%-25s|%7s|%s\n", "ID", "EMAIL", "ENABLED", "NAME")
		fmt.Printf("--|-------------------------|-------|-------------\n")
		for _, user := range users {
			fmt.Printf("%02d|%-25s|%-7t|%s,%s\n", user.ID, user.Email, user.Enabled, user.Firstname, user.Lastname)
		}
	},
}

var userUdateCmd = &cobra.Command{
	Use:   "update",
	Short: "update a user to the user table",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 || len(args) > 1 {
			log.Fatal("Need email address")
		}
		user := User{}
		db.First(&user, args[0])
		if len(firstname) > 0 {
			user.Firstname = firstname
		}
		if len(lastname) > 0 {
			user.Lastname = lastname
		}
		if len(email) > 0 {
			user.Email = email
		}
		db.Save(&user)
		log.Println("User updated")
	},
}

func changeUser(idStr string, enable bool) {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Fatal(err)
	}
	user := User{}
	db.First(&user, id)
	user.Enabled = enable
	db.Save(&user)
}

var userEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "enable user",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			log.Fatal("Need user id")
		} else {
			changeUser(args[0], true)
		}
	},
}

var userDisableCmd = &cobra.Command{
	Use:   "disable",
	Short: "disable user",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			log.Fatal("Need user id")
		} else {
			changeUser(args[0], false)
		}
	},
}

func init() {
	RootCmd.AddCommand(userCmd)
	userCmd.AddCommand(userAddCmd)
	userAddCmd.Flags().StringVar(&firstname, "firstname", "", "Firstname of the user")
	userAddCmd.Flags().StringVar(&lastname, "lastname", "", "Lastname of the user")
	userCmd.AddCommand(userLsCmd)
	userCmd.AddCommand(userEnableCmd)
	userCmd.AddCommand(userDisableCmd)
	userCmd.AddCommand(userUdateCmd)
	userUdateCmd.Flags().StringVar(&firstname, "firstname", "", "Firstname of the user")
	userUdateCmd.Flags().StringVar(&lastname, "lastname", "", "Lastname of the user")
	userUdateCmd.Flags().StringVar(&email, "email", "", "Email of the user")
}
