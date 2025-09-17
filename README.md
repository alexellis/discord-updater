# Discord Updater

This Go program watches `~/Downloads` for new `discord-*.deb` files, debounces them to ensure they're fully written, installs via `dpkg` (with retry on lock), kills existing Discord processes, and relaunches Discord as the user.

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

4. Allow passwordless sudo for dpkg:
   ```
   echo "$(whoami) ALL=(ALL) NOPASSWD: /usr/bin/dpkg" | sudo tee -a /etc/sudoers
   ```

## Notes
- Ensure `fsnotify` is installed: `go get github.com/fsnotify/fsnotify`
- The service runs as your user and monitors your Downloads folder.
- If multiple deb files are present, it will install the latest detected one after debounce.