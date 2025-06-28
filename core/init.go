package core

import (
	"reflect"

	"github.com/lugvitc/steve/config"
	"github.com/lugvitc/steve/ext"
	"github.com/lugvitc/steve/ext/context"
	"github.com/lugvitc/steve/ext/handlers"
	"github.com/lugvitc/steve/logger"

	"go.mau.fi/whatsmeow"
)

var LOGGER = logger.NewLogger(logger.LevelInfo)

type Module struct{}

func reply(client *whatsmeow.Client, msg *context.Message, text string) (whatsmeow.SendResponse, error) {
	if msg.Info.IsFromMe {
		return msg.Edit(client, text)
	}
	return msg.Reply(client, text)
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

func Load(dispatcher *ext.Dispatcher) {
	defer LOGGER.Println("Loaded all modules")
	Type := reflect.TypeOf(&Module{})
	Value := reflect.ValueOf(&Module{})
	for i := 0; i < Type.NumMethod(); i++ {
		Type.Method(i).Func.Call([]reflect.Value{Value, reflect.ValueOf(dispatcher)})
	}
}
