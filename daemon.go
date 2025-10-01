package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

func runDaemonMode() {
	home, _ := os.UserHomeDir()
	downloads := filepath.Join(home, "Downloads")

	log.Printf("discord-update watching: %s", downloads)

	// Print installed version on startup
	installedVersion := getInstalledVersion()
	if installedVersion != "" {
		log.Printf("Current installed Discord version: %s", installedVersion)
	} else {
		log.Printf("Could not determine installed Discord version")
	}

	// Check for updates on startup
	log.Printf("Performing initial update check...")
	downloadedFile := checkForUpdates(downloads)

	// Install any file that was downloaded during initial check
	if downloadedFile != "" {
		fullPath := filepath.Join(downloads, downloadedFile)
		log.Printf("Installing discord via: %s", downloadedFile)
		install(fullPath)
		relaunchDiscord()
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	defer watcher.Close()

	err = watcher.Add(downloads)
	if err != nil {
		panic(err)
	}

	debounce := make(map[string]time.Time)
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case event := <-watcher.Events:
			if event.Op&(fsnotify.Create|fsnotify.Write) != 0 {
				base := filepath.Base(event.Name)
				if strings.HasPrefix(base, "discord-") && strings.HasSuffix(base, ".deb") {
					debounce[event.Name] = time.Now()
				}
			}
		case <-time.After(1 * time.Second):
			for file, t := range debounce {
				if time.Since(t) > 2*time.Second {
					log.Printf("Installing discord via: %s", filepath.Base(file))

					delete(debounce, file)
					install(file)
					relaunchDiscord()
				}
			}
		case <-ticker.C:
			checkForUpdates(downloads)
		}
	}
}

func relaunchDiscord() {
	log.Printf("Killing discord and relaunching the new version")

	// Kill Discord processes
	exec.Command("pkill", "-9", "-f", "discord$").Run()

	// Wait a second
	time.Sleep(1 * time.Second)

	log.Printf("Relaunching Discord...")
	// Relaunch Discord
	exec.Command("discord").Start()
}
