package utils

import "app/model"

func SetupUsers() map[string]model.UserInfo {
	users := map[string]model.UserInfo{
		"u1":  {UserId: "u1", Organization: "org1", Admin: false},
		"u2":  {UserId: "u2", Organization: "org2", Admin: false},
		"u3":  {UserId: "u3", Organization: "org3", Admin: false},
		"u4":  {UserId: "u4", Organization: "org4", Admin: false},
		"u5":  {UserId: "u5", Organization: "org1", Admin: false},
		"u6":  {UserId: "u6", Organization: "org2", Admin: false},
		"u7":  {UserId: "u7", Organization: "org3", Admin: false},
		"u8":  {UserId: "u8", Organization: "org4", Admin: false},
		"u9":  {UserId: "u9", Organization: "org1", Admin: false},
		"u10": {UserId: "u10", Organization: "org2", Admin: false},
		"u11": {UserId: "u11", Organization: "org3", Admin: false},
		"u12": {UserId: "u12", Organization: "org4", Admin: false},
		//Admins
		"s1": {UserId: "s1", Organization: "org1", Admin: true},
		"s2": {UserId: "s2", Organization: "org2", Admin: true},
		"s3": {UserId: "s3", Organization: "org3", Admin: true},
		"s4": {UserId: "s4", Organization: "org4", Admin: true},
	}
	return users
}
