package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/blackreloaded/go-secret-santa"
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

var yearLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "list alls years",
	Run: func(cmd *cobra.Command, args []string) {
		years, err := secretSanta.ListYears()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%20s|%4s|%-50s|%4s\n", "ID", "YearID", "Description", "Amount")
		fmt.Printf("-----------------|----|--------------------------------------------------|-----------\n")
		for _, year := range years {
			fmt.Printf("%20s|%04d|%-50s|%-.2f\n", year.ID, year.YearID, year.Description, year.Amount)
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
		if err!=nil {
			log.Fatal(err)
		}
		log.Println("Year updated")
	},
}
