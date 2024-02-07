package routes

import (
	"fmt"
	"net/http"
	"path/filepath"

	"tsm/src/files"
	"tsm/src/game"

	"github.com/Data-Corruption/blog"
	"github.com/go-chi/chi/v5"
)

type DashboardPageData struct {
	Title   string
	Backups []files.Backup
}

func RegisterDashboardRoutes(r *chi.Mux) {
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		// get the list of backups
		backups, err := files.GetAllBackups()
		if err != nil {
			blog.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		blog.Debug(fmt.Sprintf("Backups: %v", backups))

		pageData := DashboardPageData{
			Title:   files.Config.DashboardTitle,
			Backups: backups,
		}

		// get the dashboard template path
		dashboardTemplatePath := filepath.Join("public", "dashboard.html")
		dashboardTemplate, err := LoadTemplate(dashboardTemplatePath)
		if err != nil {
			blog.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// execute the dashboard template
		if err := dashboardTemplate.Execute(w, pageData); err != nil {
			blog.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	r.Post("/restart", func(w http.ResponseWriter, r *http.Request) {
		blog.Debug("Start of restart handler")

		// lock the game mutex
		game.Process.Mutex.Lock()
		defer game.Process.Mutex.Unlock()
		blog.Debug("Locked game mutex")

		// stop the server
		if err := game.Process.Stop(); err != nil {
			blog.Error(fmt.Sprintf("Failed to stop game server: %s", err.Error()))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		blog.Debug("Stopped game server")

		// start the server again
		if err := game.Process.Start(); err != nil {
			blog.Error(fmt.Sprintf("Failed to start game server: %s", err.Error()))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		blog.Debug("Started game server")
	})

	r.Post("/backup", func(w http.ResponseWriter, r *http.Request) {
		// Parse the multipart form data
		if err := r.ParseMultipartForm(32 << 20); err != nil { // 32MB is the default used by http.MaxBytesReader
			http.Error(w, "Failed to parse multipart form", http.StatusBadRequest)
			return
		}

		// Retrieve the comment
		comment := r.FormValue("comment")
		blog.Debug(fmt.Sprintf("Comment: %s", comment))

		// lock the game mutex
		game.Process.Mutex.Lock()
		defer game.Process.Mutex.Unlock()

		// stop the server
		if err := game.Process.Stop(); err != nil {
			blog.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// create the backup
		if err := files.CreateBackup(comment); err != nil {
			blog.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// start the server again
		if err := game.Process.Start(); err != nil {
			blog.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	r.Post("/update", func(w http.ResponseWriter, r *http.Request) {
		blog.Debug("Start of update handler")

		// lock the game mutex
		game.Process.Mutex.Lock()
		defer game.Process.Mutex.Unlock()
		blog.Debug("Locked game mutex")

		// stop the server
		if err := game.Process.Stop(); err != nil {
			blog.Error(fmt.Sprintf("Failed to stop game server: %s", err.Error()))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		blog.Debug("Stopped game server")

		// update the server
		var err error
		if err = game.Update(); err != nil {
			blog.Error(fmt.Sprintf("Failed to update game server: %s", err.Error()))
		}

		// start the server again
		if err := game.Process.Start(); err != nil {
			blog.Error(fmt.Sprintf("Failed to start game server: %s", err.Error()))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		blog.Debug("Started game server")

		// if there was an error updating the server, return it
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	r.Get("/download", func(w http.ResponseWriter, r *http.Request) {
		// get the id of the backup to download
		backupId := r.URL.Query().Get("backupId")
		blog.Debug(fmt.Sprintf("Backup ID: %s", backupId))
		if backupId == "" {
			blog.Error("No backup ID provided")
			http.Error(w, "No backup ID provided", http.StatusBadRequest)
			return
		}

		// get the file path of the backup
		filePath, err := files.GetBackupFilePath(backupId)
		if err != nil {
			blog.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// send the file to the client
		if err := files.SendFileToClient(w, filePath); err != nil {
			blog.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		blog.Info("Successfully sent file to client")
	})

	r.Post("/restore", func(w http.ResponseWriter, r *http.Request) {
		// get the id of the backup to restore
		backupId := r.URL.Query().Get("backupId")
		if backupId == "" {
			blog.Error("No backup ID provided")
			http.Error(w, "No backup ID provided", http.StatusBadRequest)
			return
		}

		// get the file path of the backup
		filePath, err := files.GetBackupFilePath(backupId)
		if err != nil {
			blog.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// lock the game mutex
		game.Process.Mutex.Lock()
		defer game.Process.Mutex.Unlock()

		// stop the server
		if err := game.Process.Stop(); err != nil {
			blog.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// restore the backup
		if err := files.RestoreBackup(filePath); err != nil {
			blog.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// start the server again
		if err := game.Process.Start(); err != nil {
			blog.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}
