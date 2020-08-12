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
	dgo   *discordgo.Session
	botID string
	conf  *config.Config
	ddb   *dynamodb.DynamoDB
}

// New creates a new bot instance.
func New(conf *config.Config, ddb *dynamodb.DynamoDB) *bot {
	return &bot{
		conf: conf,
		ddb:  ddb,
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

	b.dgo.UpdateStatus(0, fmt.Sprintf("with %shelp", b.conf.Prefix()))

	fmt.Println("The bot has awoken...")
}

// messageHandler returns a handlerfunc for messages.
func (b *bot) messageHandler() func(s *discordgo.Session, m *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		var list, arg, user string
		var roles []string

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
		user = m.Author.ID
		roles = m.Member.Roles

		var resp *discordgo.MessageEmbed

		command := strings.ToLower(message[0])
		switch command {
		case "add", "a":
			resp = b.addToList(guild, list, arg, user, roles)
		case "clear", "cl":
			resp = b.clearList(guild, list, user, roles)
		case "create", "c":
			resp = b.createList(guild, list)
		case "delete", "d":
			resp = b.deleteList(guild, list, user, roles)
		case "edit", "e":
			resp = b.editInList(guild, list, arg, user, roles)
		case "get", "g":
			resp = b.getList(guild, list, user, roles)
		case "help", "h":
			resp = b.help(list)
		case "list", "l":
			resp = b.listLists(guild, user, roles)
		case "ping":
			resp = b.ping()
		case "createprivate", "cp":
			var access []string

			if len(m.MentionRoles) != 0 {
				for _, r := range m.MentionRoles {
					access = append(access, r)
				}
			}

			if len(m.Mentions) != 0 {
				for _, u := range m.Mentions {
					access = append(access, u.ID)
				}
			}

			access = append(access, m.Author.ID)

			resp = b.createPrivateList(guild, list, access)
		case "addtoPrivate", "ap":
			var access []string
			if len(m.MentionRoles) != 0 {
				for _, r := range m.MentionRoles {
					access = append(access, r)
				}
			}

			if len(m.Mentions) != 0 {
				for _, u := range m.Mentions {
					access = append(access, u.ID)
				}
			}

			resp = b.addAccessToList(guild, list, access, user, roles)
		case "random", "rv":
			resp = b.randomFromList(guild, list, user, roles)
		case "remove", "r":
			resp = b.removeFromList(guild, list, arg, user, roles)
		case "sort", "s":
			resp = b.sortList(guild, list, arg, user, roles)
		}

		if resp != nil {
			_, err := s.ChannelMessageSendEmbed(m.ChannelID, resp)
			if err != nil {
				fmt.Println("failed to send to discord", err)
				_, _ = s.ChannelMessageSendEmbed(m.ChannelID, failMsg())
			}
		}
	}
}
