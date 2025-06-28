package core

import (
	"encoding/ascii85"
	"encoding/base64"
	"strings"

	"github.com/lugvitc/steve/ext"
	"github.com/lugvitc/steve/ext/context"
	"github.com/lugvitc/steve/ext/handlers"
	waLogger "github.com/lugvitc/steve/logger"

	"go.mau.fi/whatsmeow"
)

func b64Encode(_ *whatsmeow.Client, ctx *context.Context) error {
	text := strings.TrimSpace(ctx.Args)
	if text == "" {
		return ctx.Reply("Usage: .b64 <text>")
	}
	encoded := base64.StdEncoding.EncodeToString([]byte(text))
	return ctx.Reply("🔒 Base64:\n" + encoded)
}

func b64Decode(_ *whatsmeow.Client, ctx *context.Context) error {
	text := strings.TrimSpace(ctx.Args)
	if text == "" {
		return ctx.Reply("Usage: .b64d <base64>")
	}
	decoded, err := base64.StdEncoding.DecodeString(text)
	if err != nil {
		return ctx.Reply("❌ Invalid Base64 input.")
	}
	return ctx.Reply("🔓 Decoded:\n" + string(decoded))
}

func b85Encode(_ *whatsmeow.Client, ctx *context.Context) error {
	text := strings.TrimSpace(ctx.Args)
	if text == "" {
		return ctx.Reply("Usage: .b85 <text>")
	}
	buf := make([]byte, ascii85.MaxEncodedLen(len(text)))
	n := ascii85.Encode(buf, []byte(text))
	return ctx.Reply("🔒 Base85:\n" + string(buf[:n]))
}

func b85Decode(_ *whatsmeow.Client, ctx *context.Context) error {
	text := strings.TrimSpace(ctx.Args)
	if text == "" {
		return ctx.Reply("Usage: .b85d <base85>")
	}
	decoded := make([]byte, len(text))
	n, _, err := ascii85.Decode(decoded, []byte(text), true)
	if err != nil {
		return ctx.Reply("❌ Invalid Base85 input.")
	}
	return ctx.Reply("🔓 Decoded:\n" + string(decoded[:n]))
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
