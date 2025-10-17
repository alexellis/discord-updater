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

	"github.com/alexellis/discord-updater/pkg"
)

type DiscordBuildInfo struct {
	ReleaseChannel string `json:"releaseChannel"`
	Version        string `json:"version"`
}

func main() {
	// Check if running in launcher mode (discordup)
	isLauncherMode := strings.Contains(os.Args[0], "discordup") || len(os.Args) > 1 && os.Args[1] == "--launch"

	if isLauncherMode {
		runLauncherMode()
		return
	}

	// Daemon mode (original functionality)
	runDaemonMode()
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

	req, err := http.NewRequest("GET", "https://discord.com/api/download?platform=linux", nil)
	if err != nil {
		log.Printf("Failed to create request: %v", err)
		return "", ""
	}
	req.Header.Set("User-Agent", pkg.UserAgent())

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to fetch Discord download URL: %v", err)
		return "", ""
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}

	if resp.StatusCode == http.StatusFound {
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
	req, err := http.NewRequest("GET", downloadUrl, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", pkg.UserAgent())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("bad status: %s, body: %s", resp.Status, string(body))
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

func checkForUpdates(downloads string) string {
	log.Printf("Checking for Discord updates...")

	onlineVersion, downloadUrl := getLatestOnlineVersion()
	if onlineVersion == "" {
		log.Printf("Failed to get online version")
		return ""
	}

	installedVersion := getInstalledVersion()
	latestDebVersion := getLatestDebVersion(downloads)

	log.Printf("Online version: %s, Installed: %s, Latest deb: %s", onlineVersion, installedVersion, latestDebVersion)

	if onlineVersion != installedVersion {
		log.Printf("New version available: %s", onlineVersion)

		// Check if we already have the correct .deb file
		if onlineVersion == latestDebVersion {
			// We have the file, return its name
			u, err := url.Parse(downloadUrl)
			if err == nil {
				return filepath.Base(u.Path)
			}
		} else {
			// Download the file
			err := downloadDeb(downloadUrl, downloads)
			if err != nil {
				log.Printf("Failed to download update: %v", err)
			} else {
				log.Printf("Downloaded Discord %s to %s", onlineVersion, downloads)
				// Return the filename of the downloaded file
				u, err := url.Parse(downloadUrl)
				if err == nil {
					return filepath.Base(u.Path)
				}
			}
		}
	}
	return ""
}
