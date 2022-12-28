package push

import (
	"encoding/json"
	"io/ioutil"
	"strings"
)

type Pair struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Keyword struct {
	DefaultTitle string `json:"defaultTitle"`
	DefaultBody  string `json:"defaultBody"`
	Titles       []Pair `json:"title"`
	Bodys        []Pair `json:"body"`
	titleMap     map[string]string
	bodyMap      map[string]string
}

func LoadKeyword(file string) (*Keyword, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var c Keyword
	err = json.Unmarshal(data, &c)
	if err != nil {
		return nil, err
	}
	c.titleMap = make(map[string]string)
	c.bodyMap = make(map[string]string)
	for _, v := range c.Titles {
		c.titleMap[strings.ToLower(v.Name)] = v.Value
	}
	for _, v := range c.Bodys {
		c.bodyMap[strings.ToLower(v.Name)] = v.Value
	}
	return &c, nil
}

func (c *Keyword) Replace(message *IntercomMessage) {
	lowTitle := strings.ToLower(message.Title)
	lowBody := strings.ToLower(message.Body)

	if v, ok := c.titleMap[lowTitle]; ok {
		message.Title = v
	} else {
		message.Title = c.DefaultTitle
	}
	if v, ok := c.bodyMap[lowBody]; ok {
		message.Body = v
	} else {
		message.Body = c.DefaultBody
	}
}
