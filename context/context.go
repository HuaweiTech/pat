package context

import (
	"encoding/json"
	"io/ioutil"
	"strconv"
)

type Context interface {
	PutString(k string, v string)
	GetString(k string) (string, bool)
	PutInt(k string, v int)
	GetInt(k string) (int, bool)
	PutFloat64(k string, v float64)
	GetFloat64(k string) (float64, bool)
	PutBool(k string, v bool)
	GetBool(k string) (bool, bool)
	MarshalJSON() ([]byte, error)
	UnmarshalJSON(b []byte) error
	Clone() Context
	SetUsers(userpath string) error
}

type user []string
type contextMap struct {
	c     map[string]interface{}
	users []user
	index int
}

func New() Context {
	var c = contextMap{
		c:     make(map[string]interface{}),
		users: []user{},
		index: 0,
	}
	return &c
}

func (c *contextMap) PutString(k string, v string) {
	c.c[k] = v
}

func (c *contextMap) GetString(k string) (string, bool) {
	if c.c[k] == nil {
		return "", false
	} else {
		return c.c[k].(string), true
	}
}

func (c *contextMap) PutInt(k string, v int) {
	// saves as string, avoids json.Marshal turning int into json.numbers(float64)
	c.c[k] = strconv.Itoa(v)
}

func (c *contextMap) GetInt(k string) (int, bool) {
	if c.c[k] == nil {
		return 0, false
	} else {
		value, _ := strconv.ParseInt(c.c[k].(string), 0, 0)
		return int(value), true
	}
}

func (c *contextMap) PutFloat64(k string, v float64) {
	c.c[k] = v
}

func (c *contextMap) GetFloat64(k string) (float64, bool) {
	if c.c[k] == nil {
		return 0, false
	} else {
		return c.c[k].(float64), true
	}
}

func (c *contextMap) PutBool(k string, v bool) {
	c.c[k] = v
}

func (c *contextMap) GetBool(k string) (bool, bool) {
	if c.c[k] == nil {
		return false, false
	} else {
		return c.c[k].(bool), true
	}
}

func (c *contextMap) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}(c.c))
}

func (c *contextMap) UnmarshalJSON(b []byte) error {
	var m map[string]interface{} = c.c
	return json.Unmarshal(b, &m)
}

func (c *contextMap) Clone() Context {
	var clone = contextMap{
		c: make(map[string]interface{}),
	}
	for k, v := range c.c {
		clone.c[k] = v
	}

	users := c.users
	userslen := len(users)
	if userslen > 0 {
		curindex := c.index % userslen
		clone.PutString("rest:username", users[curindex][0])
		clone.PutString("rest:password", users[curindex][1])
		c.index++
	}
	return &clone
}

func ParseUser(path string) ([]user, error) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var users []user
	err = json.Unmarshal(file, &users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (c *contextMap) SetUsers(userpath string) error {
	users, err := ParseUser(userpath)
	c.users = users
	return err
}
