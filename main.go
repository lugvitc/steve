package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/lugvitc/steve/config"
	"github.com/lugvitc/steve/core"
	"github.com/lugvitc/steve/core/sql"
	"github.com/lugvitc/steve/ext"
	"github.com/lugvitc/steve/logger"

	"github.com/mdp/qrterminal/v3"

	_ "github.com/mattn/go-sqlite3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
)

func main() {
	err := config.LoadConfig()
	if err != nil {
		panic(fmt.Errorf("failed to load config: %w", err))
	}
	dbLog := waLog.Stdout("Database", "INFO", true)
	ctx := context.Background()
	// Make sure you add appropriate DB connector imports, e.g. github.com/mattn/go-sqlite3 for SQLite
	container, err := sqlstore.New(ctx, "sqlite3", "file:waub.session?_foreign_keys=on", dbLog)
	if err != nil {
		panic(err)
	}
	// If you want multiple sessions, remember their JIDs and use .GetDevice(jid) or .GetAllDevices() instead.
	deviceStore, err := container.GetFirstDevice(ctx)
	if err != nil {
		panic(err)
	}
	client := whatsmeow.NewClient(deviceStore, waLog.Noop)
	core.LOGGER.ChangeLevel(logger.LevelInfo)
	core.LOGGER.Println("Created new client")
	dispatcher := ext.NewDispatcher(core.LOGGER)
	core.LOGGER.Println("Created new disptacher")
	dispatcher.InitialiseProcessing(ctx, client)
	db := sql.LoadDB(core.LOGGER)
	core.Load(dispatcher, client)

	if client.Store.ID == nil {
		// No ID stored, new login
		qrChan, _ := client.GetQRChannel(ctx)
		err = client.Connect()
		if err != nil {
			panic(err)
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				// Render the QR code here
				config := qrterminal.Config{
					Level:          qrterminal.L,
					Writer:         os.Stdout,
					HalfBlocks:     true,
					BlackChar:      qrterminal.BLACK_BLACK,
					WhiteBlackChar: qrterminal.WHITE_BLACK,
					WhiteChar:      qrterminal.WHITE_WHITE,
					BlackWhiteChar: qrterminal.BLACK_WHITE,
					QuietZone:      1,
				}
				qrterminal.GenerateWithConfig(evt.Code, config)
			} else {
				fmt.Println("Login event:", evt.Event)
			}
		}
	} else {
		// Already logged in, just connect
		err = client.Connect()
		if err != nil {
			panic(err)
		}
	}
	core.LOGGER.ChangeLevel(logger.LevelMain)
	core.LOGGER.Println("Whatsapp Userbot", "has been started...")
	// Listen to Ctrl+C (you can also do something else that prevents the program from exiting)
	c := make(chan os.Signal, 1) // buffered channel
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	//_ = client.SendPresence(types.PresenceUnavailable)
	client.Disconnect()
	if err := db.Close(); err != nil {
		core.LOGGER.ChangeLevel(logger.LevelError).Panicln("failed to close db:", err.Error())
	}
}
