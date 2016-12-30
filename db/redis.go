package db

import (
	"bytes"
	"strconv"
	"tiberious/types"

	"gopkg.in/redis.v5"

	"github.com/Sirupsen/logrus"
	"github.com/pborman/uuid"
	"github.com/pkg/errors"
)

/*
Data map:
	Users:
		Main Key: "user-"+<user-type>+"-"+<loginname>+<uuid> (hash)
		Joined Rooms: "user-"+<user-type+"-"+<uuid>+"rooms" (set)
		Joined Groups: "user-"+<user-type+"-"+<uuid>+"groups" (set)
	Rooms:
		Info: "room-"+<group name>+<room name>+"-info" (hash)
		User List: "room-"+<group name>+<room name>+"-list" (set)
	Groups:
		Info: "group-"+<group name>+"info" (hash)
		User List: "group-"+<group name>+"-users" (set)
		Room List: "group-"+<group name>+"-rooms" (set)
*/

var (
	errMissingDatabaseHost = errors.New("Missing DatabaseHost in config file")
)

type (
	rdisClient interface {
		shutdown()
		updateSet(key string, new []string)
		getKeySet(search string) ([]string, error)
		writeUserData(user *types.User) error
		writeRoomData(room *types.Room) error
		writeGroupData(group *types.Group) error
		getUserData(id string) (*types.User, error)
		getRoomData(gname, rname string) (*types.Room, error)
		getGroupData(gname string) (*types.Group, error)
		deleteUser(user *types.User) error

		Client() *redis.Client
	}

	rClient struct {
		client *redis.Client

		log *logrus.Logger
	}
)

func (db *dbClient) newRedisClient(log *logrus.Logger) (rdisClient, error) {
	if db.config.DatabaseAddress == "" {
		return nil, errMissingDatabaseHost
	}

	r := &rClient{
		log: log,
	}

	if db.config.DatabasePass == "" {
		r.log.Info("Insecure redis database is not recommended")
	}

	r.client = redis.NewClient(&redis.Options{
		Addr:     db.config.DatabaseAddress,
		Password: db.config.DatabasePass,
		DB:       db.config.DatabaseUser,
	})
	// Confirm we can communicate with the redis instance.
	if err := r.client.Ping().Err(); err != nil {
		return nil, errors.Wrap(err, "r.client.Ping")
	}

	return r, nil
}

// Stupid helper function because redis only handles strings.
func strbool(b bool) string {
	if b {
		return "true"
	}

	return "false"
}

func boolstr(s string) bool {
	if s == "true" {
		return true
	}

	return false
}

func (r *rClient) shutdown() {
	var (
		save *redis.StatusCmd
		//quit *redis.StatusCmd
	)
	if _, err := r.client.Pipelined(func(pipe *redis.Pipeline) error {
		save = pipe.Save()
		//quit = pipe.Quit()
		return nil
	}); err != nil {
		r.log.Error(err)
	}

	if err := save.Err(); err != nil {
		r.log.Error(err)
	}
	//if err := quit.Err(); err != nil {
	//	r.log.Error(err)
	//}
}

// This seems extremely cumbersome but it's the best way I can think to handle
// this without deleting the entire set and recreating it.
// This is failing with dynamic tests because apparently redis.Client.Select
// doesn't survive Go routines....
func (r *rClient) updateSet(key string, new []string) {
	var cmd *redis.StringSliceCmd
	if _, err := r.client.Pipelined(func(pipe *redis.Pipeline) error {
		cmd = pipe.SMembers(key)
		return nil
	}); err != nil {
		r.log.Error(err)
	}

	old, err := cmd.Result()
	if err != nil {
		r.log.Error(err)
	}

	for _, o := range old {
		var rem = true
		for _, n := range new {
			if o == n {
				rem = false
			}
		}

		if rem {
			var iCmd *redis.IntCmd
			if _, err := r.client.Pipelined(func(pipe *redis.Pipeline) error {
				iCmd = pipe.SRem(key, o)
				return nil
			}); err != nil {
				r.log.Error(err)
			}
			if err := iCmd.Err(); err != nil {
				r.log.Error(err)
			}
		}
	}

	for _, n := range new {
		var add = true
		for _, o := range old {
			if n == o {
				add = false
			}
		}

		if add {
			var iCmd *redis.IntCmd
			if _, err := r.client.Pipelined(func(pipe *redis.Pipeline) error {
				iCmd = pipe.SAdd(key, n)
				return nil
			}); err != nil {
				r.log.Error(err)
			}
			if err := iCmd.Err(); err != nil {
				r.log.Error(err)
			}
		}
	}
}

func (r *rClient) getKeySet(search string) ([]string, error) {
	var cmd *redis.StringSliceCmd
	if _, err := r.client.Pipelined(func(pipe *redis.Pipeline) error {
		cmd = pipe.Keys(search)
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "r.client.Pipelined Keys")
	}

	slice, err := cmd.Result()
	if err != nil {
		return nil, errors.Wrap(err, "r.client.Pipelined Keys")
	}

	return slice, nil
}

func (r *rClient) writeUserData(user *types.User) error {
	m := map[string]string{
		"id":        user.ID.String(),
		"type":      user.Type,
		"username":  user.Username,
		"loginname": user.LoginName,
		"email":     user.Email,
		"password":  user.Password,
		"salt":      user.Salt,
		"connected": strbool(user.Connected),
	}

	var cmd *redis.StatusCmd
	if _, err := r.client.Pipelined(func(pipe *redis.Pipeline) error {
		cmd = pipe.HMSet("user-"+user.Type+"-"+user.LoginName+"-"+user.ID.String(), m)
		return nil
	}); err != nil {
		return errors.Wrap(err, "r.client.Pipelined HMSet")
	}

	if _, err := cmd.Result(); err != nil {
		return errors.Wrap(err, "r.client.Pipelined HMSet")
	}

	go r.updateSet("user-"+user.Type+"-"+user.ID.String()+"-rooms", user.Rooms)
	go r.updateSet("user-"+user.Type+"-"+user.ID.String()+"-groups", user.Groups)

	return nil
}

func (r *rClient) writeRoomData(room *types.Room) error {
	var (
		cmd *redis.StatusCmd
		m   = map[string]string{
			"title":   room.Title,
			"group":   room.Group,
			"private": strbool(room.Private),
		}
	)
	if _, err := r.client.Pipelined(func(pipe *redis.Pipeline) error {
		cmd = pipe.HMSet("room-"+room.Group+"-"+room.Title+"-info", m)
		return nil
	}); err != nil {
		return errors.Wrap(err, "r.client.Pipelined HMSet")
	}

	if err := cmd.Err(); err != nil {
		return errors.Wrap(err, "r.client.Pipelined HMSet")
	}

	var slice []string
	for _, u := range room.Users {
		slice = append(slice, u.ID.String())
	}

	go r.updateSet("room-"+room.Group+"-"+room.Title+"-list", slice)
	return nil
}

func (r *rClient) writeGroupData(group *types.Group) error {
	var cmd *redis.BoolCmd
	if _, err := r.client.Pipelined(func(pipe *redis.Pipeline) error {
		cmd = pipe.HSet("group-"+group.Title+"-info", "title", group.Title)
		return nil
	}); err != nil {
		return errors.Wrap(err, "r.client.Pipelined HSet")
	}

	if err := cmd.Err(); err != nil {
		return errors.Wrap(err, "r.client.Pipelined HSet")
	}

	var slice []string
	for _, r := range group.Rooms {
		slice = append(slice, r.Title)
	}
	go r.updateSet("group-"+group.Title+"-rooms", slice)

	slice = nil
	for _, u := range group.Users {
		slice = append(slice, u.ID.String())
	}

	go r.updateSet("group-"+group.Title+"-users", slice)

	return nil
}

func (r *rClient) getUserData(id string) (*types.User, error) {
	keys, err := r.getKeySet("user-*-*-" + id)
	if err != nil {
		return nil, errors.Wrap(err, "r.getKeySet")
	}

	if len(keys) == 0 {
		return nil, nil
	}

	var (
		userCmd   *redis.StringStringMapCmd
		roomsCmd  *redis.StringSliceCmd
		groupsCmd *redis.StringSliceCmd
	)

	if _, err = r.client.Pipelined(func(pipe *redis.Pipeline) error {
		userCmd = pipe.HGetAll(keys[0])
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "r.client.Pipelined HGetAllMap")
	}

	info, err := userCmd.Result()
	if err != nil {
		return nil, errors.Wrap(err, "r.client.Pipelined HGetAllMap")
	}

	user := &types.User{
		ID:        uuid.Parse(info["id"]),
		Type:      info["type"],
		Username:  info["username"],
		LoginName: info["loginname"],
		Email:     info["email"],
		Password:  info["password"],
		Salt:      info["salt"],
		Connected: boolstr(info["connected"]),
	}

	if _, err = r.client.Pipelined(func(pipe *redis.Pipeline) error {
		roomsCmd = pipe.SMembers("user-" + user.Type + "-" + user.ID.String() + "-rooms")
		groupsCmd = pipe.SMembers("user-" + user.Type + "-" + user.ID.String() + "-groups")
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "r.client.Pipelined Multi")
	}

	rooms, err := roomsCmd.Result()
	if err != nil {
		return nil, errors.Wrap(err, "r.client.Pipelined SMembers")
	}

	groups, err := groupsCmd.Result()
	if err != nil {
		return nil, errors.Wrap(err, "r.client.Pipelined SMembers")
	}

	if rooms != nil {
		user.Rooms = rooms
	}
	if groups != nil {
		user.Groups = groups
	}

	return user, nil
}

func (r *rClient) getRoomData(gname, rname string) (*types.Room, error) {
	var (
		infoCmd *redis.StringStringMapCmd
		list    *redis.StringSliceCmd
	)

	if _, err := r.client.Pipelined(func(pipe *redis.Pipeline) error {
		infoCmd = pipe.HGetAll("room-" + gname + "-" + rname + "-info")
		list = pipe.SMembers("room-" + gname + "-" + rname + "-list")
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "r.client.Pipelined Multi")
	}

	info, err := infoCmd.Result()
	if err != nil {
		return nil, errors.Wrap(err, "r.client.Pipelined HGetAll")
	}

	users, err := list.Result()
	if err != nil {
		return nil, errors.Wrap(err, "r.client.Pipelined SMembers")
	}

	room := &types.Room{
		Title:   info["title"],
		Group:   info["group"],
		Private: boolstr(info["private"]),
		Users:   make(map[string]*types.User),
	}

	if len(users) > 0 {
		for _, v := range users {
			u, err := r.getUserData(v)
			if err != nil {
				return nil, errors.Wrap(err, "r.getUserData")
			}
			if u == nil {
				return nil, errors.Errorf("No ID found for user %s", v)
			}
			room.Users[u.ID.String()] = u
		}
	}

	return room, nil
}

func (r *rClient) getGroupData(gname string) (*types.Group, error) {
	group := &types.Group{
		Title: gname,
		Rooms: make(map[string]*types.Room),
		Users: make(map[string]*types.User),
	}

	var (
		usersCmd *redis.StringSliceCmd
		roomCmd  *redis.StringSliceCmd
	)
	if _, err := r.client.Pipelined(func(pipe *redis.Pipeline) error {
		usersCmd = pipe.SMembers("group-" + gname + "-users")
		roomCmd = pipe.SMembers("group-" + gname + "-rooms")
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "r.client.Pipelined Multi")
	}

	users, err := usersCmd.Result()
	if err != nil {
		return nil, errors.Wrap(err, "r.client.Pipelined SMembers")
	}

	rooms, err := roomCmd.Result()
	if err != nil {
		return nil, errors.Wrap(err, "r.client.Pipelined SMembers")
	}

	if len(users) > 0 {
		for _, v := range users {
			/* For some reason the length of this keeps coming up as 1 above
			 * the actual number of entries so confirm it's not nil before
			 * attempting to run GetUserData with the given string. */
			if v != "" {
				var u *types.User
				u, err = r.getUserData(v)
				if err != nil {
					return nil, errors.Wrap(err, "r.getUserData")
				}
				if u == nil {
					return nil, errors.Errorf("No ID found for user %s", v)
				}
				group.Users[u.ID.String()] = u
			}
		}
	}

	if len(rooms) > 0 {
		for _, v := range rooms {
			if v != "" {
				r, err := r.getRoomData(gname, v)
				if err != nil {
					return nil, errors.Wrap(err, "r.getRoomData")
				}
				group.Rooms[r.Title] = r
			}
		}
	}

	return group, nil
}

func (r *rClient) deleteUser(user *types.User) error {
	var (
		del0 *redis.IntCmd
		del1 *redis.IntCmd
		del2 *redis.IntCmd
	)
	if _, err := r.client.Pipelined(func(pipe *redis.Pipeline) error {
		del0 = pipe.Del("user-" + user.Type + "-" + user.ID.String() + "-groups")
		del1 = pipe.Del("user-" + user.Type + "-" + user.ID.String() + "-rooms")
		del2 = pipe.Del("user-" + user.Type + "-" + user.LoginName + "-" + user.ID.String())
		return nil
	}); err != nil {
		return errors.Wrap(err, "r.client.Pipelined Multi")
	}

	if err := del0.Err(); err != nil {
		return errors.Wrap(err, "r.pipe.Del")
	}
	if err := del1.Err(); err != nil {
		return errors.Wrap(err, "r.pipe.Del")
	}
	if err := del2.Err(); err != nil {
		return errors.Wrap(err, "r.pipe.Del")
	}

	return nil
}

// Client returns the underlyins *redis.Client for testing purposes.
func (r *rClient) Client() *redis.Client {
	return r.client
}

// GenRedisProto is an external function that can be used to generate Redis
// protocol text for mass insertion via CLI.
func GenRedisProto(cmd []string) string {
	var ret string
	ret = "*" + strconv.Itoa(len(cmd)) + "\r\n"
	for _, v := range cmd {
		ret += "$" + strconv.FormatInt(bytes.NewReader([]byte(v)).Size(), 10) + "\r\n"
		ret += v + "\r\n"
	}

	return ret
}
