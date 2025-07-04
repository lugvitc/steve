package core

import (
	"reflect"
	"strings"

	"github.com/lugvitc/steve/config"
	"github.com/lugvitc/steve/ext"
	"github.com/lugvitc/steve/ext/context"
	"github.com/lugvitc/steve/ext/handlers"
	"github.com/lugvitc/steve/logger"
	"google.golang.org/protobuf/proto"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
)

var LOGGER = logger.NewLogger(logger.LevelInfo)

type Module struct{}

func extractText(msg *context.Message) string {
	args := msg.Args()
	if len(args) <= 1 {
		return ""
	}
	m := msg.Message
	var value string
	if m.Message.ExtendedTextMessage != nil && m.Message.ExtendedTextMessage.ContextInfo != nil {
		qmsg := m.Message.ExtendedTextMessage.ContextInfo.QuotedMessage
		switch {
		case qmsg.Conversation != nil:
			value = qmsg.GetConversation()
		case qmsg.ExtendedTextMessage != nil:
			value = qmsg.ExtendedTextMessage.GetText()
		}
	} else {
		if len(args) == 1 {
			return ""
		}
		value = strings.Join(args[1:], " ")
	}
	return value
}

func reply(client *whatsmeow.Client, msg *context.Message, text string) (whatsmeow.SendResponse, error) {
	if msg.Info.IsFromMe {
		return msg.Edit(client, text)
	}
	var qContext *waE2E.ContextInfo
	if msg.Message.Message.ExtendedTextMessage != nil && msg.Message.Message.ExtendedTextMessage.ContextInfo != nil {
		qContext = msg.Message.Message.ExtendedTextMessage.ContextInfo
	} else {
		qContext = &waE2E.ContextInfo{
			StanzaID:      &msg.Info.ID,
			Participant:   proto.String(msg.Info.Sender.String()),
			QuotedMessage: msg.Message.Message,
			MentionedJID:  []string{msg.Info.Sender.String()},
		}
	}
	return msg.ReplyCustom(client, text, qContext)
}

func authorizedOnly(callback handlers.Response) handlers.Response {
	return func(client *whatsmeow.Client, ctx *context.Context) error {
		if !ctx.Message.Info.IsFromMe && !config.IsSudo(ctx.Message.Info.Sender.User) {
			return ext.EndGroups
		}
		return callback(client, ctx)
	}
}

func authorizedOnlyMessages(callback handlers.Response) handlers.Response {
	return func(client *whatsmeow.Client, ctx *context.Context) error {
		if !ctx.Message.Info.IsFromMe {
			return nil
		}
		return callback(client, ctx)
	}
}
func Load(dispatcher *ext.Dispatcher, client *whatsmeow.Client) { 
	defer LOGGER.Println("Loaded all modules")
	Type := reflect.TypeOf(&Module{})
	Value := reflect.ValueOf(&Module{})
	for i := 0; i < Type.NumMethod(); i++ {
		method := Type.Method(i)
		if method.Type.NumIn() == 3 && method.Type.In(2).String() == "*whatsmeow.Client" {
			method.Func.Call([]reflect.Value{Value, reflect.ValueOf(dispatcher), reflect.ValueOf(client)})
		} else if method.Type.NumIn() == 2 && method.Type.In(1).String() == "*ext.Dispatcher" {
			method.Func.Call([]reflect.Value{Value, reflect.ValueOf(dispatcher)})
		}
	}
}
