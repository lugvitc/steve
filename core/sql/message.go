package sql

type Message struct {
	ID        string `gorm:"primary_key"`
	Timestamp int64
	PushName  string
	Sender    string
}

func AddMessage(id, pushName, sender string, timestamp int64) {
	tx := SESSION.Begin()
	w := &Message{ID: id, Timestamp: timestamp, PushName: pushName, Sender: sender}
	tx.Create(w)
	tx.Commit()
}

func GetMessage(id string) *Message {
	w := Message{ID: id}
	SESSION.First(&w)
	return &w
}
