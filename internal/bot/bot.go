package bot

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/bwmarrin/discordgo"

	"github.com/DarkieSouls/listto/cmd/config"
)

// bot holds all the info that needs to be passed around the bot.
type bot struct {
	dgo *discordgo.Session
	botID string
	conf *config.Config
	ddb *dynamodb.DynamoDB
}

// New creates a new bot instance.
func New(conf *config.Config, ddb *dynamodb.DynamoDB) *bot {
	return &bot{
		conf: conf,
		ddb: ddb,
	}
}

// Config gets the config stored in the bot.
func (b *bot) Config() *config.Config {
	return b.conf
}

// DDB gets the DDB instance stored in the bot.
func (b *bot) DDB() *dynamodb.DynamoDB {
	return b.ddb
}

// Start the bot listener.
func (b *bot) Start() {
	dgo, err := discordgo.New("Bot " + b.conf.Token())
	if err != nil {
		fmt.Println("could not create session", err)
		return
	}
	b.dgo = dgo

	u, err := b.dgo.User("@me")
	if err != nil {
		fmt.Println("could not get bot user", err)
	}

	b.botID = u.ID

	b.dgo.AddHandler(b.messageHandler())

	if err := b.dgo.Open(); err != nil {
		fmt.Println("could not open session", err)
		return
	}

	fmt.Println("The bot has awoken...")
}

// messageHandler returns a handlerfunc for messages.
func (b *bot) messageHandler() func(s *discordgo.Session, m *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		var list, arg, resp string

		if m.Author.ID == b.botID {
			return
		}

		if !strings.HasPrefix(m.Content, b.conf.Prefix()) {
			return
		}

		message := strings.Split(strings.TrimPrefix(m.Content, b.conf.Prefix()), " ")
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
		guild := m.GuildID

		command := strings.ToLower(message[0])
		switch command {
		case "add", "a":
			resp = b.addToList(guild, list, arg)
		case "clear", "cl":
			resp = b.clearList(guild, list)
		case "create", "c":
			resp = b.createList(guild, list)
		case "delete", "d":
			resp = b.deleteList(guild, list)
		case "get", "g":
			resp = b.getList(guild, list)
		case "help", "h":
			resp = b.help()
		case "list", "l":
			resp = b.listLists(guild)
		case "ping":
			resp = b.ping()
		case "prefix", "p":
			// It's not actually a list, but for ease of writing code I'm reusing the variable.
			resp = b.prefix(guild, list)
		case "privatecreate", "pc":
			resp = b.createPrivateList(guild, list, arg)
		case "random", "ra":
			resp = b.randomFromList(guild, list)
		case "remove", "re":
			resp = b.removeFromList(guild, list, arg)
		case "sort", "s":
			resp = b.sortList(guild, list, arg)
		}

		if resp != "" {
			_, _ = s.ChannelMessageSend(m.ChannelID, resp)
		}
	}
}
