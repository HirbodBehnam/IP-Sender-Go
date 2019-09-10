package main

import (
	"encoding/json"
	"flag"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"golang.org/x/crypto/bcrypt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
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

const Version = "1.1.0 / Build 2"

func main() {
	var ConfigFileName string
	{ //Parse arguments
		configFileName := flag.String("config", "config.json", "The config filename")
		pass := flag.String("hash", "", "Pass a password with this to generate the hashed password.")
		verbose := flag.Bool("v", false, "Verbose mode")
		help := flag.Bool("h", false, "Show help")
		flag.Parse()

		if *help {
			fmt.Println("Created by Hirbod Behnam")
			fmt.Println("Source at https://github.com/HirbodBehnam/IP-Sender-Go")
			fmt.Println("Version", Version)
			flag.PrintDefaults()
			os.Exit(0)
		}

		if *pass != "" { //Hash the password and print it for user
			fmt.Println("Generating hash for:", *pass)
			b, _ := bcrypt.GenerateFromPassword([]byte(*pass), 14)
			fmt.Println(string(b))
			os.Exit(0)
		}

		Verbose = *verbose
		if Verbose {
			fmt.Println("Verbose mode on")
		}
		ConfigFileName = *configFileName
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

		if err = bcrypt.CompareHashAndPassword([]byte(Config.Pass), []byte(update.Message.Text)); err == nil { //Hash the password and check it with the one user specified
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
