# IP Sender
A small bot to send your IP address via a Telegram bot
## What is it used for?
Think you have dynamic IP and you are running a VNC Server on your computer(Or SSH server or remote desktop). You need your IP every time you reboot the router. This application will send your IP every time you request it.
## Features
* Ultra Lightweight: Just in 150 lines of code you can run this bot. Also memory and CPU consumption is very low.
* Password Protected: Your password is hashed in SHA-256 and stored.
## Setup
Download the binary from [releases](https://github.com/HirbodBehnam/IP-Sender-Go/releases) and the [config.json](https://raw.githubusercontent.com/HirbodBehnam/IP-Sender-Go/master/config.json). Edit the config json and enter the Token and a SHA-256 hashed password required or receiving the IP. For example here `ca978112ca1bbdcafac231b39a23dc4da786eff8147c4e72b9807785afee48bb` is letter `a`.

You can use [this](https://emn178.github.io/online-tools/sha256.html) or any other tool to generate the SHA-256 password.

After setting up the bot and starting it, just send the bot the password you defined (For example in above example you should send the bot `a`) in order to get the IP
### Build
Clone this repository and:
```bash
go get github.com/go-telegram-bot-api/telegram-bot-api
go build main.go
```
### Proxy
You can define a proxy in order to access telegram server via it. The bot supports socks5, http and https.

In config.json, just like the example, define the proxy host and the type of it.