package consts

import "time"

const (
	JwtBlackListKey = "jwt:blacklist:"
	UserIdKey = "user:id:"
	UserNameKey = "user:name:"
	RdbTTL = time.Hour * 24 * 2
	RdbAddTTL = time.Minute * 10
)