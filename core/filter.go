package core

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/lugvitc/steve/core/sql"
	"github.com/lugvitc/steve/ext"
	"github.com/lugvitc/steve/ext/context"
	"github.com/lugvitc/steve/ext/handlers"
	waLogger "github.com/lugvitc/steve/logger"

	"go.mau.fi/whatsmeow"
)

func addFilter(dispatcher *ext.Dispatcher, dispatcherGroup int) func(
	client *whatsmeow.Client, ctx *context.Context,
) error {
	return func(client *whatsmeow.Client, ctx *context.Context) error {
		args := ctx.Message.Args()
		if len(args) <= 1 {
			return ext.EndGroups
		}
		msg := ctx.Message.Message
		key := args[1]
		var value string
		if msg.Message.ExtendedTextMessage != nil && msg.Message.ExtendedTextMessage.ContextInfo != nil {
			qmsg := msg.Message.ExtendedTextMessage.ContextInfo.QuotedMessage
			switch {
			case qmsg.Conversation != nil:
				value = qmsg.GetConversation()
			case qmsg.ExtendedTextMessage != nil:
				value = qmsg.ExtendedTextMessage.GetText()
			}
		} else {
			if len(args) == 2 {
				return ext.EndGroups
			}
			value = strings.Join(args[2:], " ")
		}
		go func() {
			sql.AddFilter(strings.ToLower(key), value)
			initFilter(ctx.Logger, &sql.Filter{Name: key, Value: value}, dispatcher, dispatcherGroup)
		}()
		_, _ = reply(client, ctx.Message, fmt.Sprintf("Added Filter ```%s```.", key))
		return ext.EndGroups
	}
}

func removeFilter(client *whatsmeow.Client, ctx *context.Context) error {
	args := ctx.Message.Args()
	if len(args) == 1 {
		return ext.EndGroups
	}
	key := args[1]
	if sql.DeleteFilter(strings.ToLower(key)) {
		_, _ = reply(client, ctx.Message, fmt.Sprintf("Successfully delete filter '```%s```'.", key))
	} else {
		_, _ = reply(client, ctx.Message, "Failed to delete that filter!")
	}
	return ext.EndGroups
}

func initFilter(l *waLogger.Logger, filter *sql.Filter, dispatcher *ext.Dispatcher, dispatcherGroup int) {
	dispatcher.AddHandlerToGroup(
		handlers.NewMessage(
			func(client *whatsmeow.Client, ctx *context.Context) error {
				if ctx.Message.Info.IsFromMe {
					return nil
				}
				if ctx.Message.Info.IsGroup && !sql.GetChatSettings(ctx.Message.Info.Chat.String()).AllowFilters {
					return nil
				}
				text := ctx.Message.GetText()
				match, _ := regexp.MatchString(
					fmt.Sprintf(`(\s|^)%s(\s|$)`, filter.Name),
					strings.ToLower(text),
				)
				if !match {
					return nil
				}
				_, _ = ctx.Message.Reply(client, filter.Value)
				return nil
			},
			l.Create("filter-"+filter.Name).ChangeLevel(waLogger.LevelInfo),
		),
		dispatcherGroup,
	)
}

func loadFilters(l *waLogger.Logger, dispatcher *ext.Dispatcher, dispatcherGroup int) {
	for _, filter := range sql.GetFilters() {
		initFilter(l, &filter, dispatcher, dispatcherGroup)
	}
}

func listFilters(client *whatsmeow.Client, ctx *context.Context) error {
	text := "*List of filters*:"
	for _, filter := range sql.GetFilters() {
		text += fmt.Sprintf("\n- ```%s```", filter.Name)
	}
	if text == "*List of filters*:" {
		text = "No filters are present."
	}
	_, _ = reply(client, ctx.Message, text)
	return ext.EndGroups
}

func allowFilter(client *whatsmeow.Client, ctx *context.Context) error {
	chatId := ctx.Message.Info.Chat.String()
	args := ctx.Message.Args()
	var allowFilters bool
	var text = "Allowed filters in this chat."
	if len(args) == 1 {
		allowFilters = true
	} else {
		switch strings.ToLower(args[1]) {
		case "true", "yes", "on":
			allowFilters = true
		case "false", "no", "off":
			allowFilters = false
			text = "Disallowed filters in this chat."
		default:
			return ext.EndGroups
		}
	}
	sql.ShouldAllowFilters(chatId, allowFilters)
	reply(client, ctx.Message, text)
	return ext.EndGroups
}

func (*Module) LoadFilter(dispatcher *ext.Dispatcher) {
	ppLogger := LOGGER.Create("filter")
	defer ppLogger.Println("Loaded Filter module")
	dispatcher.AddHandler(
		handlers.NewCommand("filter", authorizedOnly(addFilter(dispatcher, 1)), ppLogger.Create("filter-cmd").
			ChangeLevel(waLogger.LevelInfo),
		).AddDescription("Add a filter."),
	)
	dispatcher.AddHandler(
		handlers.NewCommand("allowfilters", authorizedOnly(allowFilter), ppLogger.Create("allowfilters-cmd").
			ChangeLevel(waLogger.LevelInfo),
		).AddDescription("Allow filters in a group."),
	)
	dispatcher.AddHandler(
		handlers.NewCommand("remove", authorizedOnly(removeFilter), ppLogger.Create("remove-cmd").
			ChangeLevel(waLogger.LevelInfo),
		).AddDescription("Remove a filter."),
	)
	dispatcher.AddHandler(
		handlers.NewCommand("filters", authorizedOnly(listFilters), ppLogger.Create("list-filters").
			ChangeLevel(waLogger.LevelInfo),
		).AddDescription("Display keys of all saved filters."),
	)
	loadFilters(
		ppLogger.Create("filter-cmd").
			ChangeLevel(waLogger.LevelInfo),
		dispatcher,
		1,
	)
	// take you back
}
