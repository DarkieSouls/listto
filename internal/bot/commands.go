package bot

import (
	"fmt"
)

func addToList(list, arg string) string {
	return fmt.Sprintf("Will eventually work to add %s to list %s", arg, list)
}

func clearList(list string) string {
	return fmt.Sprintf("Will eventually work to clear list %s", list)
}

func createList(list string) string {
	return fmt.Sprintf("Will eventually work to create list %s", list)
}

func deleteList(list string) string {
	return fmt.Sprintf("Will eventually work to delete list %s", list)
}

func help() string {
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

func listLists() string {
	return "Will eventually return a list of lists"
}

func ping() string {
	return "pong"
}

func prefix() string {
	return "Will eventually let you update bot prefix"
}

func createPrivateList(list, arg string) string {
	return fmt.Sprintf("Will eventually work to let you create a private list %s granting access to %s", list, arg)
}

func randomFromList(list string) string {
	return fmt.Sprintf("Will eventually work to select a random entity from list %s", list)
}

func removeFromList(list, arg string) string {
	return fmt.Sprintf("Will eventually work to remove %s from list %s", arg, list)
}

func sortList(list, arg string) string {
	return fmt.Sprintf("Will eventually work to sort list %s", list)
}
