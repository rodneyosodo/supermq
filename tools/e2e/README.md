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
Tool for testing end to end flow ow mainflux by creating groups and users and assigning the together andcreating channels and things and connecting them together.
Complete documentation is available at https://docs.mainflux.io

Usage:
  e2e [flags]

Flags:
  -h, --help              help for e2e
      --host string       address for mainflux instance (default "https://localhost")
      --num int           number of users, groups, channels and things to create and connect (default 10)
  -p, --password string   mainflux users password
      --prefix string     name prefix for users, groups, things and channels
  -u, --username string   mainflux user
```

Example:

```bash
go run tools/e2e/cmd/main.go -u test@mainflux.com -p test1234 --host https://142.93.118.47
```

If you want to create a list of channels with certificates:

```bash
go run tools/provision/cmd/main.go  --host http://localhost --num 10 -u test@mainflux.com -p test1234

```

Example of output:

```bash
Created admin with token eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2NzgxOTc5NDAsImlhdCI6MTY3ODE5NzA0MCwiaWRlbnRpdHkiOiJwb2xpc2hlZC1sYWtlQGVtYWlsLmNvbSIsImlzcyI6ImNsaWVudHMuYXV0aCIsInN1YiI6IjE5NmU4N2M5LTllMjItNDRmNC1hZmY0LTM0OWM5YzcwMGFlNiIsInR5cGUiOiJhY2Nlc3MifQ.61miO5nKNhhivntR99DVIab_sPMnm8IgZ9pkrPIUkxvGN1pe80DiI0k148Lf7Ty-4KFUsd4i0Ikv5Dd0qVpmuQ
Created users, groups, things and channels
Created policies for users, groups, things and channels
Viewed users, groups, things and channels
Updated users, groups, things and channels
```