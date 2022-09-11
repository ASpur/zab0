package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

var config *Config

func main() {
	config = &Config{}
	_, err := os.OpenFile("./config.json", os.O_RDWR, 0644) //load config file
	if err == nil {
		cf, err := ioutil.ReadFile("./config.json")
		if err != nil {
			panic(err)
		}
		err = json.Unmarshal([]byte(cf), config)

		if err != nil {
			panic(err)
		}
	} else if errors.Is(err, os.ErrNotExist) {
		mar, _ := json.Marshal(Config{})
		os.WriteFile("config.json", []byte(mar), 0644)

		fmt.Println("Config file not found. Created new file at run path. Please configure file and restart program.")
		os.Exit(0)

	} else {
		panic(err)
	}

	if config.Token == "" {
		fmt.Println("It looks like you don't have your token set. Please set your token in config.json")
		os.Exit(0)
	}

	d, err := discordgo.New("Bot " + config.Token)

	if err != nil {
		panic(err)
	}

	if d == nil {
		panic("Discord Is null :(")
	}

	d.AddHandler(messageCreate)

	d.Identify.Intents = discordgo.IntentsGuildMessages

	d.Open()

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	d.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}
	//Ignore if author is not an authorized user
	var authorized = false
	for _, id := range config.AuthorizedUsers {
		if id == m.Author.ID {
			authorized = true
			break
		}
	}
	if !authorized {
		return
	}

}

type Config struct {
	Token           string
	AuthorizedUsers []string
}
