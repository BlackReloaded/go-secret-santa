package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	secretsanta "github.com/blackreloaded/go-secret-santa"
	"github.com/google/uuid"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var desc string
var amount float32

func init() {
	rootCmd.AddCommand(yearCmd)
	yearCmd.AddCommand(yearAddCmd)
	yearAddCmd.Flags().StringVar(&desc, "desc", "", "Description of the year")
	yearAddCmd.Flags().Float32Var(&amount, "amount", 30, "Amount of the year")
	yearCmd.AddCommand(yearLsCmd)
	yearCmd.AddCommand(yearUdateCmd)
	yearUdateCmd.Flags().StringVar(&desc, "desc", "", "Description of the year")
	yearUdateCmd.Flags().Float32Var(&amount, "amount", 0, "Amount of the year")
	yearCmd.AddCommand(yearPrintCmd)
	yearCmd.AddCommand(yearRmCmd)
	yearCmd.AddCommand(yearPairCmd)
}

var yearCmd = &cobra.Command{
	Use:   "year",
	Short: "year command to manage years",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.PersistentPreRun(cmd, args)
	},
}

var yearAddCmd = &cobra.Command{
	Use:   "add",
	Short: "add a year to the year table",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 || len(args) > 1 {
			log.Fatal("Need year address")
		}
		year := &secretsanta.Year{
			Description: desc,
			Amount:      amount,
		}
		id, _ := strconv.ParseUint(args[0], 10, 32)
		year.YearID = uint32(id)
		yid, err := secretSanta.AddYear(year)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Year created with id %s\n", yid)
	},
}

var yearRmCmd = &cobra.Command{
	Use:   "rm",
	Short: "rm a year from the year table",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 || len(args) > 1 {
			log.Fatal("Need year id")
		}
		err := secretSanta.RmYear(args[0])
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Year removed with id %s\n", args[0])
	},
}

var yearLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "list alls years",
	Run: func(cmd *cobra.Command, args []string) {
		years, err := secretSanta.ListYears()
		if err != nil {
			log.Fatal(err)
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"ID", "YEARID", "DESCRIPTION", "AMOUNT"})
		for _, v := range years {
			table.Append([]string{
				v.ID,
				fmt.Sprintf("%d", v.YearID),
				v.Description,
				fmt.Sprintf("%f", v.Amount),
			})
		}
		table.Render()
	},
}

var yearPrintCmd = &cobra.Command{
	Use:   "print",
	Short: "print a year",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			log.Fatal("need year id")
		}
		year, err := secretSanta.GetYear(args[0])
		if err != nil {
			log.Fatalf("faild to load year %v: %v", args[0], err)
		}
		err = secretSanta.PrintAll(os.Stdout, year)
		if err != nil {
			log.Fatalf("failed to print year %v: %v", args[0], err)
		}
	},
}

var yearUdateCmd = &cobra.Command{
	Use:   "update",
	Short: "update a year to the years table",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 || len(args) > 1 {
			log.Fatal("Need year")
		}
		year, err := secretSanta.GetYear(args[0])
		if err != nil {
			log.Fatal(err)
		}
		if len(desc) > 0 {
			year.Description = desc
		}
		if amount > 0 {
			year.Amount = amount
		}
		err = secretSanta.UdateYear(year)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Year updated")
	},
}

func loadId(id string) (string, error) {
	if _, err := uuid.Parse(id); err == nil {
		return id, nil
	}
	users, err := secretSanta.ListUsers(false)
	if err != nil {
		return "", errors.Wrap(err, "failed to load users")
	}
	for _, user := range users {
		if user.Email == id {
			return user.ID, nil
		}
	}
	return "", errors.New("no user found")
}

var yearPairCmd = &cobra.Command{
	Use:   "pair",
	Short: "pair two people for this year",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 3 {
			log.Fatal("Need year")
		}
		year, err := secretSanta.GetYear(args[0])
		if err != nil {
			log.Fatal(err)
		}
		byer, err := loadId(args[1])
		if err != nil {
			log.Fatalf("failed to load user %s: %v", args[1], err)
		}
		receiver, err := loadId(args[2])
		if err != nil {
			log.Fatalf("failed to load user %s: %v", args[2], err)
		}

		year.Pairing = append(year.Pairing, &secretsanta.Pairing{
			ByerUserID:     byer,
			ReceiverUserID: receiver,
			Rating:         0,
		})
		err = secretSanta.UdateYear(year)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Year updated")
	},
}
