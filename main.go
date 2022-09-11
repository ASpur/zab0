package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

var config *Config

func main() {
	//Load Config
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

	for key, val := range config.Procmap {
		wd, _ := os.Getwd()
		val.Args = strings.ReplaceAll(val.Args, "$WORKINGDIR", wd)
		val.Command = strings.ReplaceAll(val.Command, "$WORKINGDIR", wd)

		config.Procmap[key] = val

	}
	//check if setup
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
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, syscall.SIGTERM)
	<-sc
	fmt.Println("Closing")
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
	//return if not starting with '!'
	if []byte(m.Content)[0] != byte('!') {
		return
	}

	tokens := strings.Split(m.Content, " ")

	if len(tokens) == 2 && tokens[0] == "!launch" {
		proc, exists := config.Procmap[tokens[1]]

		if !exists {
			s.ChannelMessageSend(m.ChannelID, "Command "+tokens[1]+" is not defined in config file")
			return
		}
		fmt.Println("Running sub process " + tokens[1])

		if proc.Single && proc.cmd != nil {
			err := proc.cmd.Process.Signal(syscall.Signal(0))
			fmt.Println(err)
			if err == nil {
				fmt.Println("already running")
			} else if string(err.Error()) == "not supported by windows" {
				_, err = os.FindProcess(proc.cmd.Process.Pid) //seems to print to console sometimes.
				if err == nil {
					fmt.Println("already running")
					return
				}
			}
		}

		cmd := exec.Command(proc.Command, proc.Args)
		if proc.Single {
			proc.cmd = cmd
			config.Procmap[tokens[1]] = proc
		}
		if proc.Output {
			o, err := cmd.Output()
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error starting process")
				fmt.Println(err)
				return
			}
			s.ChannelMessageSend(m.ChannelID, string(string(o)))
		} else {
			err := cmd.Start()

			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error starting process")
				fmt.Println(err)
			}
		}
		// fmt.Println("hello")
		// config.Procmap[tokens[1]] = proc
		// fmt.Println(config.Procmap[tokens[1]].cmd)
	}

}

type Proc struct {
	Command string
	Args    string
	Output  bool
	Single  bool
	cmd     *exec.Cmd
}

type Config struct {
	Token           string
	AuthorizedUsers []string
	Procmap         map[string]Proc
}
