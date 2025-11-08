package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	secretsanta "github.com/blackreloaded/go-secret-santa"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
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
	userCmd.AddCommand(userRemoveCmd)
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
		log.Printf("user with id '%d' created\n", id)
	},
}

var userLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "list alls users",
	Run: func(cmd *cobra.Command, args []string) {
		users, err := secretSanta.ListUsers(true)
		if err != nil {
			log.Fatalf("failed to load users: %v", err)
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.Header("ID", "EMAIL", "ENABLED", "FIRSTNAME", "LASTNAME")
		for _, v := range users {
			table.Append(
				v.ID,
				v.Email,
				fmt.Sprintf("%t", v.Enabled),
				v.Firstname,
				v.Lastname,
			)
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
		id, _ := strconv.ParseUint(args[0], 10, 32)
		user, err := secretSanta.GetUser(uint32(id))
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
		secretSanta.UpdateUser(user)
		log.Printf("User with id '%s' updated\n", args[0])
	},
}

var userRemoveCmd = &cobra.Command{
	Use:   "rm",
	Short: "remove a user to the user table",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 || len(args) > 1 {
			log.Fatal("Need user id")
		}
		id, _ := strconv.ParseUint(args[0], 10, 32)
		err := secretSanta.RemoveUser(uint32(id))
		if err != nil {
			log.Fatalf("failed to remove user: %v", err)
		}
		log.Printf("User with id '%s' removed\n", args[0])
	},
}

func changeUser(idStr string, enable bool) error {
	id, _ := strconv.ParseUint(idStr, 10, 32)
	user, err := secretSanta.GetUser(uint32(id))
	if err != nil {
		return fmt.Errorf("failed to load user: %s : %w", idStr, err)
	}
	user.Enabled = enable
	return secretSanta.UpdateUser(user)
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
