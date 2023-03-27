# Mainflux Users Groups Things and Channels E2E Testing Tool

A simple utility to create a list of groups and users connected to these groups and channels and things connected to these channels.

## Installation

```bash
cd tools/e2e
make
```

### Usage

```bash
./e2e --help
Tool for testing end to end flow of mainflux by creating groups and users and assigning the together andcreating channels and things and connecting them together.
Complete documentation is available at https://docs.mainflux.io

Usage:
  e2e [flags]

Flags:
  -h, --help                   help for e2e
  -H, --host string            address for mainflux instance (default "localhost")
  -n, --num uint               number of users, groups, channels and things to create and connect (default 10)
  -N, --num_of_messages uint   number of messages to send (default 10)
  -p, --prefix string          name prefix for users, groups, things and channels
```

Example:

```bash
go run tools/e2e/cmd/main.go --host 142.93.118.47
```

If you want to create a list of channels with certificates:

```bash
go run tools/e2e/cmd/main.go --host localhost --num 10 --num_of_messages 100 --prefix e2e
```

Example of output:

```bash
created user with token eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2Nzk5MTE2NjQsImlhdCI6MTY3OTkxMDc2NCwiaWRlbnRpdHkiOiItZHJ5LWJyZWV6ZUBlbWFpbC5jb20iLCJpc3MiOiJjbGllbnRzLmF1dGgiLCJzdWIiOiI3OTRiOWZjNS1jM2MwLTQ4NGQtYWFkZi1hMWU0NjUyMjU5ZmEiLCJ0eXBlIjoiYWNjZXNzIn0.4MG0D_6vBleUAR9sbOOm1VHaIucrbTZYK_KMSkyg6a1uZyzpS6zK-oeD6jmGmQvwR1ALfjtZ0VRZiN0q4eQv_w
created users, groups, things and channels
created policies for users, groups, things and channels
viewed users, groups, things and channels
updated users, groups, things and channels
sent and received messages from channels 
```
