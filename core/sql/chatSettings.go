package sql

type ChatSettings struct {
	ChatId       string `gorm:"primary_key"`
	AllowFilters bool
}

func ShouldAllowFilters(chatId string, shouldAllow bool) {
	w := &ChatSettings{ChatId: chatId}
	tx := SESSION.Begin()
	tx.FirstOrCreate(w)
	w.AllowFilters = shouldAllow
	tx.Save(w)
	tx.Commit()
}

func GetChatSettings(chatId string) *ChatSettings {
	w := ChatSettings{ChatId: chatId}
	SESSION.First(&w)
	return &w
}
