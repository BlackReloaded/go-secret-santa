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

type Year struct {
	gorm.Model
	Description string
	Amount      float32 `gorm:"not null"`
}

var desc string
var amount float32

var yearCmd = &cobra.Command{
	Use:   "year",
	Short: "year command to manage years",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		RootCmd.PersistentPreRun(cmd, args)
		db.AutoMigrate(&Year{})
	},
}

var yearAddCmd = &cobra.Command{
	Use:   "add",
	Short: "add a year to the year table",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 || len(args) > 1 {
			log.Fatal("Need year address")
		}
		year := Year{
			Description: desc,
			Amount:      amount,
		}
		id, _ := strconv.ParseUint(args[0], 10, 32)
		year.ID = uint(id)
		db.Create(&year)
		log.Println("Year created")
	},
}

var yearLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "list alls years",
	Run: func(cmd *cobra.Command, args []string) {
		years := []Year{}
		db.Find(&years)
		fmt.Printf("%4s|%-50s|%4s\n", "Year", "Description", "Amount")
		fmt.Printf("----|--------------------------------------------------|-----------\n")
		for _, year := range years {
			fmt.Printf("%04d|%-50s|%-.2f\n", year.ID, year.Description, year.Amount)
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
		year := Year{}
		db.First(&year, args[0])
		if len(desc) > 0 {
			year.Description = desc
		}
		if amount > 0 {
			year.Amount = amount
		}
		db.Save(&year)
		log.Println("Year updated")
	},
}

func init() {
	RootCmd.AddCommand(yearCmd)
	yearCmd.AddCommand(yearAddCmd)
	yearAddCmd.Flags().StringVar(&desc, "desc", "", "Description of the year")
	yearAddCmd.Flags().Float32Var(&amount, "amount", 30, "Amount of the year")
	yearCmd.AddCommand(yearLsCmd)
	yearCmd.AddCommand(yearUdateCmd)
	yearUdateCmd.Flags().StringVar(&desc, "desc", "", "Description of the year")
	yearUdateCmd.Flags().Float32Var(&amount, "amount", 0, "Amount of the year")
}
