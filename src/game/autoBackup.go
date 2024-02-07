package game

import (
	"time"
	"tsm/src/files"

	"github.com/Data-Corruption/blog"
)

var (
	AutoBackupStopChan = make(chan struct{})
	AutoBackupDoneChan = make(chan struct{})
)

func InitAutoBackup() {
	go autoBackupGoroutine()
}

func StopAutoBackup() {
	AutoBackupStopChan <- struct{}{}
	<-AutoBackupDoneChan
}

func backup() error {
	Process.Mutex.Lock()
	defer Process.Mutex.Unlock()

	if err := Process.Stop(); err != nil {
		return err
	}
	if err := files.CreateBackup("Automatic"); err != nil {
		return err
	}
	if err := Process.Start(); err != nil {
		return err
	}
	return nil
}

func timeUntilMidnight() time.Duration {
	now := time.Now()
	nextMidnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	return nextMidnight.Sub(now)
}

func autoBackupGoroutine() {
	// Calculate initial delay until next midnight
	delay := timeUntilMidnight()
	ticker := time.NewTicker(delay)

	for {
		select {
		case <-ticker.C:
			if err := backup(); err != nil {
				blog.Error(err.Error())
				panic(err)
			}
			blog.Info("Performed automatic backup")
			ticker.Reset(timeUntilMidnight())
		case <-AutoBackupStopChan:
			ticker.Stop()
			AutoBackupDoneChan <- struct{}{}
			return
		}
	}
}
