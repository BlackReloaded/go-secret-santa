package main

import (
	"log"

	"github.com/blackreloaded/go-secret-santa"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.PersistentFlags().StringVar(&dbFile, "dbfile", "./secretsanta.db", "db file for users")
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}

var dbFile string
var secretSanta *secretsanta.SecretSanta

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "secretsanta",
	Short: "secretsanta is a little programm to generate pairs of user",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		log.Printf("Using database \"%s\"\n", dbFile)
		var err error
		secretSanta, err = secretsanta.New(dbFile)
		if err != nil {
			log.Fatal(err)
		}
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		err := secretSanta.Close()
		if err != nil {
			log.Fatal(err)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Usage()
	},
}
