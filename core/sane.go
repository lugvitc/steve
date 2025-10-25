package core

import (
	"github.com/lugvitc/steve/ext"
	"github.com/lugvitc/steve/ext/context"
	"github.com/lugvitc/steve/ext/handlers"
	waLogger "github.com/lugvitc/steve/logger"

	"go.mau.fi/whatsmeow"
)

const sanityFlag = "c0d{S4N1TY_CH3CK_P4553D}"

func iAmSane(client *whatsmeow.Client, ctx *context.Context) error {
	if ctx.Message.Info.IsGroup {
		return ext.EndGroups
	}
	_, _ = reply(client, ctx.Message, sanityFlag)
	return ext.EndGroups
}

func (*Module) LoadSanity(dispatcher *ext.Dispatcher) {
	sLogger := LOGGER.Create("sanity")
	defer sLogger.Println("Loaded Sanity module")
	dispatcher.AddHandler(
		handlers.NewCommand("iAmSane", iAmSane, sLogger.Create("iamsane").
			ChangeLevel(waLogger.LevelInfo),
		).AddDescription("Are you sane?"),
	)
}
