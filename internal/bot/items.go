package bot

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

// addToList adds a value to a list.
func (b *bot) addToList(guild, list, arg, user string, roles []string) *discordgo.MessageEmbed {
	lis, msg := b.getDDBList(guild, list, user)
	if msg != nil {
		return msg
	}

	if !lis.CanAccess(user, roles) {
		return noPerms(list)
	}

	dupe := "!"
	for _, l := range lis.List {
		if l.Value == arg {
			dupe = ", again"
		}
	}

	lis.AddItem(arg, time.Now().Unix())

	if err := b.DDB.PutList(lis); err != nil {
		err.LogError()
		return &discordgo.MessageEmbed{
			Description: fmt.Sprintf("I couldn't add %s to %s", arg, list),
			Color:       red,
		}
	}

	ls := list + dupe

	return &discordgo.MessageEmbed{
		Description: fmt.Sprintf("I added %s to %s", arg, ls),
		Color:       green,
	}
}

func (b *bot) editInList(guild, list, arg, user string, roles []string) *discordgo.MessageEmbed {
	lis, msg := b.getDDBList(guild, list, user)
	if msg != nil {
		return msg
	}

	if !lis.CanAccess(user, roles) {
		return noPerms(list)
	}

	var updated string

	args := strings.Split(arg, `" "`)
	switch len(args) {
	case 1:
		args = strings.Split(arg, " ")
		i, err := strconv.Atoi(args[0])
		if err != nil {
			return &discordgo.MessageEmbed{
				Description: "The first argument needs to be a number or existing value!",
				Color:       yellow,
			}
		}

		newVal := strings.Join(args[1:], " ")

		updated = lis.EditIndex(i, newVal)
		if updated == "" {
			return &discordgo.MessageEmbed{
				Description: fmt.Sprintf("%s doesn't seem to have that many items!", list),
				Color:       yellow,
			}
		}
	case 2:
		updated = strings.TrimPrefix(args[0], "\"")
		s := lis.EditItem(updated, strings.TrimSuffix(args[1], "\""))
		if s == "" {
			return &discordgo.MessageEmbed{
				Description: fmt.Sprintf("%s doesn't seem to contain %s", list, updated),
				Color:       yellow,
			}
		}
	default:
		return &discordgo.MessageEmbed{
			Description: "You can only specify two arguments",
			Color:       yellow,
		}
	}

	if err := b.DDB.PutList(lis); err != nil {
		err.LogError()
		return failMsg()
	}

	return &discordgo.MessageEmbed{
		Description: fmt.Sprintf("I have updated %s in %s", updated, list),
		Color:       green,
	}
}

// randomFromList selects a random element from the list.
func (b *bot) randomFromList(guild, list, user string, roles []string) *discordgo.MessageEmbed {
	lis, msg := b.getDDBList(guild, list, user)
	if msg != nil {
		return msg
	}

	if !lis.CanAccess(user, roles) {
		return noPerms(list)
	}

	random := lis.SelectRandom()

	return &discordgo.MessageEmbed{
		Description: fmt.Sprintf("A random element from %s is %s", list, random),
		Color:       green,
	}
}

// removeFromList removes an item from the list.
func (b *bot) removeFromList(guild, list, arg, user string, roles []string) *discordgo.MessageEmbed {
	lis, msg := b.getDDBList(guild, list, user)
	if msg != nil {
		return msg
	}

	if !lis.CanAccess(user, roles) {
		return noPerms(list)
	}

	i, err := strconv.Atoi(arg)
	if err != nil {
		s := lis.RemoveItem(arg)
		if s == "" {
			return &discordgo.MessageEmbed{
				Description: fmt.Sprintf("%s doesn't seem to contain %s", list, arg),
				Color:       yellow,
			}
		}
	} else {
		arg = lis.RemoveIndex(i)
		if arg == "" {
			return &discordgo.MessageEmbed{
				Description: fmt.Sprintf("%s doesn't seem to have that many items!", list),
				Color:       yellow,
			}
		}
	}

	if lisErr := b.DDB.PutList(lis); lisErr != nil {
		lisErr.LogError()
		return failMsg()
	}

	return &discordgo.MessageEmbed{
		Description: fmt.Sprintf("I have removed %s from %s", arg, list),
		Color:       green,
	}
}
