package bot

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"

	"github.com/DarkieSouls/listto/internal/lists"
	"github.com/DarkieSouls/listto/internal/listtoErr"
)

const (
	table = "listto_lists"
	failMsg = "Oops, I had a problem doing that for you"
)

func noList(list string) string {
	return fmt.Sprintf("I couldn't find a list called %s", list)
}

// addToList adds a value to a list.
func (b *bot) addToList(guild, list, arg string) string {
	lis, err := b.getDDB(guild, list)
	if err != nil {
		if err.Code() == listtoErr.ListNotFound {
			return noList(list)
		}
		err.LogError()
		return failMsg
	}

	lis.AddItem(arg, time.Now().Unix())

	if err := b.putDDB(lis); err != nil {
		err.LogError()
		return fmt.Sprintf("I couldn't add %s to %s", arg, list)
	}

	return fmt.Sprintf("I added %s to %s!", arg, list)
}

// clearList wipes a list of it's values.
func (b *bot) clearList(guild, list string) string {
	lis, err := b.getDDB(guild, list)
	if err != nil {
		if err.Code() == listtoErr.ListNotFound {
			return noList(list)
		}
		err.LogError()
		return failMsg
	}

	lis.Clear()

	if err := b.putDDB(lis); err != nil {
		err.LogError()
		return fmt.Sprintf("I couldn't clear %s", list)
	}

	return fmt.Sprintf("I've cleared %s", list)
}

// createList creates a new list.
func (b *bot) createList(guild, list string) string {
	lis := lists.NewList(guild, list, false)

	_, err := b.getDDB(guild, list)
	if err == nil {
		return fmt.Sprintf("I found another list already called %s", list)
	}
	if err.Code() != listtoErr.ListNotFound {
		err.LogError()
		return failMsg
	}

	if err := b.putDDB(lis); err != nil {
		err.LogError()
		return fmt.Sprintf("I couldn't create a list called %s", list)
	}

	return fmt.Sprintf("%s list created!", list)
}

// deleteList deletes a list.
func (b *bot) deleteList(guild, list string) string {
	input := (&dynamodb.DeleteItemInput{}).SetTableName(table).SetKey(map[string]*dynamodb.AttributeValue{
		"guild": (&dynamodb.AttributeValue{}).SetS(guild),
		"name":(&dynamodb.AttributeValue{}).SetS(list),
	})

	_, err := b.ddb.DeleteItem(input)
	if err != nil {
		fmt.Println("failed to delete item", err)
		return fmt.Sprintf("I couldn't delete %s", list)
	}

	return fmt.Sprintf("I have deleted %s", list)
}

// getList gets a list.
func (b *bot) getList(guild, list string) string {
	lis, err := b.getDDB(guild, list)
	if err != nil {
		if err.Code() == listtoErr.ListNotFound {
			return fmt.Sprintf("I couldn't find a list called %s", list)
		}
		err.LogError()
		return failMsg
	}

	resp := fmt.Sprintf("Your list %s:", list)
	for _, l := range lis.List {
		resp = fmt.Sprintf("%s\n%s", resp, l.Value)
	}

	return resp
}

// help prints how to use the bot.
func (b *bot) help() string {
	resp := "Listto does some list management things! Here's what we've got so far:\n" +
		"add, a: adds a value to a list\n" +
		"clear, cl: clears the list\n" +
		"create, c: creates the list\n" +
		"delete, d: deletes the list\n" +
		"help, h: displays this message!\n" +
		"list, l: lists the lists\n" +
		"random, ra: selects a random item from the list\n" +
		"remove, re: removes a value from the list\n" +
		"sort, s: sorts the list"

	return resp
}

// listLists prints a list of lists on the server.
func (b *bot) listLists(guild string) string {
	input := (&dynamodb.QueryInput{}).SetTableName(table).SetKeyConditionExpression(fmt.Sprintf("guild = :%s", guild))

	output, err := b.ddb.Query(input)
	if err != nil {
		fmt.Println("failed to list lists", err)
		return failMsg
	}

	if len(output.Items) < 1 {
		return "I couldn't find any lists for you"
	}

	resp := "I found these lists for you:"
	for _, v := range output.Items {
		lis := new(lists.ListtoList)
		if err := dynamodbattribute.UnmarshalMap(v, &lis); err != nil {
			return failMsg
		}
		resp = fmt.Sprintf("%s\n%s", resp, lis)
	}

	return resp
}

// ping the bot.
func (b *bot) ping() string {
	return "pong"
}

// prefix updates the bot prefix.
func (b *bot) prefix(guild, p string) string {
	// todo: make this bot owner only
	b.Config().SetPrefix(p)
	return fmt.Sprintf("Set Listto prefix to %s", p)
}

// createPrivateList creates a list with limited access.
func (b *bot) createPrivateList(guild, list, arg string) string {
	// todo: make this actually have an effect
	lis := lists.NewList(guild, list, true)

	_, err := b.getDDB(guild, list)
	if err != nil {
		return fmt.Sprintf("I found another list already called %s", list)
	}
	if err.Code() != listtoErr.ListNotFound {
		err.LogError()
		return failMsg
	}

	if err := b.putDDB(lis); err != nil {
		err.LogError()
		return fmt.Sprintf("I couldn't create a list called %s", list)
	}

	return fmt.Sprintf("I created %s for you, but privacy currently doesn't do anything", list)
}

// randomFromList selects a random element from the list.
func (b *bot) randomFromList(guild, list string) string {
	lis, err := b.getDDB(guild, list)
	if err != nil {
		if err.Code() == listtoErr.ListNotFound {
			return noList(list)
		}
		err.LogError()
		return failMsg
	}

	random := lis.SelectRandom()

	return fmt.Sprintf("A random element from %s is %s", list, random)
}

// removeFromList removes an item from the list.
func (b *bot) removeFromList(guild, list, arg string) string {
	lis, err := b.getDDB(guild, list)
	if err != nil {
		if err.Code() == listtoErr.ListNotFound {
			return noList(list)
		}
		err.LogError()
		return failMsg
	}

	lis.RemoveItem(arg)

	return fmt.Sprintf("I have removed %s from %s", arg, list)
}

// sortList sorts the list.
func (b *bot) sortList(guild, list, arg string) string {
	sort := strings.ToLower(arg)
	if sort != "name" && sort != "time" {
		return "Sorry! I only sort by \"name\" or \"time\"!"
	}

	lis, err := b.getDDB(guild, list)
	if err != nil {
		if err.Code() == listtoErr.ListNotFound {
			return noList(list)
		}
		err.LogError()
		return failMsg
	}

	lis.Sort(sort)

	if err := b.putDDB(lis); err != nil {
		err.LogError()
		return failMsg
	}

	return fmt.Sprintf("I have sorted %s by %s!", list, arg)
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
		"name": (&dynamodb.AttributeValue{}).SetS(lis),
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
