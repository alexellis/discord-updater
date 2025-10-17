# Discord Updater

This Go program automatically keeps Discord up-to-date with two modes:

**1. Daemon**: Monitors `~/Downloads` for new `discord-*.deb` files and performs automatic hourly update checks.

**2. Discord command wrapper**: Acts as a Discord wrapper that checks for updates on startup then launches Discord (run as an alias named `discordup` or with `--launch` flag). No need for a background service, but a bit more manual to install/consume updates.

## Why do we need this tool?

Discord has [an awful upgrade process](https://x.com/alexellisuk/status/1968230950342652296) on Linux:

1. You open the program and it shows a "Download" button and stops the world. The older version won't even open.
2. You hit download, but have 10 other older .deb files and can't find the right one, so you have to run `ls -ltras ~/Downloads/discord*.deb`
3. You find the right one and have to `dpkg -i` it as `sudo`
4. You have to relaunch Discord

One alternative is to use Flatpak or snap.. I really want to keep that stuff off my system. This program automates the entire process for you and it's open source, so you can hack on it as much as you like, or adapt it to update other similar packages.

## Setup as a daemon (background checking)

1. Get or build the program:

   ```
   # Arkade can be installed without sudo if you are worried.
   # it's only used to move arkade to /usr/local/bin/
   curl -SLs https://get.arkade.dev | sudo sh
   arkade get discord-update
   mkdir -p ~/go/bin/
   cp ~/.arkade/bin/discord-updater ~/go/bin/
   ```

   Or build from source:

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

## Setup for the wrapper mode

1. Build the program:
   ```
   go install
   ```

   If you don't have Go, `sudo -E arkade system install go` is a quick way to get it via [arkade](https://arkade.dev).

2. Option A) create a symlink or alias named `discordup` to the built binary for launcher mode:
   ```
   ln -s $(which discord-updater) ~/bin/discordup
   ```

3. Option B) just use the `--launch` flag when running the binary:

   ```
   discord-updater --launch
   ```

## Check Logs
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

The program detects installed Discord versions by reading `/usr/share/discord/resources/build_info.json`, which contains the exact version information in JSON format. This method is more reliable and accurate than scanning directory names.

# License

[MIT](LICENSE.md] - no warranty of any kind. Use and adapt at your own risk.

