package main

import (
	"fmt"
	"log"
	"os"

	"github.com/blackreloaded/go-secret-santa"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/pkg/errors"
)

var firstname string
var lastname string
var email string

func init() {
	rootCmd.AddCommand(userCmd)
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

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "user command to manage users",
	Args:  cobra.ExactArgs(1),
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.PersistentPreRun(cmd, args)
	},
}

var userAddCmd = &cobra.Command{
	Use:   "add",
	Short: "add a user to the user table",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 || len(args) > 1 {
			log.Fatal("Need email address")
		}
		user := &secretsanta.User{
			Firstname: firstname,
			Lastname:  lastname,
			Email:     args[0],
			Enabled:   true,
		}
		id, err := secretSanta.AddUser(user)
		if err != nil {
			log.Fatalf("failed to add user: %v", err)
		}
		log.Printf("user with id '%s' created\n", id)
	},
}

var userLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "list alls users",
	Run: func(cmd *cobra.Command, args []string) {
		users, err := secretSanta.ListUsers()
		if err != nil {
			log.Fatalf("failed to load users: %v", err)
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"ID", "EMAIL", "ENABLED", "FIRSTNAME", "LASTNAME"})
		for _, v := range users {
			table.Append([]string{
				v.ID,
				v.Email,
				fmt.Sprintf("%t", v.Enabled),
				v.Firstname,
				v.Lastname,
			})
		}
		table.Render()
	},
}

var userUdateCmd = &cobra.Command{
	Use:   "update",
	Short: "update a user to the user table",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 || len(args) > 1 {
			log.Fatal("Need user id")
		}
		user, err := secretSanta.GetUser(args[0])
		if err != nil {
			log.Fatalf("failed to load user: %v", err)
		}
		if firstname != "" {
			user.Firstname = firstname
		}
		if lastname != "" {
			user.Lastname = lastname
		}
		if email != "" {
			user.Email = email
		}
		secretSanta.UdateUser(user)
		log.Printf("User with id '%s' updated\n", args[0])
	},
}

func changeUser(idStr string, enable bool) error {
	user, err := secretSanta.GetUser(idStr)
	if err!=nil {
		return errors.Wrapf(err, "failed to load user: %s", idStr)
	}
	user.Enabled = enable
	return secretSanta.UdateUser(user)
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
