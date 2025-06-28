package core

import (
	"github.com/lugvitc/steve/ext"
	"github.com/lugvitc/steve/ext/context"
	"github.com/lugvitc/steve/ext/handlers"
	waLogger "github.com/lugvitc/steve/logger"

	"go.mau.fi/whatsmeow"
)

func ping(client *whatsmeow.Client, ctx *context.Context) error {
	_, _ = reply(client, ctx.Message, "Pong!")
	return ext.EndGroups
}

func (*Module) LoadPing(dispatcher *ext.Dispatcher) {
	ppLogger := LOGGER.Create("ping")
	defer ppLogger.Println("Loaded Ping module")
	dispatcher.AddHandler(
		handlers.NewCommand("ping", authorizedOnly(ping), ppLogger.Create("ping-cmd").
			ChangeLevel(waLogger.LevelInfo),
		).AddDescription("To ping the userbot."),
	)
}
