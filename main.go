package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

func main() {
	home, _ := os.UserHomeDir()
	downloads := filepath.Join(home, "Downloads")

	log.Printf("discord-update watching: %s", downloads)

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
		}
	}
}

func relaunchDiscord() {
	log.Printf("Killing discord and relaunching the new version")

	// Kill Discord processes
	exec.Command("pkill", "-9", "-f", "discord$").Run()

	// Wait a second
	time.Sleep(1 * time.Second)

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
