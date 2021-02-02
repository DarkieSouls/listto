package bot

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/DarkieSouls/listto/internal/lists"
	"github.com/DarkieSouls/listto/internal/listtoErr"
)

// clearList wipes a list of it's values.
func (b *bot) clearList(guild, list, user string, roles []string) *discordgo.MessageEmbed {
	lis, err := b.DDB.GetList(guild, list)
	if err != nil {
		if err.Code == listtoErr.ListNotFound {
			return noList(list)
		}
		err.LogError()
		return failMsg()
	}

	if !lis.CanAccess(user, roles) {
		return noPerms(list)
	}

	lis.Clear()

	if err := b.DDB.PutList(lis); err != nil {
		err.LogError()
		return &discordgo.MessageEmbed{
			Description: fmt.Sprintf("I couldn't clear %s", list),
			Color:       red,
		}
	}

	return &discordgo.MessageEmbed{
		Description: fmt.Sprintf("I've cleared %s", list),
		Color:       green,
	}
}

// createList creates a new list.
func (b *bot) createList(guild, list string) *discordgo.MessageEmbed {
	lis := lists.NewList(guild, list, lists.PublicList)

	_, err := b.DDB.GetList(guild, list)
	if err == nil {
		return &discordgo.MessageEmbed{
			Description: fmt.Sprintf("I found another list already called %s", list),
			Color:       yellow,
		}
	}
	if err.Code != listtoErr.ListNotFound {
		err.LogError()
		return failMsg()
	}

	if err := b.DDB.PutList(lis); err != nil {
		err.LogError()
		return &discordgo.MessageEmbed{
			Description: fmt.Sprintf("I couldn't create a list called %s", list),
			Color:       red,
		}
	}

	return &discordgo.MessageEmbed{
		Description: fmt.Sprintf("%s list created!", list),
		Color:       green,
	}
}

// deleteList deletes a list.
func (b *bot) deleteList(guild, list, user string, roles []string) *discordgo.MessageEmbed {
	lis, aucErr := b.DDB.GetList(guild, list)
	if aucErr != nil {
		if aucErr.Code == listtoErr.ListNotFound {
			return noList(list)
		}
		aucErr.LogError()
		return failMsg()
	}

	if !lis.CanAccess(user, roles) {
		return noPerms(list)
	}

	err := b.DDB.DeleteList(guild, user)
	if err != nil {
		fmt.Println("failed to delete item", err)
		return &discordgo.MessageEmbed{
			Description: fmt.Sprintf("I couldn't delete %s", list),
			Color:       red,
		}
	}

	return &discordgo.MessageEmbed{
		Description: fmt.Sprintf("I have deleted %s", list),
		Color:       green,
	}
}

// getList gets a list.
func (b *bot) getList(guild, list, arg, user string, roles []string) *discordgo.MessageEmbed {
	lis, err := b.DDB.GetList(guild, list)
	if err != nil {
		if err.Code == listtoErr.ListNotFound {
			return noList(list)
		}
		err.LogError()
		return failMsg()
	}

	if !lis.CanAccess(user, roles) {
		return noPerms(list)
	}

	var fields []*discordgo.MessageEmbedField

	var values, desc string
	if arg == "" {
		desc = "Your List"
		for _, l := range lis.List {
			if len(values)+len(l.Value) > 1024 {
				fields = append(fields, &discordgo.MessageEmbedField{Name: list, Value: values})
				list = l.Value
				values = ""
				continue
			}
			values = fmt.Sprintf("%s\n%s", values, l.Value)
		}

		if values == "" {
			values = "This list is empty!"
		}

		fields = append(fields, &discordgo.MessageEmbedField{Name: list, Value: values})

		fields = append(fields, &discordgo.MessageEmbedField{Name: "List Entries", Value: fmt.Sprintf("%d", len(lis.List))})
	} else {
		desc = "Your Item"
		i, err := strconv.Atoi(arg)
		if err != nil {
			return &discordgo.MessageEmbed{
				Description: "The searched item needs to be a number!",
				Color:       yellow,
			}
		}

		values = lis.SelectItem(i)
		if values == "" {
			return &discordgo.MessageEmbed{
				Description: "I couldn't find an item at that position!",
				Color:       yellow,
			}
		}

		fields = append(fields, &discordgo.MessageEmbedField{Name: fmt.Sprintf("Item at position %d", i), Value: values})
	}

	return &discordgo.MessageEmbed{
		Description: desc,
		Color:       green,
		Fields:      fields,
	}
}

// listLists prints a list of lists on the server.
func (b *bot) listLists(guild, user string, roles []string) *discordgo.MessageEmbed {
	listtoLists, err := b.DDB.GetAllLists(guild)
	if err != nil {
		if err.Code == listtoErr.ListNotFound {
			return &discordgo.MessageEmbed{
				Description: "I couldn't find any lists for you",
				Color:       yellow,
			}
		}
		err.LogError()
		return failMsg()
	}

	var values string
	for _, lis := range listtoLists {
		if lis.CanAccess(user, roles) {
			values = fmt.Sprintf("%s\n%s", values, lis.Name)
		}
	}

	return &discordgo.MessageEmbed{
		Description: "I found these lists!",
		Color:       green,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Your lists",
				Value: values,
			},
		},
	}
}

// createPrivateList creates a list with limited access.
func (b *bot) createPrivateList(guild, list string, access []string) *discordgo.MessageEmbed {
	// todo: make this actually have an effect
	lis := lists.NewList(guild, list, lists.PrivateList)

	lis.AddAccess(access)

	_, err := b.DDB.GetList(guild, list)
	if err == nil {
		return &discordgo.MessageEmbed{
			Description: fmt.Sprintf("I found another list already called %s", list),
			Color:       yellow,
		}
	}
	if err.Code != listtoErr.ListNotFound {
		err.LogError()
		return failMsg()
	}

	if err := b.DDB.PutList(lis); err != nil {
		err.LogError()
		return &discordgo.MessageEmbed{
			Description: fmt.Sprintf("I couldn't create a list called %s", list),
			Color:       red,
		}
	}

	return &discordgo.MessageEmbed{
		Description: fmt.Sprintf("I created a private list called %s for you", list),
		Color:       green,
	}
}

// addAccessToList adds the supplied users and roles to the allowed users on a list.
func (b *bot) addAccessToList(guild, list string, access []string, user string, roles []string) *discordgo.MessageEmbed {
	lis, err := b.DDB.GetList(guild, list)
	if err != nil {
		if err.Code == listtoErr.ListNotFound {
			return noList(list)
		}
		err.LogError()
		return failMsg()
	}

	if !lis.CanAccess(user, roles) {
		return noPerms(list)
	}

	lis.AddAccess(access)

	if err := b.DDB.PutList(lis); err != nil {
		err.LogError()
		return &discordgo.MessageEmbed{
			Description: fmt.Sprintf("I couldn't update the permissions for %s", list),
			Color:       red,
		}
	}

	return &discordgo.MessageEmbed{
		Description: fmt.Sprintf("I have added those tags to allowed users on %s", list),
		Color:       green,
	}
}

func (b *bot) removeAccessFromList(guild, list string, access []string, user string, roles []string) *discordgo.MessageEmbed {
	lis, err := b.DDB.GetList(guild, list)
	if err != nil {
		if err.Code == listtoErr.ListNotFound {
			return noList(list)
		}
		err.LogError()
		return failMsg()
	}

	if !lis.CanAccess(user, roles) {
		return noPerms(list)
	}

	lis.RemoveAccess(access)

	if err := b.DDB.PutList(lis); err != nil {
		err.LogError()
		return &discordgo.MessageEmbed{
			Description: fmt.Sprintf("I couldn't update the permissions for %s", list),
			Color:       red,
		}
	}

	return &discordgo.MessageEmbed{
		Description: fmt.Sprintf("I have removed those tags from allowed users on %s", list),
		Color:       green,
	}
}

// sortList sorts the list.
func (b *bot) sortList(guild, list, arg, user string, roles []string) *discordgo.MessageEmbed {
	lis, err := b.DDB.GetList(guild, list)
	if err != nil {
		if err.Code == listtoErr.ListNotFound {
			return noList(list)
		}
		err.LogError()
		return failMsg()
	}

	if !lis.CanAccess(user, roles) {
		return noPerms(list)
	}

	sort := strings.ToLower(arg)
	if sort != "name" && sort != "time" {
		return &discordgo.MessageEmbed{
			Description: "Sorry! I only sort by \"name\" or \"time\"!",
			Color:       yellow,
		}
	}

	lis.Sort(sort)

	if err := b.DDB.PutList(lis); err != nil {
		err.LogError()
		return failMsg()
	}

	return &discordgo.MessageEmbed{
		Description: fmt.Sprintf("I have sorted %s by %s!", list, arg),
		Color:       green,
	}
}
