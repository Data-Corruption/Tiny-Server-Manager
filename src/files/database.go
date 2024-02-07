package files

import (
	"time"

	"github.com/Data-Corruption/blog"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ==== Variables =============================================================

var (
	DB              *gorm.DB
	DbPath          string
	closeCleanerReq = make(chan bool)
	closeCleanerAck = make(chan bool)
)

// ==== Model Definitions =====================================================

// RateLimitedIp represents an IP address that has been rate limited.
type RateLimitedIp struct {
	gorm.Model
	Ip  string
	Exp time.Time
}

type Backup struct {
	gorm.Model
	Path    string
	Name    string
	Comment string
}

// Session represents a user session in the system.
type Session struct {
	gorm.Model
	ID  uuid.UUID `gorm:"primaryKey"` // v7 UUID
	IP  string
	Exp time.Time
}

// ==== Model Hooks ===========================================================

func (s *Session) BeforeCreate(tx *gorm.DB) (err error) {
	s.ID, err = uuid.NewV7()
	return err
}

// ==== Database Initialization ===============================================

// RemoveExpiredEntries removes entries that are past their expiration date.
func RemoveExpiredEntries(model interface{}) error {
	// Current time
	now := time.Now()

	// Delete entries where the expiration time is in the past
	result := DB.Where("exp < ?", now).Delete(model)

	return result.Error
}

func startCleaner() {
	ticker := time.NewTicker(time.Minute * 10)

	go func() {
		for {
			select {
			case <-ticker.C:
				if err := RemoveExpiredEntries(&RateLimitedIp{}); err != nil {
					blog.Error(err.Error())
				}
				if err := RemoveExpiredEntries(&Session{}); err != nil {
					blog.Error(err.Error())
				}
			case <-closeCleanerReq:
				ticker.Stop()
				closeCleanerAck <- true
				return
			}
		}
	}()
}

func InitDatabase() {
	DbPath = "tsm.db"

	// Open the database
	db, err := gorm.Open(sqlite.Open(DbPath), &gorm.Config{
		PrepareStmt: true,
	})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schemas
	if err = db.AutoMigrate(&RateLimitedIp{}, &Session{}, &Backup{}); err != nil {
		panic("failed to migrate database")
	}

	// Set the global database variable
	DB = db

	// Start the cleaner
	startCleaner()
}

func CloseDatabase() {
	closeCleanerReq <- true
	<-closeCleanerAck
}
