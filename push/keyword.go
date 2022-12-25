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
	return &c, nil
}

func (c *Keyword) Replace(message *IntercomMessage) {
	found := false
	for _, v := range c.Titles {
		if strings.EqualFold(v.Name, message.Title) {
			message.Title = v.Value
			found = true
			break
		}
	}
	if !found {
		message.Title = c.DefaultTitle
	}
	found = false
	for _, v := range c.Bodys {
		if strings.EqualFold(v.Name, message.Body) {
			message.Body = v.Value
			found = true
			break
		}
	}
	if !found {
		message.Body = c.DefaultBody
	}
}
