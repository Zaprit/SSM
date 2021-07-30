package main

import (
	"encoding/json"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/zalando/go-keyring"
	"log"
	"os"
)

type account struct {
	Username string
	Password string
}

func main() {
	service := "ssm-email"
	accountList, err := keyring.Get(service, "accounts")
	firstTimeSetup := false
	if err != nil {
		print("account list not found, asking user some dumb questions about e-mail")
		firstTimeSetup = true
	}
	users := []string{}
	if accountList != "" {
		err = json.Unmarshal([]byte(accountList), &users)
		if err != nil {
			log.Fatal(err)
		}
	}
	accounts := []account{}
	for _, acc := range users {
		pass, err := keyring.Get(service, acc)
		if err != nil {
			log.Fatal(err)
		}
		accounts = append(accounts, account{Username: acc, Password: pass})
	}
	print(accountList)

	print(firstTimeSetup)

	const appID = "org.gtk.example"
	application, err := gtk.ApplicationNew(appID, glib.APPLICATION_FLAGS_NONE)
	// Check to make sure no errors when creating Gtk Application
	if err != nil {
		log.Fatal("Could not create application.", err)
	}
	application.Connect("activate", func() { onActivate(application) })
	// Run Gtk application
	os.Exit(application.Run(os.Args))
}
func onActivate(application *gtk.Application) {
	// Create ApplicationWindow
	appWindow, err := gtk.ApplicationWindowNew(application)
	if err != nil {
		log.Fatal("Could not create application window.", err)
	}
	// Set ApplicationWindow Properties
	appWindow.SetTitle("")
	appWindow.SetDefaultSize(400, 400)
	appWindow.Show()
}

func mailTest() {
	log.Println("Connecting to server...")

	// Connect to server
	c, err := client.DialTLS("mail.valtek.uk:993", nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected")

	// Don't forget to logout
	defer func(c *client.Client) {
		err := c.Logout()
		if err != nil {
			print(err)
		}
	}(c)

	// Login
	if err := c.Login("garbagecan@valtek.uk", ":,87,arNACHiE"); err != nil {
		log.Fatal(err)
	}
	log.Println("Logged in")

	// List mailboxes
	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.List("", "*", mailboxes)
	}()

	log.Println("Mailboxes:")
	for m := range mailboxes {
		log.Println("* " + m.Name)
	}

	if err := <-done; err != nil {
		log.Fatal(err)
	}

	// Select INBOX
	mbox, err := c.Select("INBOX", false)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Flags for INBOX:", mbox.Flags)

	// Get the last 4 messages
	from := uint32(1)
	to := mbox.Messages
	if mbox.Messages > 3 {
		// We're using unsigned integers here, only subtract if the result is > 0
		from = mbox.Messages - 3
	}
	seqset := new(imap.SeqSet)
	seqset.AddRange(from, to)

	messages := make(chan *imap.Message, 10)
	done = make(chan error, 1)
	go func() {
		done <- c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope}, messages)
	}()

	log.Println("Last 4 messages:")
	for msg := range messages {
		log.Println("* " + msg.Envelope.Subject)
	}

	if err := <-done; err != nil {
		log.Fatal(err)
	}

	log.Println("Done!")
}
