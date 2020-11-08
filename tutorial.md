
## Required Material
* Raspberry Pi
* 8GB or 16GB MicroSD and Power Supply

## Setup Github
1. Create a Github user which has read access to all your org repositories.
2. Go into [Personal access tokens](https://github.com/settings/tokens) and create one for the Github Mirror.

## Setup Raspberry Pi
1. Download [RASPBIAN STRETCH LITE](https://www.raspberrypi.org/downloads/raspbian/)
2. Copy onto your SD card, with [Etcher](https://etcher.io) or manually:
3.
```
diskutil list
diskutil unmountDisk /dev/diskX
sudo dd bs=4m if=IMG of=/dev/rdiskX
```
3. Enable SSH by creating a file called SSH on boot volume. `touch /Volumes/boot/SSH`
4. Put SD card into Raspberry Pi. Connect Raspberry Pi to network and power.

>We never use Wi-Fi for serious applications.

5. Login into pi, password is `raspberry`.
`ssh pi@raspberrypi.local`

6. Change password by executing `passwd`. [Setup password less SSH access](https://www.raspberrypi.org/documentation/remote-access/ssh/passwordless.md)

7. Change hostname to something that makes sense, `sudo raspi-config`
8. Configure timezone: `sudo raspi-config`

9. Configure static IP. `nano /etc/dhcpcd.conf`
And add at the bottom:

```
interface eth0
static ip_address=192.168.1.10/24
static routers=192.168.1.1
static domain_name_servers=192.168.1.1 8.8.8.8 4.2.2.1
```
Restart with `sudo reboot`

> Alternative you can assign a fix IP in DHCP.

10. Optional update raspberry pi

```
sudo apt-get update
sudo apt-get dist-upgrade
sudo apt-get clean
```
11. Install Git: `sudo apt-get install git -y`
12. Download Binary...

13. Setup systemd, a service for the timer:
`sudo nano /etc/systemd/system/githubmirror.service`
(check that your binary path in ExecStart is correct)

> It does not have to be enabled, it will be run from the timer


```
[Unit]
Description=Github Mirror
After=network.target

[Service]
User=pi
ExecStart=/home/pi/github-mirror -user=... -secret=... -github-path=/orgs/...

[Install]
WantedBy=multi-user.target
```

14. Setup systemd, timer for the service:
`sudo nano /etc/systemd/system/githubmirror.timer`

```
[Unit]
Description=Run githubmirror

[Timer]
OnCalendar=*-*-* 09:00:00

[Install]
WantedBy=timers.target
```

15. `sudo systemctl daemon-reload`
16. `sudo systemctl enable githubmirror.timer`
17. `sudo systemctl start githubmirror.timer`


## Systemd commands
* show timers: `systemctl list-timers`
* Status: `systemctl status githubmirror`
* start: `sudo systemctl start githubmirror`
* show log: `journalctl -f -u githubmirror`


## Testing
- [X] turn raspberry pi off and on again, backup should run next time
- [X] cd into a repo, git log to check if changes are there
