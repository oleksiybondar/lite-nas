# Before-startup hook

`lite-nas-before-startup.service` runs `/etc/lite-nas/before-startup.sh` once
before LiteNAS services start. The default script exits successfully without
changing the host. The script and service run as root, and the installed script
is root-owned with mode `0700`.

Use the script for host-specific preparation that LiteNAS services depend on,
such as importing and mounting storage, preparing devices, or checking required
host resources. Edit the script on the installed host. Configuration deployment
and package upgrades preserve its contents. Keep it self-contained and return
a nonzero status when preparation fails: systemd will leave dependent LiteNAS
services stopped. Output and errors appear in `journalctl -u
lite-nas-before-startup.service`.

The one-shot unit remains active after a successful run, so restarting an
individual LiteNAS service does not rerun the script. To run it again, use
`sudo systemctl restart lite-nas-before-startup.service` after considering the
effect on services already running. The `before-startup` backup target saves
the script and restores its file mode and ownership. A restore does not execute
the script; the restored version runs at the next boot or explicit restart.
