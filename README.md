# Discord Updater

Discord has [an awful upgrade process](https://x.com/alexellisuk/status/1968230950342652296) on Linux:

1. You open the program and it shows a "Download" button and stops the world. The older version won't even open.
2. You hit download, but have 10 other older .deb files and can't find the right one, so you have to run `ls -ltras ~/Downloads/discord*.deb`
3. You find the right one and have to `dpkg -i` it as `sudo`
4. You have to relaunch Discord

One alternative is to use Flatpak or snap.. I really want to keep that stuff off my system.

This Go program watches `~/Downloads` for new `discord-*.deb` files, debounces them to ensure they're fully written, installs via `dpkg` (with retry on lock), kills existing Discord processes, and relaunches Discord as your user.

## Setup

1. Build the program:
   ```
   go install
   ```

2. Install the systemd service:
   ```
   mkdir -p ~/.config/systemd/user/
   cp discord-updater.service ~/.config/systemd/user/
   ```

3. Enable and start the service:
   ```
   systemctl --user daemon-reload
   systemctl --user enable discord-updater
   systemctl --user start discord-updater
   ```

4. Allow passwordless `sudo` for your current user to use `dpkg -i`:
   ```
   echo "$(whoami) ALL=(ALL) NOPASSWD: /usr/bin/dpkg" | sudo tee -a /etc/sudoers
   ```

Check the logs with:

```bash
journalctl --user -u discord-updater -f
```

## Notes
- The service runs as your user and monitors your Downloads folder.
- If multiple deb files are present, it will install the latest detected one after debounce.
