package database

type DatabaseSeeder struct {
	DB Database
}

func NewDatabaseSeeder(database Database) *DatabaseSeeder {
	return &DatabaseSeeder{
		DB: database,
	}
}
