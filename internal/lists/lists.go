package lists

import (
	"math/rand"
	"time"
)

// ListtoList defines the list object that holds all needed data for each list
type ListtoList struct {
	Guild string `json:"guild"`
	Name string `json:"name"`
	Private bool `json:"private"`
	List []ListItem `json:"list"`
}

// ListItem represents a single value in a list.
type ListItem struct {
	Value string `json:"value"`
	TimeAdded time.Time `json:"timeAdded"`
}

func NewList(guild, name string, private bool) *ListtoList {
	return &ListtoList{
		Guild: guild,
		Name: name,
		Private: private,
	}
}

func (l *ListtoList) AddItem(item string, timeAdded time.Time) {
	l.List = append(l.List, ListItem{Value: item, TimeAdded: timeAdded})
}

func (l *ListtoList) RemoveItem(item string) {
	for i, v := range l.List {
		if v.Value == item {
			l.List = append(l.List[:i], l.List[i+1:]...)
			break
		}
	}
}

func (l *ListtoList) Clear() {
	l.List = make([]ListItem, 0)
}

func (l *ListtoList) SelectRandom() string {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	i := r.Intn(len(l.List))
	return l.List[i].Value
}
