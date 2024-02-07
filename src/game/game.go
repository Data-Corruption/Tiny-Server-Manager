package game

import (
	"errors"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"

	"tsm/src/files"

	"github.com/Data-Corruption/blog"
)

// Process is the game server process manager
var Process *ProcessManager

// Starts the server
func InitGameServer() {
	if files.Config.GameExePath == "" {
		panic("game exe path not set")
	}

	// check if the exe path is valid
	if !files.FileExists(files.Config.GameExePath) {
		panic("invalid game exe path")
	}

	// init the process manager
	Process = NewProcessManager(files.Config.GameExePath)

	// start the server
	if err := Process.Start(); err != nil {
		panic(err)
	}
}

// Updates the server (runs the UpdateCommand from the config) assumes server is stopped and Mutex is locked
func Update() error {
	if files.Config.UpdateCommand == "" {
		return errors.New("update command not set")
	}

	// Tokenize the update command
	args := strings.Fields(files.Config.UpdateCommand)
	if len(args) == 0 {
		return errors.New("update command is empty after tokenizing")
	}

	// create and run the update command
	cmd := exec.Command(args[0], args[1:]...)
	if err := cmd.Run(); err != nil {
		return err
	}

	// sleep for 1 second (probably not necessary)
	time.Sleep(1 * time.Second)

	return nil
}

type ProcessManager struct {
	// used by routes and auto backup
	Mutex        sync.Mutex
	running      bool   // Is the process running
	status       string // Verbose status of the process
	runningMutex sync.Mutex
	statusMutex  sync.Mutex
	command      string    // command to run
	cmd          *exec.Cmd // command object
	// channels for communication with the run goroutine and it's child
	stopChan chan struct{}
	doneChan chan error
}

func NewProcessManager(command string) *ProcessManager {
	return &ProcessManager{
		running:  false,
		status:   "Hasn't started yet",
		command:  command,
		stopChan: make(chan struct{}),
		doneChan: make(chan error),
	}
}

// ==== Getters and setters ===================================================

func (pm *ProcessManager) GetRunning() bool {
	pm.runningMutex.Lock()
	defer pm.runningMutex.Unlock()
	return pm.running
}

func (pm *ProcessManager) SetRunning(running bool) {
	pm.runningMutex.Lock()
	defer pm.runningMutex.Unlock()
	pm.running = running
}

func (pm *ProcessManager) GetStatus() string {
	pm.statusMutex.Lock()
	defer pm.statusMutex.Unlock()
	return pm.status
}

func (pm *ProcessManager) SetStatus(status string) {
	pm.statusMutex.Lock()
	defer pm.statusMutex.Unlock()
	pm.status = status
}

// ==== Process management ====================================================

func (pm *ProcessManager) Start() error {
	if pm.GetRunning() {
		blog.Error("Tried to start game server when it was already running")
		return errors.New("process already running")
	}

	pm.cmd = exec.Command(pm.command)
	pm.cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	blog.Debug("Created command")

	go pm.runProcess()
	blog.Debug("Started run goroutine")

	return nil
}

func (pm *ProcessManager) Stop() error {
	blog.Debug("Stopping game server")

	if !pm.GetRunning() {
		blog.Error("Tried to stop game server when it was not running")
		return errors.New("process not running")
	}

	pm.stopChan <- struct{}{}
	blog.Debug("Sent stop signal")
	err := <-pm.doneChan
	blog.Debug("Received done signal")

	// err should now be 'signal: terminated' or nil, else return error
	if err != nil && err.Error() != "signal: terminated" {
		return err
	} else {
		return nil
	}
}

func (pm *ProcessManager) runProcess() {
	if err := pm.cmd.Start(); err != nil {
		pm.doneChan <- err
		pm.SetRunning(false)
		return
	}

	pm.SetRunning(true)

	go func() {
		<-pm.stopChan
		blog.Debug("Received stop signal from channel")
		if pm.cmd.Process != nil {
			blog.Debug("Sending SIGTERM to child process")
			pgid, err := syscall.Getpgid(pm.cmd.Process.Pid)
			if err != nil {
				blog.Error(err.Error())
				return
			}
			if err := syscall.Kill(-pgid, syscall.SIGTERM); err != nil {
				blog.Error(err.Error())
			}
		}
	}()

	err := pm.cmd.Wait()
	// add extra buffer of 3 seconds to allow for graceful shutdown
	time.Sleep(3 * time.Second)
	blog.Debug("Child process exited")
	pm.SetRunning(false)
	pm.doneChan <- err
	blog.Debug("Sent done signal")
}
