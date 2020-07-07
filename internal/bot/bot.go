package bot

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/DarkieSouls/listto/cmd/config"
)

var (
	dgo *discordgo.Session
	botID string
	conf *config.Config
)

func Start(c *config.Config) {
	conf = c

	dgo, err := discordgo.New("Bot " + conf.Token())
	if err != nil {
		fmt.Println("That's an error!", err)
		return
	}

	u, err := dgo.User("@me")
	if err != nil {
		fmt.Println("That's an error!", err)
	}

	botID = u.ID

	dgo.AddHandler(messageHandler)

	if err := dgo.Open(); err != nil {
		fmt.Println("That's an error!", err)
		return
	}

	fmt.Println("Some success")
}

func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	var list, arg, resp string

	if m.Author.ID == botID {
		return
	}

	if !strings.HasPrefix(m.Content, conf.Prefix()) {
		return
	}

	message := strings.Split(strings.TrimPrefix(m.Content, conf.Prefix()), " ")
	if len(message) == 0 {
		return
	}

	if len(message) > 1 {
		list = message[1]
	}
	if len(message) > 2 {
		arg = message[2]
		for i := 3; i < len(message); i++ {
			arg = arg + " " + message[i]
		}
	}

	command := strings.ToLower(message[0])
	switch command {
	case "add", "a":
		resp = addToList(list, arg)
	case "clear", "cl":
		resp = clearList(list)
	case "create", "c":
		resp = createList(list)
	case "delete", "d":
		resp = deleteList(list)
	case "help", "h":
		resp = help()
	case "list", "l":
		resp = listLists()
	case "ping":
		resp = ping()
	case "prefix", "p":
		// It's not actually a list, but for ease of writing code I'm reusing teh variable.
		resp = prefix(list, conf)
	case "privatecreate", "pc":
		resp = createPrivateList(list, arg)
	case "random", "ra":
		resp = randomFromList(list)
	case "remove", "re":
		resp = removeFromList(list, arg)
	case "sort", "s":
		resp = sortList(list, arg)
	}

	if resp != "" {
		_, _ = s.ChannelMessageSend(m.ChannelID, resp)
	}
}
