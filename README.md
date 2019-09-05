# Sonos Volume Logger

Is someone in your office playing silly buggers with the volume on the Sonos?

This stupid tool will connect to a Sonos device using the uPnP API and log the volume level at specified intervals so you can gather proof of Sonos Shenanigans™

## Origin Story
![Slack conversation](https://i.imgur.com/bQtytYt.png)

## Installation
```bash
go install github.com/pthm/sonosvollog
sonosvollog
```

## Usage
```
Searching for Sonos devices...
Found 1 Sonos devices:
	1. Sesh Pit
Choose: 1
Logging volume for Sesh Pit
Vol @ 2019-09-05 20:48:09: 15
Vol @ 2019-09-05 20:48:19: 18
Vol @ 2019-09-05 20:48:29: 21
```