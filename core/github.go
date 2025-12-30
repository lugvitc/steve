package core

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/lugvitc/steve/config"
	"github.com/lugvitc/steve/ext"
	"github.com/lugvitc/steve/logger"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
)

func (m *Module) LoadGithub(dispatcher *ext.Dispatcher) {
	ppLogger := LOGGER.Create("github")
	defer ppLogger.Println("Loaded Github module")
	http.HandleFunc("/github/webhook", githubWebhookHandler(m.client, ppLogger))
	go func() {
		err := http.ListenAndServe(fmt.Sprintf(":%d", config.GetGithubWebhookPort()), nil)
		if err != nil {
			ppLogger.ChangeLevel(logger.LevelError).Println("Failed to start GitHub webhook server:", err.Error())
		}
	}()
}

func githubWebhookHandler(client *whatsmeow.Client, ppLogger *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		event := r.Header.Get("X-GitHub-Event")
		delivery := r.Header.Get("X-GitHub-Delivery")

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read body", http.StatusBadRequest)
			return
		}

		log.Printf("📩 Event=%s Delivery=%s\n", event, delivery)

		msg := HandleGitHubEvent(event, ppLogger, body)
		if msg != "" {
			ppLogger.Println("Sending GitHub event message")
			ppLogger.Println(msg)
			Send(client, ppLogger, msg)
		}

		w.WriteHeader(http.StatusOK)
	}
}

func HandleGitHubEvent(event string, ppLogger *logger.Logger, body []byte) string {
	switch event {

	case "push":
		return handlePush(body, ppLogger)

	case "pull_request":
		return handlePullRequest(body, ppLogger)

	case "issues":
		return handleIssue(body, ppLogger)

	case "release":
		return handleRelease(body, ppLogger)

	default:
		return handleGeneric(event, body, ppLogger)
	}
}

func handlePush(body []byte, ppLogger *logger.Logger) string {
	var p struct {
		Ref        string `json:"ref"`
		Repository struct {
			FullName string `json:"full_name"`
		} `json:"repository"`
		Pusher struct {
			Name string `json:"name"`
		} `json:"pusher"`
		Commits []struct {
			Message string `json:"message"`
		} `json:"commits"`
		HeadCommit struct {
			Message string `json:"message"`
			URL     string `json:"url"`
		} `json:"head_commit"`
	}

	if err := json.Unmarshal(body, &p); err != nil {
		ppLogger.ChangeLevel(logger.LevelError).Println("Failed to unmarshal push event:", err.Error())
		return ""
	}

	branch := strings.TrimPrefix(p.Ref, "refs/heads/")

	return fmt.Sprintf(
		"🚀 *Push Event*\n"+
			"📦 Repo: %s\n"+
			"👤 Pusher: %s\n"+
			"🌿 Branch: %s\n"+
			"🧾 Commits: %d\n"+
			"📝 Latest: %s\n"+
			"🔗 %s",
		p.Repository.FullName,
		p.Pusher.Name,
		branch,
		len(p.Commits),
		p.HeadCommit.Message,
		p.HeadCommit.URL,
	)
}

func handlePullRequest(body []byte, ppLogger *logger.Logger) string {
	var p struct {
		Action      string `json:"action"`
		PullRequest struct {
			Title   string `json:"title"`
			HTMLURL string `json:"html_url"`
			User    struct {
				Login string `json:"login"`
			} `json:"user"`
			Base struct {
				Ref string `json:"ref"`
			} `json:"base"`
			Head struct {
				Ref string `json:"ref"`
			} `json:"head"`
		} `json:"pull_request"`
		Repository struct {
			FullName string `json:"full_name"`
		} `json:"repository"`
	}

	if err := json.Unmarshal(body, &p); err != nil {
		ppLogger.ChangeLevel(logger.LevelError).Println("Failed to unmarshal pull request event:", err.Error())
		return ""
	}

	return fmt.Sprintf(
		"✨ *Pull Request %s*\n"+
			"📦 Repo: %s\n"+
			"👤 Author: %s\n"+
			"🔀 %s → %s\n"+
			"📝 %s\n"+
			"🔗 %s",
		strings.Title(p.Action),
		p.Repository.FullName,
		p.PullRequest.User.Login,
		p.PullRequest.Head.Ref,
		p.PullRequest.Base.Ref,
		p.PullRequest.Title,
		p.PullRequest.HTMLURL,
	)
}

func handleIssue(body []byte, ppLogger *logger.Logger) string {
	var p struct {
		Action string `json:"action"`
		Issue  struct {
			Title   string `json:"title"`
			HTMLURL string `json:"html_url"`
			Body    string `json:"body"`
			User    struct {
				Login string `json:"login"`
			} `json:"user"`
		} `json:"issue"`
		Repository struct {
			FullName string `json:"full_name"`
		} `json:"repository"`
	}
	err := json.Unmarshal(body, &p)
	if err != nil {
		ppLogger.ChangeLevel(logger.LevelError).Println("Failed to unmarshal issue event:", err.Error())
		return ""
	}

	return fmt.Sprintf(
		"🐞 *Issue %s*\n📦 %s\n👤 %s\n📝 %s\n🔗 %s\n\n%s",
		p.Action,
		p.Repository.FullName,
		p.Issue.User.Login,
		p.Issue.Title,
		p.Issue.HTMLURL,
		p.Issue.Body,
	)
}

func handleRelease(body []byte, ppLogger *logger.Logger) string {
	var p map[string]any
	json.Unmarshal(body, &p)

	repo := p["repository"].(map[string]any)["full_name"]
	action := p["action"]
	tag := p["release"].(map[string]any)["tag_name"]
	url := p["release"].(map[string]any)["html_url"]

	return fmt.Sprintf(
		"🏷️ *Release %v*\n📦 %v\n🔖 Tag: %v\n🔗 %v",
		action, repo, tag, url,
	)
}

func handleGeneric(event string, body []byte, ppLogger *logger.Logger) string {
	var p map[string]any
	if json.Unmarshal(body, &p) != nil {
		return ""
	}

	repo := "unknown"
	if r, ok := p["repository"].(map[string]any); ok {
		if name, ok := r["full_name"].(string); ok {
			repo = name
		}
	}

	sender := "unknown"
	if s, ok := p["sender"].(map[string]any); ok {
		if login, ok := s["login"].(string); ok {
			sender = login
		}
	}

	return fmt.Sprintf(
		"📦 *GitHub Event*\n"+
			"📣 Type: %s\n"+
			"📦 Repo: %s\n"+
			"👤 Sender: %s",
		event, repo, sender,
	)
}

func Send(client *whatsmeow.Client, ppLogger *logger.Logger, message string) {
	jid, _ := types.ParseJID(config.GetConfig().DeliveryJID)
	_, err := client.SendMessage(context.Background(), jid, &waE2E.Message{
		Conversation: &message,
	})
	if err != nil {
		ppLogger.ChangeLevel(logger.LevelError).Println("Failed to send GitHub message:", err.Error())
	}
}
