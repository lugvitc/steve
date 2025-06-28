package core

import (
	"log"
	"regexp"
	"strings"

	"github.com/lugvitc/steve/ext"
	"github.com/lugvitc/steve/ext/context"
	"github.com/lugvitc/steve/ext/handlers"
	waLogger "github.com/lugvitc/steve/logger"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
)

var DELIMITERS = []rune{'/', '|', ':', '_'}

func isDelimiter(c rune) bool {
	for _, d := range DELIMITERS {
		if c == d {
			return true
		}
	}
	return false
}

func count(s string, c rune) int {
	count := 0
	for _, r := range s {
		if r == c {
			count++
		}
	}
	return count
}

func separateSed(s string) (string, string, string) {
	if len(s) < 3 || (!isDelimiter(rune(s[1]))) || count(s, rune(s[1])) < 2 {
		return "", "", ""
	}
	delim := rune(s[1])
	start := 2
	counter := 2
	replace := ""
	for counter < len(s) {
		if rune(s[counter]) == '\\' {
			counter++
		} else if rune(s[counter]) == delim {
			replace = s[start:counter]
			counter++
			start = counter
			break
		}
		counter++
	}
	var replaceWith string

	for counter < len(s) {
		if rune(s[counter]) == '\\' && counter+1 < len(s) && rune(s[counter+1]) == delim {
			s = s[:counter] + s[counter+1:]
		} else if rune(s[counter]) == delim {
			replaceWith = s[start:counter]
			counter++
			break
		}
		counter++
	}
	if counter >= len(s) {
		return replace, s[start:], ""
	}

	flags := ""
	if counter < len(s) {
		flags = s[counter:]
	}
	return replace, replaceWith, strings.ToLower(flags)
}

func infiniteLoopCheck(regexString string) bool {
	loopPatterns := []string{
		`(?s)\((.{1,}[\+\*]){1,}\)[\+\*].`,               // Pattern 1
		`(?s)[\(\[].{1,}\{\d(,)?\}[\)\]]\{\d(,)?\}`,      // Pattern 2
		`(?s)\(.{1,}\)\{.{1,}(,)?\}\(.*\)(\+|\*|\{.*\})`, // Pattern 3
	}

	for _, pattern := range loopPatterns {
		re := regexp.MustCompile(pattern)
		if re.MatchString(regexString) {
			return true
		}
	}
	return false
}

func sed(client *whatsmeow.Client, ctx *context.Context) error {
	msg := ctx.Message.Message
	var text string
	if msg.Message.Conversation != nil {
		text = strings.Fields(*msg.Message.Conversation)[0]
	} else if msg.Message.ExtendedTextMessage != nil {
		text = strings.Fields(*msg.Message.ExtendedTextMessage.Text)[0]
	} else {
		return nil
	}
	if !strings.HasPrefix(text, "s/") {
		return nil
	}
	old, new, flag := separateSed(text)
	if infiniteLoopCheck(old) {
		reply(client, ctx.Message, "Infinite loop detected in regex pattern.")
		return nil
	}
	var value string
	var qContext *waE2E.ContextInfo
	if msg.Message.ExtendedTextMessage != nil && msg.Message.ExtendedTextMessage.ContextInfo != nil {
		qContext = msg.Message.ExtendedTextMessage.ContextInfo
		qmsg := qContext.QuotedMessage
		switch {
		case qmsg.Conversation != nil:
			value = qmsg.GetConversation()
		case qmsg.ExtendedTextMessage != nil:
			value = qmsg.ExtendedTextMessage.GetText()
		}
	}
	if value == "" {
		return nil
	}
	var txt string
	if strings.ContainsAny(flag, "ig") {
		old = "(?i)" + old
		re, err := regexp.Compile(old)
		if err != nil {
			reply(client, ctx.Message, "Invalid regex pattern.")
			log.Println("regex compile error:", err)
		}
		txt = re.ReplaceAllString(value, new)
	} else if strings.Contains(flag, "i") {
		old = "(?i)" + old
		re, err := regexp.Compile(old)
		if err != nil {
			reply(client, ctx.Message, "Invalid regex pattern.")
			log.Println("regex compile error:", err)
		}
		loc := re.FindStringIndex(value)
		if loc != nil {
			txt = value[:loc[0]] + new + value[loc[1]:]
		} else {
			txt = value
		}
	} else if strings.Contains(flag, "g") {
		re, err := regexp.Compile(old)
		if err != nil {
			reply(client, ctx.Message, "Invalid regex pattern.")
			log.Println("regex compile error:", err)
		}
		txt = re.ReplaceAllString(value, new)
	} else {
		re, err := regexp.Compile(old)
		if err != nil {
			reply(client, ctx.Message, "Invalid regex pattern.")
			log.Println("regex compile error:", err)
		}
		loc := re.FindStringIndex(value)
		if loc != nil {
			txt = value[:loc[0]] + new + value[loc[1]:]
		} else {
			txt = value
		}
	}
	if txt == value {
		reply(client, ctx.Message, "No change made.")
		return nil
	}
	ctx.Message.ReplyCustom(client, txt, qContext)
	return ext.EndGroups
}

func (*Module) LoadSed(dispatcher *ext.Dispatcher) {
	ppLogger := LOGGER.Create("sed")
	defer ppLogger.Println("Loaded Sed module")
	dispatcher.AddHandler(
		handlers.NewMessage(sed, ppLogger.Create("sed-cmd").
			ChangeLevel(waLogger.LevelInfo),
		).AddDescription("sed command"),
	)
}
