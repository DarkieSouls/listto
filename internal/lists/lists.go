package lists

import (
	"math/rand"
	"sort"
	"strings"
	"time"
)

// ListtoList defines the list object that holds all needed data for each list
type ListtoList struct {
	Guild   string     `json:"guild"`
	Name    string     `json:"name"`
	Private bool       `json:"private"`
	Access  []string   `json:"access"`
	List    []ListItem `json:"list"`
}

// ListItem represents a single value in a list.
type ListItem struct {
	Value     string `json:"value"`
	TimeAdded int64  `json:"timeAdded"`
}

// NewList returns a new ListtoList object.
func NewList(guild, name string, private bool) *ListtoList {
	return &ListtoList{
		Guild:   guild,
		Name:    name,
		Private: private,
	}
}

// AddItem to a ListtoList.
func (l *ListtoList) AddItem(item string, timeAdded int64) {
	l.List = append(l.List, ListItem{Value: item, TimeAdded: timeAdded})
}

// RemoveItem from a ListtoList.
func (l *ListtoList) RemoveItem(item string) {
	for i, v := range l.List {
		if v.Value == item {
			l.List = append(l.List[:i], l.List[i+1:]...)
			break
		}
	}
}

// Clear a ListtoList of all Items.
func (l *ListtoList) Clear() {
	l.List = make([]ListItem, 0)
}

// SelectRandom Item from a ListtoList.
func (l *ListtoList) SelectRandom() string {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	i := r.Intn(len(l.List))
	return l.List[i].Value
}

// Sort a ListtoList by a value.
func (l *ListtoList) Sort(sorter string) {
	if sorter == "name" {
		sort.Slice(l.List, func(i, j int) bool {
			return strings.ToLower(l.List[i].Value) < strings.ToLower(l.List[j].Value)
		})
	} else if sorter == "time" {
		sort.Slice(l.List, func(i, j int) bool {
			return l.List[i].TimeAdded < l.List[j].TimeAdded
		})
	}
}

// AddAccess to certain perties to a private ListtoList.
func (l *ListtoList) AddAccess(access []string) {
	if !l.Private {
		return
	}

	for _, a := range access {
		var dupe bool
		for _, v := range l.Access {
			if a == v {
				dupe = true
				break
			}
		}
		if !dupe {
			l.Access = append(l.Access, a)
		}
	}
}

// CanAccess returns if the caller can access the ListtoList.
func (l *ListtoList) CanAccess(user string, roles []string) bool {
	if !l.Private {
		return true
	}

	for _, a := range l.Access {
		if a == user {
			return true
		}

		for _, r := range roles {
			if a == r {
				return true
			}
		}
	}

	return false
}
