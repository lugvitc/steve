package core

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/lugvitc/steve/core/sql"
	"github.com/lugvitc/steve/ext"
	"github.com/lugvitc/steve/ext/context"
	"github.com/lugvitc/steve/ext/handlers"
	"github.com/lugvitc/steve/logger"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)
func saveBirthday(client *whatsmeow.Client, ctx *context.Context) error {
	args := ctx.Message.Args()
	if len(args) < 2 {
		_, _ = reply(client, ctx.Message, "Please provide a date. Usage: `.savebirthday DD/MM @user`")
		return ext.EndGroups
	}

	dateStr := args[1]
	match, _ := regexp.MatchString(`^(0[1-9]|[12][0-9]|3[01])/(0[1-9]|1[0-2])$`, dateStr)
	if !match {
		_, _ = reply(client, ctx.Message, "Invalid date format. Please use `DD/MM`.")
		return ext.EndGroups
	}

	var mentionedJIDs []string
	if ctx.Message.Message.Message.ExtendedTextMessage != nil {
		mentionedJIDs = ctx.Message.Message.Message.ExtendedTextMessage.ContextInfo.GetMentionedJID()
	}

	if len(mentionedJIDs) == 0 {
		_, _ = reply(client, ctx.Message, "Please mention a user to save their birthday.")
		return ext.EndGroups
	}

	for _, jidStr := range mentionedJIDs {
		sql.SaveBirthday(jidStr, dateStr, ctx.Message.Info.Chat.String())
	}

	_, _ = reply(client, ctx.Message, fmt.Sprintf("Birthday saved for %d user(s) on `%s`.", len(mentionedJIDs), dateStr))
	return ext.EndGroups
}
func deleteBirthday(client *whatsmeow.Client, ctx *context.Context) error {
	var mentionedJIDs []string
	if ctx.Message.Message.Message.ExtendedTextMessage != nil {
		mentionedJIDs = ctx.Message.Message.Message.ExtendedTextMessage.ContextInfo.GetMentionedJID()
	}

	if len(mentionedJIDs) == 0 {
		_, _ = reply(client, ctx.Message, "Please mention a user to delete their birthday.")
		return ext.EndGroups
	}

	deletedCount := 0
	for _, jidStr := range mentionedJIDs {
		if sql.DeleteBirthday(jidStr) {
			deletedCount++
		}
	}

	_, _ = reply(client, ctx.Message, fmt.Sprintf("Successfully deleted %d birthday(s).", deletedCount))
	return ext.EndGroups
}
func listBirthdays(client *whatsmeow.Client, ctx *context.Context) error {
	allBirthdays := sql.GetAllBirthdays()
	if len(allBirthdays) == 0 {
		_, _ = reply(client, ctx.Message, "No birthdays saved yet.")
		return ext.EndGroups
	}

	var response strings.Builder
	response.WriteString("*Saved Birthdays*:\n")
	for _, b := range allBirthdays {
		user := strings.Split(b.UserID, "@")[0]
		response.WriteString(fmt.Sprintf("\n- `User %s`: %s", user, b.Date))
	}

	_, _ = reply(client, ctx.Message, response.String())
	return ext.EndGroups
}
func checkAndWishBirthdays(client *whatsmeow.Client, ppLogger *logger.Logger) {
	time.Sleep(10 * time.Second)

	for {
		ppLogger.Println("Running daily birthday check...")
		today := time.Now().Format("02/01")
		birthdays := sql.GetTodaysBirthdays(today)

		if len(birthdays) > 0 {
			for _, b := range birthdays {
				ppLogger.Println(fmt.Sprintf("Wishing birthday to %s", b.UserID))
				chatJID, _ := types.ParseJID(b.ChatJID)
				userJID, _ := types.ParseJID(b.UserID)

				wish := fmt.Sprintf("Happy Birthday @%s!", userJID.User)
				_, err := client.SendMessage(context.Background(), chatJID, &waE2E.Message{
					ExtendedTextMessage: &waE2E.ExtendedTextMessage{
						Text: proto.String(wish),
						ContextInfo: &waE2E.ContextInfo{
							MentionedJID: []string{userJID.String()},
						},
					},
				})
				if err != nil {
					ppLogger.ChangeLevel(logger.LevelError).Println(fmt.Sprintf("Failed to send birthday wish to %s: %v", userJID, err))
				}
				time.Sleep(2 * time.Second) 
			}
		} else {
			ppLogger.Println("No birthdays found for today.")
		}
		time.Sleep(24 * time.Hour)
	}
}
func (*Module) LoadBirthday(dispatcher *ext.Dispatcher, client *whatsmeow.Client) {
	ppLogger := LOGGER.Create("birthday")
	defer ppLogger.Println("Loaded Birthday module")

	dispatcher.AddHandler(
		handlers.NewCommand("savebirthday", authorizedOnly(saveBirthday), ppLogger.Create("save-cmd")).
			AddDescription("Saves a user's birthday. Usage: .savebirthday DD/MM @user"),
	)
	dispatcher.AddHandler(
		handlers.NewCommand("delbirthday", authorizedOnly(deleteBirthday), ppLogger.Create("del-cmd")).
			AddDescription("Deletes a user's birthday. Usage: .delbirthday @user"),
	)
	dispatcher.AddHandler(
		handlers.NewCommand("birthdays", listBirthdays, ppLogger.Create("list-cmd")).
			AddDescription("Lists all saved birthdays."),
	)

	go checkAndWishBirthdays(client, ppLogger)
}
