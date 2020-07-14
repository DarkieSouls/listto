package bot

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/bwmarrin/discordgo"

	"github.com/DarkieSouls/listto/internal/lists"
	"github.com/DarkieSouls/listto/internal/listtoErr"
)

const (
	table  = "listto_lists"
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

// addToList adds a value to a list.
func (b *bot) addToList(guild, list, arg, user string, roles []string) *discordgo.MessageEmbed {
	lis, err := b.getDDB(guild, list)
	if err != nil {
		if err.Code() == listtoErr.ListNotFound {
			return noList(list)
		}
		err.LogError()
		return failMsg()
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

	if err := b.putDDB(lis); err != nil {
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

// clearList wipes a list of it's values.
func (b *bot) clearList(guild, list, user string, roles []string) *discordgo.MessageEmbed {
	lis, err := b.getDDB(guild, list)
	if err != nil {
		if err.Code() == listtoErr.ListNotFound {
			return noList(list)
		}
		err.LogError()
		return failMsg()
	}

	if !lis.CanAccess(user, roles) {
		return noPerms(list)
	}

	lis.Clear()

	if err := b.putDDB(lis); err != nil {
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
	lis := lists.NewList(guild, list, false)

	_, err := b.getDDB(guild, list)
	if err == nil {
		return &discordgo.MessageEmbed{
			Description: fmt.Sprintf("I found another list already called %s", list),
			Color:       yellow,
		}
	}
	if err.Code() != listtoErr.ListNotFound {
		err.LogError()
		return failMsg()
	}

	if err := b.putDDB(lis); err != nil {
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
	lis, aucErr := b.getDDB(guild, list)
	if aucErr != nil {
		if aucErr.Code() == listtoErr.ListNotFound {
			return noList(list)
		}
		aucErr.LogError()
		return failMsg()
	}

	if !lis.CanAccess(user, roles) {
		return noPerms(list)
	}

	input := (&dynamodb.DeleteItemInput{}).SetTableName(table).SetKey(map[string]*dynamodb.AttributeValue{
		"guild": (&dynamodb.AttributeValue{}).SetS(guild),
		"name":  (&dynamodb.AttributeValue{}).SetS(list),
	})

	_, err := b.ddb.DeleteItem(input)
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
func (b *bot) getList(guild, list, user string, roles []string) *discordgo.MessageEmbed {
	lis, err := b.getDDB(guild, list)
	if err != nil {
		if err.Code() == listtoErr.ListNotFound {
			return noList(list)
		}
		err.LogError()
		return failMsg()
	}

	if !lis.CanAccess(user, roles) {
		return noPerms(list)
	}

	var values string
	for _, l := range lis.List {
		values = fmt.Sprintf("%s\n%s", values, l.Value)
	}

	if values == "" {
		values = "This list is empty!"
	}

	return &discordgo.MessageEmbed{
		Description: "Your list",
		Color:       green,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  list,
				Value: values,
			},
			{
				Name:  "List entries",
				Value: fmt.Sprintf("%d", len(lis.List)),
			},
		},
	}
}

// help prints how to use the bot.
func (b *bot) help() *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Description: "Listto does some list management things! Here's what I can do so far:",
		Color:       blue,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "add, a",
				Value: "Adds an item to a list, items can have spaces\nExample: ^add MyList My Item",
			},
			{
				Name:  "clear, cl",
				Value: "Clears a list\nExample: ^clear MyList",
			},
			{
				Name:  "create, c",
				Value: "Creates a new list, lists cannot contain spaces\nExample: ^create MyList",
			},
			{
				Name:  "delete, d",
				Value: "Deletes a list\nExample: ^delete MyList",
			},
			{
				Name:  "get, g",
				Value: "Gets a list\nExample: ^get MyList",
			},
			{
				Name:  "help, h",
				Value: "Displays this message!\nExample: ^h",
			},
			{
				Name:  "list, l",
				Value: "Lists all lists on the server\nExample: ^l",
			},
			{
				Name:  "random, rv",
				Value: "Selects a random item from a list\nExample: ^rv MyList",
			},
			{
				Name:  "remove, r",
				Value: "Removes an item from a list\nExample: ^remove MyList MyItem",
			},
			{
				Name:  "sort, s",
				Value: "sorts a list by either name or time\nExample ^sort MyList name",
			},
		},
	}
}

// listLists prints a list of lists on the server.
func (b *bot) listLists(guild, user string, roles []string) *discordgo.MessageEmbed {
	input := (&dynamodb.QueryInput{}).SetTableName(table).SetKeyConditionExpression("guild = :v1").
		SetExpressionAttributeValues(map[string]*dynamodb.AttributeValue{":v1": (&dynamodb.AttributeValue{}).SetS(guild)})

	output, err := b.ddb.Query(input)
	if err != nil {
		fmt.Println("failed to list lists", err)
		return failMsg()
	}

	if len(output.Items) < 1 {
		return &discordgo.MessageEmbed{
			Description: "I couldn't find any lists for you",
			Color:       yellow,
		}
	}

	var values string
	for _, v := range output.Items {
		lis := new(lists.ListtoList)
		if err := dynamodbattribute.UnmarshalMap(v, &lis); err != nil {
			return failMsg()
		}
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

// ping the bot.
func (b *bot) ping() *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Description: "pong",
		Color:       green,
	}
}

// createPrivateList creates a list with limited access.
func (b *bot) createPrivateList(guild, list string, access []string) *discordgo.MessageEmbed {
	// todo: make this actually have an effect
	lis := lists.NewList(guild, list, true)

	lis.AddAccess(access)

	_, err := b.getDDB(guild, list)
	if err == nil {
		return &discordgo.MessageEmbed{
			Description: fmt.Sprintf("I found another list already called %s", list),
			Color:       yellow,
		}
	}
	if err.Code() != listtoErr.ListNotFound {
		err.LogError()
		return failMsg()
	}

	if err := b.putDDB(lis); err != nil {
		err.LogError()
		return &discordgo.MessageEmbed{
			Description: fmt.Sprintf("I couldn't create a list called %s", list),
			Color:       red,
		}
	}

	return &discordgo.MessageEmbed{
		Description: fmt.Sprintf("I created %s for you, but privacy currently doesn't do anything", list),
		Color:       green,
	}
}

// addAccessToList adds the supplied users and roles to the allowed users on a list.
func (b *bot) addAccessToList(guild, list string, access []string, user string, roles []string) *discordgo.MessageEmbed {
	lis, err := b.getDDB(guild, list)
	if err != nil {
		if err.Code() == listtoErr.ListNotFound {
			return noList(list)
		}
		err.LogError()
		return failMsg()
	}

	if !lis.CanAccess(user, roles) {
		return noPerms(list)
	}

	lis.AddAccess(access)

	return &discordgo.MessageEmbed{
		Description: fmt.Sprintf("I have added those tags to allowed users on %s", list),
		Color:       green,
	}
}

// randomFromList selects a random element from the list.
func (b *bot) randomFromList(guild, list, user string, roles []string) *discordgo.MessageEmbed {
	lis, err := b.getDDB(guild, list)
	if err != nil {
		if err.Code() == listtoErr.ListNotFound {
			return noList(list)
		}
		err.LogError()
		return failMsg()
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
	lis, err := b.getDDB(guild, list)
	if err != nil {
		if err.Code() == listtoErr.ListNotFound {
			return noList(list)
		}
		err.LogError()
		return failMsg()
	}

	if !lis.CanAccess(user, roles) {
		return noPerms(list)
	}

	lis.RemoveItem(arg)

	if err := b.putDDB(lis); err != nil {
		err.LogError()
		return failMsg()
	}

	return &discordgo.MessageEmbed{
		Description: fmt.Sprintf("I have removed %s from %s", arg, list),
		Color:       green,
	}
}

// sortList sorts the list.
func (b *bot) sortList(guild, list, arg, user string, roles []string) *discordgo.MessageEmbed {
	lis, err := b.getDDB(guild, list)
	if err != nil {
		if err.Code() == listtoErr.ListNotFound {
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

	if err := b.putDDB(lis); err != nil {
		err.LogError()
		return failMsg()
	}

	return &discordgo.MessageEmbed{
		Description: fmt.Sprintf("I have sorted %s by %s!", list, arg),
		Color:       green,
	}
}

func (b *bot) putDDB(in interface{}) (lisErr *listtoErr.ListtoError) {
	item, err := dynamodbattribute.MarshalMap(in)
	if err != nil {
		lisErr = listtoErr.ConvertError(err)
		return
	}

	input := (&dynamodb.PutItemInput{}).SetTableName(table).SetItem(item)

	_, err = b.DDB().PutItem(input)
	if err != nil {
		lisErr = listtoErr.ConvertError(err)
	}

	return
}

func (b *bot) getDDB(guild, lis string) (list *lists.ListtoList, lisErr *listtoErr.ListtoError) {
	defer func() {
		if lisErr != nil {
			lisErr.SetCallingMethodIfNil("getDDB")
		}
	}()

	input := (&dynamodb.GetItemInput{}).SetTableName(table).SetKey(map[string]*dynamodb.AttributeValue{
		"guild": (&dynamodb.AttributeValue{}).SetS(guild),
		"name":  (&dynamodb.AttributeValue{}).SetS(lis),
	})

	output, err := b.ddb.GetItem(input)
	if err != nil {
		lisErr = listtoErr.ConvertError(err)
		return
	}

	if len(output.Item) < 1 {
		lisErr = listtoErr.ListNotFoundError(lis)
		return
	}

	list = new(lists.ListtoList)
	if err := dynamodbattribute.UnmarshalMap(output.Item, &list); err != nil {
		lisErr = listtoErr.ConvertError(err)
	}

	return
}
