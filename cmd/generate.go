package main

import (
	"bufio"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/blackreloaded/go-secret-santa"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var server string
var from string
var template string
var auth bool
var send bool

func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.Flags().BoolVar(&auth, "auth", true, "Smtp need auth")
	generateCmd.Flags().BoolVar(&send, "send", false, "Send mails")
	generateCmd.Flags().StringVar(&server, "server", "", "Mail server")
	generateCmd.Flags().StringVar(&from, "from", "", "Mail sender")
	generateCmd.Flags().StringVar(&template, "template", "", "Text template for mail")
}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "generate command to create a year an write the output",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.PersistentPreRun(cmd, args)
	},
	Run: func(cmd *cobra.Command, args []string) {
		year := &secretsanta.Year{
			YearID: uint32(time.Now().Year()),
			Amount: 30.0,
		}
		if len(args) > 0 {
			yearid, err := strconv.Atoi(args[0])
			if err != nil {
				log.Fatal("failed to parse yearid")
			}
			year.YearID = uint32(yearid)
		}
		if len(args) > 1 {
			amount, err := strconv.ParseFloat(args[1], 32)
			if err != nil {
				log.Fatalf("failed to parse amount: %v", err)
			}
			year.Amount = float32(amount)
		}
		if len(args) > 2 {
			year.Description = args[2]
		}
		err := secretSanta.Generate(year, false, 0)
		if err != nil {
			log.Fatalf("Failed to generate year: %v", err)
		}
		err = secretSanta.UpdateYear(year)
		if err != nil {
			log.Fatalf("failed to save year: %v", err)
		}
		if send {
			idx := strings.Index(server, ":")
			if idx < 0 {
				log.Fatal("--server must be in the form host:port")
			}
			hostname := server[:idx]
			var username string
			var password string
			if auth {
				reader := bufio.NewReader(os.Stdin)

				fmt.Print("Enter Username: ")
				username, err = reader.ReadString('\n')
				if err != nil {
					log.Fatalf("failed to read username: %v", err)
				}

				fmt.Print("Enter Password: ")
				bytePassword, err := term.ReadPassword(int(syscall.Stdin))
				if err != nil {
					log.Fatalf("failed to read password: %v", err)
				}
				password = string(bytePassword)

				username, password = strings.TrimSpace(username), strings.TrimSpace(password)
			}
			if from == "" {
				from = username
			}
			log.Printf("year generated, sending messages with %s\n", hostname)
			session := &mailSession{
				from:   from,
				server: server,
				year:   year.YearID,
				auth: smtp.PlainAuth(
					"",
					username,
					password,
					hostname,
				),
			}
			err = secretSanta.SendInformation(year, template, session.sendInfoMail)
			if err != nil {
				log.Fatalf("failed to send year information: %v", err)
			}
		} else {
			secretSanta.SendInformation(year, template, func(to, text string) error {
				_, err := fmt.Printf("%s\n%s", to, text)
				return err
			})
		}
	},
}

type mailSession struct {
	auth   smtp.Auth
	server string
	from   string
	year   uint32
}

func (ms *mailSession) sendInfoMail(to, text string) error {
	log.Printf("Send mail to %s", to)
	// Connect to the server, authenticate, set the sender and recipient,
	// and send the email all in one step.
	date := time.Now().Format("Mon, 2 Jan 2006 15:04:05 -0700")
	mail := fmt.Sprintf("Subject: Wichteln %d\r\nFrom:%s\r\nDate:%s\r\nTo:%s\r\n\r\n%s", ms.year, ms.from, date, to, text)
	return smtp.SendMail(
		ms.server,
		ms.auth,
		ms.from,
		[]string{to},
		[]byte(mail),
	)
}
