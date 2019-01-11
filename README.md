# What is Hudor?

Hudor is a telegram bot with only one purpose and that is hold the door against spammer bots.

## Requirements

- [Go](https://golang.org) `v1.11` or later.
- [Redis](https://redis.io/) database.

## Installing

Using Hudor bot is easy, first use `go get` to install latest version of hudor.

```bash
go get -u github.com/amir-yaghoobi/hudorBot/hudor
```

## Configurations

In order to hudor can operate you must provide a configuration file in one of following paths:

- `/etc/hudor/`
- `$HOME/.hudor/`
- current directory

File name must be `config.xxx` (`json`, `yml`, `toml`, ...)

### Example

```toml
# config.toml

# Obtain it from @botFather
# Read more: https://core.telegram.org/bots#3-how-do-i-create-a-bot
telegramToken = "BOT_TOKEN"

[redis]
  db = 0
  port = 6379
  hostname = "localhost"
  password = ""

# you can either pass nanoseconds
# or formatted string like "h|m|s|ms|us|ns"
# example:
#     - 21600000000000 == "6h"
#     - 604800000000000 == "168h" (1week)
# 
# -------------------------------------------------------------------------
# warn:  after an usual user adding a non whitelisted bot to they group
#        they will get a warning and default expiry is set to 7 days
# -------------------------------------------------------------------------
# state: user state (pv state) and defualt expiry is set to 6 hours
# -------------------------------------------------------------------------

[expiry]
  warn  = "168h"
  state = "6h"
```
