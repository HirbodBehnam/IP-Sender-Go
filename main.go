package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

type config struct {
	Token string
	Pass  string
	Proxy proxyConfig
}
type proxyConfig struct {
	Type string
	Host string
}

var Verbose = false

const Version = "1.0.0 / Build 1"

func main() {
	var ConfigFileName string
	{ //Parse arguments
		configFileName := flag.String("config", "config.json", "The config filename")
		verbose := flag.Bool("v", false, "Verbose mode")
		help := flag.Bool("h", false, "Show help")
		flag.Parse()

		Verbose = *verbose
		if Verbose {
			fmt.Println("Verbose mode on")
		}
		ConfigFileName = *configFileName

		if *help {
			fmt.Println("Created by Hirbod Behnam")
			fmt.Println("Source at https://github.com/HirbodBehnam/IP-Sender-Go")
			fmt.Println("Version", Version)
			flag.PrintDefaults()
			os.Exit(0)
		}
	}

	//Parse config file
	var Config config
	{
		confF, err := ioutil.ReadFile(ConfigFileName)
		if err != nil {
			panic("Cannot read the config file. (io Error) " + err.Error())
		}

		err = json.Unmarshal(confF, &Config)
		if err != nil {
			panic("Cannot read the config file. (Parse Error) " + err.Error())
		}

		Config.Pass = strings.ToLower(Config.Pass)
	}

	//Set proxy if needed
	if Config.Proxy.Type != "" {
		err := os.Setenv("HTTP_PROXY", Config.Proxy.Type+"://"+Config.Proxy.Host)
		if err != nil {
			panic(err.Error())
		}
	}

	//Start the bot
	bot, err := tgbotapi.NewBotAPI(Config.Token)
	if err != nil {
		log.Panic(err)
	}
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		h := sha256.New()
		h.Write([]byte(update.Message.Text))
		if hex.EncodeToString(h.Sum(nil)) == Config.Pass { //Hash the password and check it with the one user specified
			go func(chatID int64, firstName, lastName string) {
				msg := tgbotapi.NewMessage(chatID, "")
				page := "https://api.ipify.org"
				tr := &http.Transport{ //Use this to do not use proxy
					Proxy: nil,
				}
				client := &http.Client{Transport: tr}
				res, err := client.Get(page)
				if err != nil {
					msg.Text = "Error receiving IP:" + err.Error()
					LogVerbose("Error receiving IP:", err.Error())
				} else {
					ip, err := ioutil.ReadAll(res.Body)
					if err != nil {
						msg.Text = "Error reading web page: " + err.Error()
						LogVerbose("Error reading web page:", err.Error())
					} else {
						msg.Text = string(ip)
					}
					LogVerbose("Sending IP to ID", chatID, "; Name:", firstName, lastName)
				}
				_, err = bot.Send(msg)
				if err != nil {
					LogVerbose("Error sending IP:", err.Error())
				}
			}(update.Message.Chat.ID, update.Message.From.FirstName, update.Message.From.LastName)
		} else { //Password does not match
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Invalid password")
			LogVerbose("Invalid password from", update.Message.From.FirstName, update.Message.From.LastName, ", Username", update.Message.From.UserName, ",ID", update.Message.From.ID)
			_, err = bot.Send(msg)
			if err != nil {
				LogVerbose("Error sending IP:", err.Error())
			}
		}
	}
}

func LogVerbose(v ...interface{}) {
	if Verbose {
		log.Println(v)
	}
}
