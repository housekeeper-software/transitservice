package push

import "encoding/json"

//https://doc.yunxin.163.com/messaging/docs/jYxMjQ1NTk?platform=server

//自定义系统通知的具体内容，开发者组装的字符串，建议 JSON 格式，最大长度 4096 字符
type Attach struct {
	Type            string          `json:"type"`
	IntercomContent IntercomMessage `json:"intercomContent"`
}

type Alert struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

//https://faq.yunxin.163.com/kb/main/#/item/0198
type ApsField struct {
	MutableContent int    `json:"mutable-content"`
	Sound          string `json:"sound"`
	AlertInfo      Alert  `json:"alert"`
}

//https://faq.yunxin.163.com/kb/main/#/item/KB0291
type Payload struct {
	PushTitle string   `json:"pushTitle"`
	Aps       ApsField `json:"apsField"`
}

func CreateAttach(msgTag string, message *IntercomMessage) ([]byte, error) {
	attach := Attach{
		Type:            msgTag,
		IntercomContent: *message, //mobile use this value
	}
	data, err := json.Marshal(attach)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func CreatePayload(message *IntercomMessage) ([]byte, error) {
	payload := Payload{
		PushTitle: message.Title,
		Aps: ApsField{
			MutableContent: 1,
			Sound:          "default",
			AlertInfo: Alert{
				Title: message.Title,
				Body:  message.Body,
			},
		},
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return data, nil
}
