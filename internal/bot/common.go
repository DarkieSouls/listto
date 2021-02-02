package bot

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

const (
	red    = 0xDD3311
	yellow = 0xFFDD11
	green  = 0x33DD33
	blue   = 0x2255EE
)

func failMsg() *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Description: "Oops, I had a problem doing that for you",
		Color:       red,
	}
}

func noList(list string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Description: fmt.Sprintf("I couldn't find a list called %s", list),
		Color:       yellow,
	}
}

func noPerms(list string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Description: fmt.Sprintf("You have not been given permission to use %s", list),
		Color:       yellow,
	}
}

// ping the bot.
func (b *bot) ping() *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Description: "pong",
		Color:       green,
	}
}

// help prints how to use the bot.
func (b *bot) help(arg string) *discordgo.MessageEmbed {
	p := b.Config.Prefix

	switch strings.ToLower(arg) {
	case "lists":
		return &discordgo.MessageEmbed{
			Description: "Here are some commands involving lists:",
			Color:       blue,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:  "clear, cl",
					Value: fmt.Sprintf("Clears a list\n__Example__:\n%sclear MyList", p),
				},
				{
					Name:  "create, c",
					Value: fmt.Sprintf("Creates a new list. Lists cannot contain spaces\n__Example__:\n%screate MyList", p),
				},
				{
					Name: "createprivate, cp",
					Value: fmt.Sprintf("Creates a new private list. You can specify allowed users and roles after the list name. Will default to just you if left blank"+
						"\n__Examples__:\n%screateprivate MyList @UserOne\n%scp MyList @MyRole", p, p),
				},
				{
					Name:  "addtoprivate, ap",
					Value: fmt.Sprintf("Adds the specified roles or users to a private list\n__Example__:\n%saddtoprivate MyList @UserOne", p),
				},
				{
					Name:  "removefromprivate, rp",
					Value: fmt.Sprintf("Removes the specified roles or users from a private list\n__Example__:\n%srp MyList @Role", p),
				},
				{
					Name:  "delete, d",
					Value: fmt.Sprintf("Deletes a list\n__Example__:\n%sdelete MyList", p),
				},
				{
					Name:  "get, g",
					Value: fmt.Sprintf("Gets a list\n__Example__:\n%sget MyList", p),
				},
				{
					Name:  "sort, s",
					Value: fmt.Sprintf("Sorts a list by either name or time\n__Example__\n%ssort MyList name", p),
				},
			},
		}
	case "items":
		return &discordgo.MessageEmbed{
			Description: "Here are some commands involving list items:",
			Color:       blue,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:  "add, a",
					Value: fmt.Sprintf("Adds an item to a list, items can have spaces\n__Example__:\n%sadd MyList My Item", p),
				},
				{
					Name: "edit, e",
					Value: fmt.Sprintf("Edit an item in a list. You can specify the item to edit either by it's index, or it's value. If you search by index, then note that 0 is the first item in the list,"+
						" and the new value should not be surrounded by \"s. If you search by value, then both values need to be surrounded with \"s"+
						"\n__Example__:\n%sedit MyList 0 My new and improved item\n%se MyList \"My Old Item\" \"My New Item\"", p, p),
				},
				{
					Name: "get, g",
					Value: fmt.Sprintf("Get an item from a list. You specify the item by using the index as specified above."+
						"\n__Example__:\n%sg MyList 0", p),
				},
				{
					Name:  "random, rv",
					Value: fmt.Sprintf("Selects a random item from a list\n__Example__:\n%srv MyList", p),
				},
				{
					Name: "remove, r",
					Value: fmt.Sprintf("Removes an item from a list. You can either type the item in full, or the item index"+
						"\n__Example__:\n%sremove MyList MyItem\n%sr MyList 0", p, p),
				},
			},
		}
	default:
		return &discordgo.MessageEmbed{
			Description: "Listto does some list management things! Here's some generic commands:",
			Color:       blue,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:  "help, h",
					Value: fmt.Sprintf("Displays a help message!\nCan accept arguments of lists and items\n__Examples__:\n%shelp\n%sh lists", p, p),
				},
				{
					Name:  "list, l",
					Value: fmt.Sprintf("Lists all lists on the server\n__Example__:\n%sl", p),
				},
				{
					Name:  "ping",
					Value: fmt.Sprintf("Check if I'm alive\n__Example__:\n%sping", p),
				},
			},
		}
	}
}
