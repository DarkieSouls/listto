package bot

import (
	"fmt"
	"strconv"
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
		if err.Code == listtoErr.ListNotFound {
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
	if err.Code != listtoErr.ListNotFound {
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
		if aucErr.Code == listtoErr.ListNotFound {
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

	_, err := b.DDB.DeleteItem(input)
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

func (b *bot) editInList(guild, list, arg, user string, roles []string) *discordgo.MessageEmbed {
	lis, err := b.getDDB(guild, list)
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
				Description: fmt.Sprintf("%s doesn't seem to contain $s", list, updated),
				Color:       yellow,
			}
		}
	default:
		return &discordgo.MessageEmbed{
			Description: "You can only specify two arguments",
			Color:       yellow,
		}
	}

	if err := b.putDDB(lis); err != nil {
		err.LogError()
		return failMsg()
	}

	return &discordgo.MessageEmbed{
		Description: fmt.Sprintf("I have updated %s in %s", updated, list),
		Color:       green,
	}
}

// getList gets a list.
func (b *bot) getList(guild, list, arg, user string, roles []string) *discordgo.MessageEmbed {
	lis, err := b.getDDB(guild, list)
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

// listLists prints a list of lists on the server.
func (b *bot) listLists(guild, user string, roles []string) *discordgo.MessageEmbed {
	input := (&dynamodb.QueryInput{}).SetTableName(table).SetKeyConditionExpression("guild = :v1").
		SetExpressionAttributeValues(map[string]*dynamodb.AttributeValue{":v1": (&dynamodb.AttributeValue{}).SetS(guild)})

	output, err := b.DDB.Query(input)
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
	if err.Code != listtoErr.ListNotFound {
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
		Description: fmt.Sprintf("I created a private list called %s for you", list),
		Color:       green,
	}
}

// addAccessToList adds the supplied users and roles to the allowed users on a list.
func (b *bot) addAccessToList(guild, list string, access []string, user string, roles []string) *discordgo.MessageEmbed {
	lis, err := b.getDDB(guild, list)
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

	if err := b.putDDB(lis); err != nil {
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
	lis, err := b.getDDB(guild, list)
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

	if err := b.putDDB(lis); err != nil {
		err.LogError()
		return &discordgo.MessageEmbed{
			Description: fmt.Sprintf("I couldn't update the permissions for %s", list),
			Color: red,
		}
	}

	return &discordgo.MessageEmbed{
		Description: fmt.Sprintf("I have removed those tags from allowed users on %s", list),
		Color: green,
	}
}

// randomFromList selects a random element from the list.
func (b *bot) randomFromList(guild, list, user string, roles []string) *discordgo.MessageEmbed {
	lis, err := b.getDDB(guild, list)
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

	random := lis.SelectRandom()

	return &discordgo.MessageEmbed{
		Description: fmt.Sprintf("A random element from %s is %s", list, random),
		Color:       green,
	}
}

// removeFromList removes an item from the list.
func (b *bot) removeFromList(guild, list, arg, user string, roles []string) *discordgo.MessageEmbed {
	lis, lisErr := b.getDDB(guild, list)
	if lisErr != nil {
		if lisErr.Code == listtoErr.ListNotFound {
			return noList(list)
		}
		lisErr.LogError()
		return failMsg()
	}

	if !lis.CanAccess(user, roles) {
		return noPerms(list)
	}

	i, err := strconv.Atoi(arg)
	if err != nil {
		s := lis.RemoveItem(arg)
		if s == "" {
			return &discordgo.MessageEmbed{
				Description: fmt.Sprintf("%s doesn't seem to contain $s", list, arg),
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

	if lisErr := b.putDDB(lis); lisErr != nil {
		lisErr.LogError()
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

	_, err = b.DDB.PutItem(input)
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

	output, err := b.DDB.GetItem(input)
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
