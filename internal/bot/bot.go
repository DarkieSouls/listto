package bot

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/DarkieSouls/listto/cmd/config"
	"github.com/DarkieSouls/listto/internal/lists"
	"github.com/DarkieSouls/listto/internal/listtoErr"
)

type DDB interface {
	GetList(string, string) (*lists.ListtoList, *listtoErr.ListtoError)
	GetAllLists(string) ([]*lists.ListtoList, *listtoErr.ListtoError)
	PutList(interface{}) *listtoErr.ListtoError
	DeleteList(string, string) *listtoErr.ListtoError
}

// bot holds all the info that needs to be passed around the bot.
type bot struct {
	Dgo    *discordgo.Session
	BotID  string
	Config *config.Config
	DDB    DDB
}

// New creates a new bot instance.
func New(conf *config.Config, ddb DDB) *bot {
	return &bot{
		Config: conf,
		DDB:    ddb,
	}
}

// Start the bot listener.
func (b *bot) Start() {
	dgo, err := discordgo.New("Bot " + b.Config.Token)
	if err != nil {
		fmt.Println("could not create session", err)
		return
	}
	b.Dgo = dgo

	u, err := b.Dgo.User("@me")
	if err != nil {
		fmt.Println("could not get bot user", err)
	}

	b.BotID = u.ID

	b.Dgo.AddHandler(b.messageHandler())

	if err := b.Dgo.Open(); err != nil {
		fmt.Println("could not open session", err)
		return
	}

	b.Dgo.UpdateStatus(0, fmt.Sprintf("with %shelp", b.Config.Prefix))

	fmt.Println("The bot has awoken...")
}

// messageHandler returns a handlerfunc for messages.
func (b *bot) messageHandler() func(s *discordgo.Session, m *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		var list, arg, user string
		var roles []string

		if m.Author.ID == b.BotID {
			return
		}

		if !strings.HasPrefix(m.Content, b.Config.Prefix) {
			return
		}

		message := strings.Split(strings.TrimPrefix(m.Content, b.Config.Prefix), " ")
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
			resp = b.getList(guild, list, arg, user, roles)
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
		case "addtoprivate", "ap":
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
		case "removefromprivate", "rp":
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

			resp = b.removeAccessFromList(guild, list, access, user, roles)
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
