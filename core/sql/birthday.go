package sql

import (
	"log"
)
type Birthday struct {
	JID      string `gorm:"primary_key"`
	Birthday string
}
func AddBirthday(jid, birthday string) {
	b := &Birthday{JID: jid}
	tx := SESSION.Begin()
	tx.FirstOrCreate(b)
	b.Birthday = birthday
	if err := tx.Save(b).Error; err != nil {
		log.Printf("Failed to save birthday: %v", err)
		tx.Rollback() 
		return
	}
	tx.Commit()
}
func GetBirthdaysForDate(date string) []Birthday {
	var birthdays []Birthday
	SESSION.Where("birthday = ?", date).Find(&birthdays)
	return birthdays
}
// assuming gorm's AutoMigrate should handle table creation
