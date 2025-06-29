package core

import (
	"encoding/ascii85"
	"encoding/base64"

	"github.com/lugvitc/steve/ext"
	"github.com/lugvitc/steve/ext/context"
	"github.com/lugvitc/steve/ext/handlers"

	"go.mau.fi/whatsmeow"
)

func b64Encode(client *whatsmeow.Client, ctx *context.Context) error {
	text := extractText(ctx.Message)
	if text == "" {
		_, err := reply(client, ctx.Message, "Usage: .b64 <text>")
		return err
	}
	encoded := base64.StdEncoding.EncodeToString([]byte(text))
	_, err := reply(client, ctx.Message, "🔒 Base64:\n"+encoded)
	return err
}

func b64Decode(client *whatsmeow.Client, ctx *context.Context) error {
	text := extractText(ctx.Message)
	if text == "" {
		_, err := reply(client, ctx.Message, "Usage: .b64d <base64>")
		return err
	}
	decoded, err := base64.StdEncoding.DecodeString(text)
	if err != nil {
		_, err := reply(client, ctx.Message, "❌ Invalid Base64 input.")
		return err
	}
	_, err = reply(client, ctx.Message, "🔓 Decoded:\n"+string(decoded))
	return err
}

func b85Encode(client *whatsmeow.Client, ctx *context.Context) error {
	text := extractText(ctx.Message)
	if text == "" {
		_, err := reply(client, ctx.Message, "Usage: .b85 <text>")
		return err
	}
	buf := make([]byte, ascii85.MaxEncodedLen(len(text)))
	n := ascii85.Encode(buf, []byte(text))
	_, err := reply(client, ctx.Message, "🔒 Base85:\n"+string(buf[:n]))
	return err
}

func b85Decode(client *whatsmeow.Client, ctx *context.Context) error {
	text := extractText(ctx.Message)
	if text == "" {
		_, err := reply(client, ctx.Message, "Usage: .b85d <base85>")
		return err
	}
	decoded := make([]byte, len(text))
	n, _, err := ascii85.Decode(decoded, []byte(text), true)
	if err != nil {
		_, err := reply(client, ctx.Message, "❌ Invalid Base85 input.")
		return err
	}
	_, err = reply(client, ctx.Message, "🔓 Decoded:\n"+string(decoded[:n]))
	return err
}

func (*Module) LoadBase(dispatcher *ext.Dispatcher) {
	log := LOGGER.Create("base")
	defer log.Println("Loaded Base Encode/Decode module")

	dispatcher.AddHandler(
		handlers.NewCommand("b64", authorizedOnly(b64Encode), log.Create("b64")).
			AddDescription("Encode text to Base64."),
	)
	dispatcher.AddHandler(
		handlers.NewCommand("b64d", authorizedOnly(b64Decode), log.Create("b64d")).
			AddDescription("Decode Base64 to text."),
	)
	dispatcher.AddHandler(
		handlers.NewCommand("b85", authorizedOnly(b85Encode), log.Create("b85")).
			AddDescription("Encode text to Base85."),
	)
	dispatcher.AddHandler(
		handlers.NewCommand("b85d", authorizedOnly(b85Decode), log.Create("b85d")).
			AddDescription("Decode Base85 to text."),
	)
}
