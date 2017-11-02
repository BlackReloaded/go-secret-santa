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
	"log"
	"fmt"
	"os"
	"strconv"
	"github.com/spf13/cobra"
)

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "user command to manage users",
	PersistentPreRun: func(cmd *cobra.Command, args []string){
		RootCmd.PersistentPreRun(cmd, args)
		sqlStmt := `
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER NOT NULL PRIMARY KEY, 
			firstname TEXT, 
			lastname TEXT, 
			enabled TEXT, 
			email TEXT NOT NULL UNIQUE);
		`
		_, err := db.Exec(sqlStmt)
		handleErr(err)		
	},
}

var firstname string
var lastname string

var userAddCmd = &cobra.Command{
	Use:   "add",
	Short: "add a user to the user table",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 || len(args)>1 {
			log.Fatal("Need email address");
			os.Exit(1)
		}
		tx, err := db.Begin()
		handleErr(err)		
		defer tx.Rollback()
		stmt, err := tx.Prepare("INSERT INTO users(firstname, lastname, enabled, email) VALUES (?,?,1,?)")
		handleErr(err)		
		defer stmt.Close() // danger!
		_, err = stmt.Exec(firstname, lastname, args[0])
		handleErr(err)		
		err = tx.Commit()
		handleErr(err)		
		log.Println("User created")
	},
}

var userLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "list alls users",
	Run: func(cmd *cobra.Command, args []string) {
		rows, err := db.Query("SELECT * FROM users")
		handleErr(err)

		defer rows.Close()
		fmt.Printf("%2s|%-25s|%7s|%s\n", "ID", "EMAIL", "ENABLED", "NAME")
		fmt.Printf("--|-------------------------|-------|-------------\n")
		for rows.Next() {
			var id int
			var email string
			var firstname string
			var lastname string
			var enabled bool
			err = rows.Scan(&id, &firstname, &lastname, &enabled, &email)
			handleErr(err)
			fmt.Printf("%02d|%-25s|%-7t|%s,%s\n", id, email, enabled, firstname, lastname )
		}
	},
}

func changeUser(idStr string, enable bool) {
	id, err := strconv.Atoi(idStr)
	handleErr(err)
	
	tx, err := db.Begin()
	handleErr(err)

	stmt, err := tx.Prepare("UPDATE users SET enabled=? WHERE id=?")
	handleErr(err)
	defer stmt.Close()

	_, err = stmt.Exec(enable, id)
	if err != nil {
		tx.Rollback()
		handleErr(err)		
	} else {
		tx.Commit()
	}
}

var userEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "enable user",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args)!=1 {
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
		if len(args)!=1 {
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
}