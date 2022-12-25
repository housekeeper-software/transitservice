package push

/*

void MakePostData(const User &user,
                  const protocol::ProbufMessage *message,
                  std::string *output) {
  base::DictionaryValue root;
  root.SetString("device", user.push_id);
  root.SetInteger("platform", user.platform);
  root.SetString("cid", user.cid);
  root.SetString("fid", user.fid);
  root.SetString("scheme", message->scheme());
  root.SetString("cmd", message->cmd());
  root.SetDouble("time", base::Time::Now().ToJsTime());

  bool ack = false;
  int result = 0;
  {
    scoped_ptr<base::Value> value(base::JSONReader::Read(message->payload()));
    if (value) {
      const base::DictionaryValue *tmp = nullptr;
      if (value->GetAsDictionary(&tmp) && tmp) {
        tmp->GetBoolean("ack", &ack);
        tmp->GetInteger("result", &result);
      }
    }
  }
  std::string body = message->cmd();

  if (base::CompareCaseInsensitiveASCII(message->scheme(), "icom") == 0
      && base::CompareCaseInsensitiveASCII(message->cmd(), "call") == 0) {
    if (!ack) {
      body = "indoorcall";
    } else {
      if (result) {
        body = "callout_failed";
      } else {
        body = "callout_success";
      }
    }
  }

  if (base::CompareCaseInsensitiveASCII(body, "switch") == 0) {
    body = "security_switch";
  }

  root.SetString("title", message->scheme());
  root.SetString("body", body);
  root.SetBoolean("ack", ack);
  root.SetInteger("result", result);

  std::string str;
  message->SerializeToString(&str);
  base::Base64Encode(str, &str);
  root.SetString("message", str);

  base::JSONWriter::Write(root, output);
}
*/

type IntercomMessage struct {
	Device     string  `json:"device"` // device accid which push to
	Platform   int     `json:"platform"`
	Cid        string  `json:"cid"`
	Fid        string  `json:"fid"`
	Scheme     string  `json:"scheme"`
	Cmd        string  `json:"cmd"`
	CreateTime float64 `json:"time"`
	Title      string  `json:"title"`
	Body       string  `json:"body"`
	Ack        bool    `json:"ack"`
	Result     int     `json:"result"`
	Message    string  `json:"message"` //base64 actually intercom message
}
