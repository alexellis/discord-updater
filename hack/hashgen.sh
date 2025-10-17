#!/bin/sh

for f in bin/discord-updater*; do shasum -a 256 $f > $f.sha256; done