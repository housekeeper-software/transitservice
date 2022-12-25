package opensips

import (
	"jingxi.cn/transitservice/utils"
	"strings"
)

/*
DROP TABLE IF EXISTS `subscriber`;
CREATE TABLE `subscriber`  (
  `id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT,
  `username` char(64) CHARACTER SET latin1 COLLATE latin1_swedish_ci NOT NULL DEFAULT '',
  `domain` char(64) CHARACTER SET latin1 COLLATE latin1_swedish_ci NOT NULL DEFAULT '',
  `password` char(32) CHARACTER SET latin1 COLLATE latin1_swedish_ci NOT NULL DEFAULT '',
  `email_address` char(64) CHARACTER SET latin1 COLLATE latin1_swedish_ci NOT NULL DEFAULT '',
  `ha1` char(64) CHARACTER SET latin1 COLLATE latin1_swedish_ci NOT NULL DEFAULT '',
  `ha1b` char(64) CHARACTER SET latin1 COLLATE latin1_swedish_ci NOT NULL DEFAULT '',
  `rpid` char(64) CHARACTER SET latin1 COLLATE latin1_swedish_ci NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `account_idx`(`username`, `domain`) USING BTREE,
  INDEX `username_idx`(`username`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 90 CHARACTER SET = latin1 COLLATE = latin1_swedish_ci ROW_FORMAT = Dynamic;

SET FOREIGN_KEY_CHECKS = 1;
*/

//opensips User object
type User struct {
	Username     string
	Domain       string
	Password     string
	EmailAddress string
	Ha1          string
	Ha1b         string
	Rpid         string
}

//make a new User object from domain,username, password
func NewUser(domain string, user string, password string) *User {
	newUser := &User{
		Username:     user,
		Domain:       domain,
		Password:     password,
		EmailAddress: "",
		Ha1:          "",
		Ha1b:         "",
		Rpid:         "",
	}
	if len(password) < 1 {
		//we use long password
		newUser.Password = utils.RandString(16)
	}
	newUser.Ha1 = GetHa1(newUser)
	newUser.Ha1b = GetHa1b(newUser)
	return newUser
}

func (u *User) SetPassword(password string) {
	u.Password = password
	u.Ha1 = GetHa1(u)
	u.Ha1b = GetHa1b(u)
}

//md5('did_cid_sn')
func CreateUserId(did string, cid string, sn string) string {
	var builder strings.Builder
	builder.WriteString(did)
	builder.WriteByte('_')
	builder.WriteString(cid)
	builder.WriteByte('_')
	builder.WriteString(sn)
	return utils.Md5String(builder.String())
}

//md5('username:domain:password')
func GetHa1(user *User) string {
	var builder strings.Builder
	builder.WriteString(user.Username)
	builder.WriteByte(':')
	builder.WriteString(user.Domain)
	builder.WriteByte(':')
	builder.WriteString(user.Password)
	return utils.Md5String(builder.String())
}

//md5('username@domain:domain:password')
func GetHa1b(user *User) string {
	var builder strings.Builder
	builder.WriteString(user.Username)
	builder.WriteByte('@')
	builder.WriteString(user.Domain)
	builder.WriteByte(':')
	builder.WriteString(user.Domain)
	builder.WriteByte(':')
	builder.WriteString(user.Password)
	return utils.Md5String(builder.String())
}

//check user in database is valid
func IsUserValid(user *User, domain string) bool {
	if len(user.Password) < 1 {
		return false
	}
	if !strings.EqualFold(user.Domain, domain) {
		return false
	}

	ha1 := GetHa1(user)
	hab1 := GetHa1b(user)
	if !strings.EqualFold(user.Ha1, ha1) {
		return false
	}
	if !strings.EqualFold(user.Ha1b, hab1) {
		return false
	}
	return true
}
