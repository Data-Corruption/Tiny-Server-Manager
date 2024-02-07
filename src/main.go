package main

import (
	"tsm/src/files"
	"tsm/src/game"
	"tsm/src/server"

	"github.com/Data-Corruption/blog"
)

func initLogger() {
	// make sure log directory exists
	logPath := "logs"
	if _, err := files.CreateDirIfNotExists(logPath); err != nil {
		panic(err)
	}

	// convert log level string from config to enum
	logLevel, ok := blog.LogLevelFromString(files.Config.LogLevel)
	if !ok {
		panic("invalid log level in config file")
	}

	// init the logger
	if err := blog.Init(logPath, logLevel); err != nil {
		blog.Error("Issue initializing blog, falling back to console logging")
	}

	// if debug mode, also log to console
	if logLevel == blog.DEBUG {
		blog.SetUseConsole(true)
	}
}

func startup() {
	files.InitDatabase()
	files.LoadConfig()
	initLogger()
	files.InitBackupPaths()
	game.InitGameServer()
	game.InitAutoBackup()
}

func cleanup() {
	game.StopAutoBackup()
	game.Process.Stop()
	files.CloseDatabase()
	blog.SyncFlush(0)
}

func main() {
	defer cleanup()
	startup()
	server.Instance.Start()
}
