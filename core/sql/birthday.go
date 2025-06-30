package sql

type Birthday struct {
	UserID  string `gorm:"primaryKey"` 
	Date    string
	ChatJID string
func SaveBirthday(userID, date, chatJID string) {
	birthday := &Birthday{UserID: userID}
	tx := SESSION.Begin()
	tx.FirstOrCreate(birthday)
	birthday.Date = date
	birthday.ChatJID = chatJID
	tx.Save(birthday)
	tx.Commit()
}
func DeleteBirthday(userID string) bool {
	birthday := &Birthday{UserID: userID}
	return SESSION.Delete(birthday).RowsAffected != 0
}
func GetTodaysBirthdays(todayDate string) []Birthday {
	var birthdays []Birthday
	SESSION.Where("date = ?", todayDate).Find(&birthdays)
	return birthdays
}
func GetAllBirthdays() []Birthday {
	var birthdays []Birthday
	SESSION.Find(&birthdays)
	return birthdays
}
