package main

import (
	"encoding/json"
	"errors"
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

var firstTimeSetup = false

func main() {
	service := "ssm-email"
	accountList, err := keyring.Get(service, "accounts")
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
	// Get the GtkBuilder UI definition in the glade file.
	builder, err := gtk.BuilderNewFromFile("ui/SSM.glade")
	errorCheck(err)

	// Map the handlers to callback functions, and connect the signals
	// to the Builder.
	signals := map[string]interface{}{
		"on_main_window_destroy": onMainWindowDestroy,
	}
	builder.ConnectSignals(signals)

	// Get the object with the id of "main_window".
	obj, err := builder.GetObject("MainWindow")
	errorCheck(err)

	// Verify that the object is a pointer to a gtk.ApplicationWindow.
	win, err := isWindow(obj)
	errorCheck(err)

	// Show the Window and all of its components.
	win.Show()
	application.AddWindow(win)

}
func onMainWindowDestroy() {
	log.Println("onMainWindowDestroy")
}
func isWindow(obj glib.IObject) (*gtk.Window, error) {
	// Make type assertion (as per gtk.go).
	if win, ok := obj.(*gtk.Window); ok {
		return win, nil
	}
	return nil, errors.New("not a *gtk.Window")
}
func errorCheck(e error) {
	if e != nil {
		// panic for any errors.
		log.Panic(e)
	}
}
