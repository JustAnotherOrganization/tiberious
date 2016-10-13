package main

/* Generate redis insert text for mass insertion in redis using cli.
 * go run redis_helper.go | redis-cli --pipe
 */

import (
	"fmt"

	"tiberious/db"
)

func main() {
	fmt.Print(db.GenRedisProto([]string{"SELECT", "1"}))
	fmt.Print(db.GenRedisProto([]string{"HMSET", "user-admin-IngCr3at1on-7e03d41c-9219-4fa5-810a-d7fd6e2f39de", "id", "7e03d41c-9219-4fa5-810a-d7fd6e2f39de", "type", "admin", "username", "IngCr3at1on", "loginname", "IngCr3at1on", "email", "nathan@projectopencannibal.org", "password", "wooga wooga wooga", "salt", "salty", "connected", "false"}))
	fmt.Print(db.GenRedisProto([]string{"HMSET", "user-admin-Grim-8812886f-6fc8-4b55-abf3-1607294dd31f", "id", "8812886f-6fc8-4b55-abf3-1607294dd31f", "type", "admin", "username", "grim", "loginname", "grim", "email", "grim@nowhere.nowhere", "password", "whimmy wham wham wozzle", "salt", "extra salty", "connected", "false"}))
}
