# Discord Updater

This Go program automatically keeps Discord up-to-date by monitoring `~/Downloads` for new `discord-*.deb` files and performing automatic hourly update checks. It debounces files to ensure they're fully written, installs via `dpkg` (with retry on lock), kills existing Discord processes, and relaunches Discord as the user.

## Why do we need this tool?

Discord has [an awful upgrade process](https://x.com/alexellisuk/status/1968230950342652296) on Linux:

1. You open the program and it shows a "Download" button and stops the world. The older version won't even open.
2. You hit download, but have 10 other older .deb files and can't find the right one, so you have to run `ls -ltras ~/Downloads/discord*.deb`
3. You find the right one and have to `dpkg -i` it as `sudo`
4. You have to relaunch Discord

One alternative is to use Flatpak or snap.. I really want to keep that stuff off my system. This program automates the entire process for you and it's open source, so you can hack on it as much as you like, or adapt it to update other similar packages.

## Features

- **Automatic Update Checking**: Checks Discord's download URL every hour for new versions
- **Startup Version Display**: Shows current installed Discord version on startup
- **Initial Update Check**: Performs an immediate update check when the program starts
- **Smart Version Comparison**: Only downloads updates if they're newer than both installed version and existing downloads
- **File Monitoring**: Watches `~/Downloads` for manually downloaded .deb files
- **Automatic Installation**: Installs new versions and relaunches Discord automatically
- **Debounce Protection**: Waits for files to be fully written before processing

## How It Works

1. **On Startup**:
   - Displays current installed Discord version
   - Performs immediate update check
   - Starts monitoring `~/Downloads`

2. **Hourly Checks**:
   - Fetches latest version from Discord's download URL (without following redirects)
   - Compares with installed version and latest downloaded .deb
   - Downloads new version if different to `~/Downloads`, which triggers the file detection

3. **File Detection**:
   - Monitors `~/Downloads` for new `discord-*.deb` files
   - Debounces files for 2 seconds to ensure complete downloads
   - Automatically installs and relaunches Discord

## Setup

1. Build the program:
   ```
   go install
   ```

   If you don't have Go, `sudo -E arkade system install go` is a quick way to get it via [arkade](https://arkade.dev).

2. Install the systemd service:

   Replace my username in the template with your own:

   ```
   sed -i "s/alex/$(whoami)/g" discord-updater.service
   mkdir -p ~/.config/systemd/user/
   cp discord-updater.service ~/.config/systemd/user/
   ```

3. Enable and start the service as a user-level service:
   ```
   systemctl --user daemon-reload
   systemctl --user enable discord-updater
   systemctl --user start discord-updater
   ```

If your user does not have passwordless `sudo`, then you'll need it for `dpkg` to work:

```
echo "$(whoami) ALL=(ALL) NOPASSWD: /usr/bin/dpkg" | sudo tee -a /etc/sudoers
```

## Usage

### Manual Run
```bash
./discord-updater
```

### Check Logs

```bash
journalctl --user -u discord-updater -f
```

### Example Output

```
discord-update watching: /home/alex/Downloads
Current installed Discord version: 0.0.123
Performing initial update check...
Checking for Discord updates...
Online version: 0.0.124, Installed: 0.0.123, Latest deb:
New version available: 0.0.124
Downloaded Discord 0.0.124 to /home/alex/Downloads
Installing discord via: discord-0.0.124.deb
Discord updated in 1 attempts
Killing discord and relaunching the new version
```

## Version Detection

The program detects installed Discord versions by scanning `~/.config/discord` directories for version patterns (e.g., `0.0.123`).

## Notes

- The service runs as your user and monitors your Downloads folder
- If multiple deb files are present, it will install the latest detected one after debounce
- Automatic downloads are saved to `~/Downloads` where they're automatically detected and installed
- The program handles dpkg lock conflicts with retry logic
- All operations are logged for monitoring and debugging
