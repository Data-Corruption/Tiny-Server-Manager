package files

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"gorm.io/gorm"
)

var (
	BackupsPath string
	SaveDirPath string
	SavePath    string
	saveIsDir   bool
)

func InitBackupPaths() {
	BackupsPath = "backups"
	CreateDirIfNotExists(BackupsPath)

	// Check if the game save directory exists
	SavePath = Config.GameSavePath
	SaveDirPath = filepath.Dir(SavePath)
	if DirExists(SavePath) {
		saveIsDir = true
	} else if FileExists(SavePath) {
		saveIsDir = false
	} else {
		panic("game save path does not exist")
	}
}

// GetBackupFilePath gets the backup from the database using its ID and returns its file path.
func GetBackupFilePath(ID string) (string, error) {
	var backup Backup

	// Convert the ID string to a uint.
	backupId, err := strconv.ParseUint(ID, 10, 32)
	if err != nil {
		return "", err
	}

	// Query the database for the backup with the given ID.
	result := DB.Where("id = ?", backupId).First(&backup)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return "", nil
		}
		return "", result.Error
	}

	// Return the path of the found backup.
	return backup.Path, nil
}

// helper function that reverses a slice of backups
func reverseBackups(backups []Backup) {
	for i, j := 0, len(backups)-1; i < j; i, j = i+1, j-1 {
		backups[i], backups[j] = backups[j], backups[i]
	}
}

// GetAllBackups gets all backups from the database and returns them.
func GetAllBackups() ([]Backup, error) {
	var backups []Backup

	// Query the database for all backups.
	result := DB.Find(&backups)
	if result.Error != nil {
		return nil, result.Error
	}

	// Reverse the slice of backups so that the newest backup is first.
	reverseBackups(backups)

	return backups, nil
}

// assumes server is stopped
func CreateBackup(comment string) error {
	if !Exists(SavePath) {
		return errors.New("game save location does not exist")
	}

	// create filename for the backup zip file using the current date and time
	outName := time.Now().Format("2006-01-02_15-04-05") + ".zip"
	outPath := filepath.Join(BackupsPath, outName)

	// zip the game save to the backups directory
	if saveIsDir {
		if err := ZipDir(SavePath, outPath); err != nil {
			return err
		}
	} else {
		if err := ZipFile(SavePath, outPath); err != nil {
			return err
		}
	}

	// Create a new Backup instance.
	backup := Backup{
		Path:    outPath,
		Name:    outName,
		Comment: comment,
	}

	// Add the new record to the database.
	result := DB.Create(&backup)

	return result.Error
}

// assumes server is stopped
func RestoreBackup(backupPath string) error {
	// Check if the backup file exists.
	if !Exists(backupPath) {
		return errors.New("backup file does not exist")
	}

	// clean the game save
	err := os.RemoveAll(SavePath)
	if err != nil {
		return err
	}

	// unzip the backup file to the game save directory
	if err := UnZipDir(backupPath, SaveDirPath); err != nil {
		return err
	}

	return nil
}
