package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

type DiscordBuildInfo struct {
	ReleaseChannel string `json:"releaseChannel"`
	Version        string `json:"version"`
}

func main() {
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
	checkForUpdates(downloads)

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

func install(file string) {
	attempts := 10
	interval := time.Second * 5

	taken := 0
	for i := 0; i < attempts; i++ {
		cmd := exec.Command("sudo", "dpkg", "-i", file)
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		err := cmd.Run()
		if err == nil {
			taken++
			break
		}

		if strings.Contains(fmt.Sprintf("%v", err), "lock") {
			time.Sleep(interval)
			continue
		}
		fmt.Println("Install failed:", err)

		return
	}

	if taken > 1 {
		log.Printf("Discord updated in %d attempts", taken)
	}
}

func getInstalledVersion() string {
	buildInfoPath := "/usr/share/discord/resources/build_info.json"

	data, err := os.ReadFile(buildInfoPath)
	if err != nil {
		log.Printf("Failed to read Discord build info: %v", err)
		return ""
	}

	var buildInfo DiscordBuildInfo
	err = json.Unmarshal(data, &buildInfo)
	if err != nil {
		log.Printf("Failed to parse Discord build info JSON: %v", err)
		return ""
	}

	return buildInfo.Version
}

func getLatestDebVersion(downloads string) string {
	files, err := os.ReadDir(downloads)
	if err != nil {
		log.Printf("Failed to read downloads dir: %v", err)
		return ""
	}

	re := regexp.MustCompile(`discord-(\d+\.\d+\.\d+)\.deb`)
	var latestVersion string

	for _, file := range files {
		if strings.HasPrefix(file.Name(), "discord-") && strings.HasSuffix(file.Name(), ".deb") {
			matches := re.FindStringSubmatch(file.Name())
			if len(matches) > 1 {
				version := matches[1]
				if version > latestVersion {
					latestVersion = version
				}
			}
		}
	}
	return latestVersion
}

func getLatestOnlineVersion() (string, string) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Get("https://discord.com/api/download?platform=linux")
	if err != nil {
		log.Printf("Failed to fetch Discord download URL: %v", err)
		return "", ""
	}
	defer resp.Body.Close()

	if resp.StatusCode == 302 {
		location := resp.Header.Get("Location")
		if location != "" {
			u, err := url.Parse(location)
			if err != nil {
				log.Printf("Failed to parse location URL: %v", err)
				return "", ""
			}
			re := regexp.MustCompile(`discord-(\d+\.\d+\.\d+)\.deb`)
			matches := re.FindStringSubmatch(u.Path)
			if len(matches) > 1 {
				return matches[1], location
			}
		}
	}
	return "", ""
}

func downloadDeb(downloadUrl, downloads string) error {
	resp, err := http.Get(downloadUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	u, err := url.Parse(downloadUrl)
	if err != nil {
		return err
	}
	filename := filepath.Base(u.Path)
	destPath := filepath.Join(downloads, filename)

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func checkForUpdates(downloads string) {
	log.Printf("Checking for Discord updates...")

	onlineVersion, downloadUrl := getLatestOnlineVersion()
	if onlineVersion == "" {
		log.Printf("Failed to get online version")
		return
	}

	installedVersion := getInstalledVersion()
	latestDebVersion := getLatestDebVersion(downloads)

	log.Printf("Online version: %s, Installed: %s, Latest deb: %s", onlineVersion, installedVersion, latestDebVersion)

	if onlineVersion != installedVersion && onlineVersion != latestDebVersion {
		log.Printf("New version available: %s", onlineVersion)
		err := downloadDeb(downloadUrl, downloads)
		if err != nil {
			log.Printf("Failed to download update: %v", err)
		} else {
			log.Printf("Downloaded Discord %s to %s", onlineVersion, downloads)
		}
	}
}
