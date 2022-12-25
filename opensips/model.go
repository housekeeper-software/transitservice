package opensips

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
)

/*
std::string SipRequestParam::CreateRequestData() {
  Json::Value root;
  root["user"] = auth_user_;
  root["pwd"] = auth_pwd_;
  root["did"] = device_id_;
  Json::Value v;
  net_client_.ToSimplifyJson(v);
  root["client"] = v;

  Json::FastWriter writer;
  std::string post_data = writer.write(root);
  post_data = ("data=" + post_data);
  return post_data;
}

void NetClient::ToSimplifyJson(Json::Value &v) const {
  v["c"] = client_id_;
  v["f"] = family_id_;
  v["t"] = type_;
  v["s"] = sub_type_;
  v["b"] = button_key_;
  v["a"] = alias_name_;
  v["p"] = platform_;
  v["v"] = version_;
  v["n"] = sn_;
  v["m"] = number_;
}
*/
//Intercom device Request
type UserRequest struct {
	User   string `json:"user"`
	Pwd    string `json:"pwd"`
	Did    string `json:"did"`
	Client struct {
		ClientId     string `json:"c"`
		FamilyId     string `json:"f"`
		Type         int    `json:"t"`
		SubType      int    `json:"s"`
		ButtonKey    string `json:"b"`
		AliasName    string `json:"a"`
		Platform     int    `json:"p"`
		Version      int    `json:"v"`
		SerialNumber string `json:"n"`
		Number       string `json:"m"`
	} `json:"client"`
}

type SipAuth struct {
	Username string `json:"username"`
	Userid   string `json:"userid"`
	Passwd   string `json:"passwd"`
}

type SipProxy struct {
	Proxy    string `json:"proxy"`
	Identity string `json:"identity"`
	Expires  int    `json:"expires"`
}

type IceConf struct {
	TurnServer     string `json:"turn_server"`
	TurnUsername   string `json:"turn_username"`
	TurnPwd        string `json:"turn_pwd"`
	TurnAuthType   string `json:"turn_auth_type"`
	TurnPwdType    string `json:"turn_pwd_type"`
	TurnConnType   string `json:"turn_conn_type"`
	TurnAuthrealm  string `json:"turn_auth_realm"`
	StunServer     string `json:"stun_server"`
	EnableLoopaddr bool   `json:"enable_loopaddr"`
}

type SipConf struct {
	SessionExpires                int        `json:"session_expires"`
	UseRport                      bool       `json:"use_rport"`
	ReuseAuthorization            bool       `json:"reuse_authorization"`
	ExpireOldRegistrationContacts bool       `json:"expire_old_registration_contacts"`
	Transport                     string     `json:"transport"`
	AddDates                      bool       `json:"add_dates"`
	UserAgent                     string     `json:"user_agent"`
	MaxCalls                      int        `json:"max_calls"`
	Auth                          []SipAuth  `json:"auth"`
	Proxy                         []SipProxy `json:"proxy"`
}

type SipIceConfig struct {
	Sip SipConf `json:"sip"`
	Ice IceConf `json:"ice"`
}

func LoadSipIceTemplate(file string) (*SipIceConfig, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var c SipIceConfig
	err = json.Unmarshal(data, &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *SipIceConfig) ReplaceStunServer(server string) {
	s.Ice.StunServer = server
}

func (s *SipIceConfig) ReplaceSipServer(server string) {
	for k, _ := range s.Sip.Proxy {
		//we only replace once ,so we use fmt.Sprintf,it's low effective
		//<sip:server> is sip proxy formatter
		s.Sip.Proxy[k].Proxy = fmt.Sprintf("<sip:%s>", server)
	}
}

func (s *SipIceConfig) ReplaceUser(server string, username string, password string) {
	for k, _ := range s.Sip.Auth {
		s.Sip.Auth[k].Username = username
		s.Sip.Auth[k].Userid = username
		s.Sip.Auth[k].Passwd = password
	}
	for k, _ := range s.Sip.Proxy {
		//<sip:username@ip:port>
		var builder strings.Builder
		builder.WriteString("<sip:")
		builder.WriteString(username)
		builder.WriteByte('@')
		builder.WriteString(server)
		builder.WriteByte('>')
		s.Sip.Proxy[k].Identity = builder.String()
	}
}
