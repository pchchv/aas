package sqlitedb

import (
	"database/sql"

	"github.com/pchchv/aas/src/models"
)

func (d *SQLiteDB) CreateUser(tx *sql.Tx, user *models.User) error {
	return d.CommonDB.CreateUser(tx, user)
}

func (d *SQLiteDB) UpdateUser(tx *sql.Tx, user *models.User) error {
	return d.CommonDB.UpdateUser(tx, user)
}

func (d *SQLiteDB) GetUsersByIds(tx *sql.Tx, userIds []int64) (map[int64]models.User, error) {
	return d.CommonDB.GetUsersByIds(tx, userIds)
}

func (d *SQLiteDB) GetUserById(tx *sql.Tx, userId int64) (*models.User, error) {
	return d.CommonDB.GetUserById(tx, userId)
}

func (d *SQLiteDB) GetUserByUsername(tx *sql.Tx, username string) (*models.User, error) {
	return d.CommonDB.GetUserByUsername(tx, username)
}

func (d *SQLiteDB) GetUserBySubject(tx *sql.Tx, subject string) (*models.User, error) {
	return d.CommonDB.GetUserBySubject(tx, subject)
}

func (d *SQLiteDB) GetUserByEmail(tx *sql.Tx, email string) (*models.User, error) {
	return d.CommonDB.GetUserByEmail(tx, email)
}

func (d *SQLiteDB) GetLastUserWithOTPState(tx *sql.Tx, otpEnabledState bool) (*models.User, error) {
	return d.CommonDB.GetLastUserWithOTPState(tx, otpEnabledState)
}

func (d *SQLiteDB) SearchUsersPaginated(tx *sql.Tx, query string, page int, pageSize int) ([]models.User, int, error) {
	return d.CommonDB.SearchUsersPaginated(tx, query, page, pageSize)
}

func (d *SQLiteDB) DeleteUser(tx *sql.Tx, userId int64) error {
	return d.CommonDB.DeleteUser(tx, userId)
}

func (d *SQLiteDB) UsersLoadPermissions(tx *sql.Tx, users []models.User) error {
	return d.CommonDB.UsersLoadPermissions(tx, users)
}

func (d *SQLiteDB) UserLoadAttributes(tx *sql.Tx, user *models.User) error {
	return d.CommonDB.UserLoadAttributes(tx, user)
}

func (d *SQLiteDB) UserLoadPermissions(tx *sql.Tx, user *models.User) error {
	return d.CommonDB.UserLoadPermissions(tx, user)
}

func (d *SQLiteDB) UsersLoadGroups(tx *sql.Tx, users []models.User) error {
	return d.CommonDB.UsersLoadGroups(tx, users)
}

func (d *SQLiteDB) UserLoadGroups(tx *sql.Tx, user *models.User) error {
	return d.CommonDB.UserLoadGroups(tx, user)
}
