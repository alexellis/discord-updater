package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func runLauncherMode() {
	log.SetFlags(0) // Simpler logging for launcher mode

	home, _ := os.UserHomeDir()
	downloads := filepath.Join(home, "Downloads")

	log.Printf("Discord Updater - Launcher Mode")

	// Print installed version
	installedVersion := getInstalledVersion()
	if installedVersion != "" {
		log.Printf("Current version: %s", installedVersion)
	} else {
		log.Printf("Could not determine installed Discord version")
	}

	// Quick update check
	log.Printf("Checking for updates...")
	onlineVersion, downloadUrl := getLatestOnlineVersion()

	if onlineVersion == "" {
		log.Printf("Failed to check for updates")
		// If Discord is already running, leave it be
		if isDiscordRunning() {
			log.Printf("Discord is already running, leaving it as-is")
			return
		}
		log.Printf("Launching Discord...")
		launchDiscord()
		return
	}

	log.Printf("Latest online version: %s", onlineVersion)

	needsUpdate := onlineVersion != installedVersion

	if needsUpdate {
		log.Printf("Update available! Downloading %s...", onlineVersion)
		err := downloadDeb(downloadUrl, downloads)
		if err != nil {
			log.Printf("Failed to download update: %v", err)
			// If Discord is already running, leave it be
			if isDiscordRunning() {
				log.Printf("Discord is already running, leaving it as-is")
				return
			}
			log.Printf("Launching current version...")
			launchDiscord()
			return
		}

		log.Printf("Downloaded update, installing...")
		// Wait a moment for file to be fully written
		time.Sleep(2 * time.Second)

		// Install the downloaded deb
		filename := filepath.Base(downloadUrl)
		debPath := filepath.Join(downloads, filename)
		install(debPath)

		log.Printf("Installation complete, launching Discord...")
		launchDiscord()
	} else {
		log.Printf("Already up to date")
		// If Discord is already running, leave it be
		if isDiscordRunning() {
			log.Printf("Discord is already running, no action needed")
			return
		}
		log.Printf("Launching Discord...")
		launchDiscord()
	}
}

func isDiscordRunning() bool {
	cmd := exec.Command("pgrep", "-f", "discord$")
	err := cmd.Run()
	return err == nil
}

func launchDiscord() {
	// Kill any existing Discord processes
	exec.Command("pkill", "-9", "-f", "discord$").Run()
	time.Sleep(1 * time.Second)

	// Launch Discord as detached process
	cmd := exec.Command("discord")
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil

	err := cmd.Start()
	if err != nil {
		log.Printf("Failed to launch Discord: %v", err)
		return
	}

	log.Printf("Discord launched (PID: %d)", cmd.Process.Pid)

	// Detach from the process
	cmd.Process.Release()
}