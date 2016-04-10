package types

type Group struct {
	Title string
	// List all the rooms in a group
	Rooms map[string]*Room
	// List all the users in a group
	Users map[string]*User
}

/*
func (group Group) IsUserMember(id string) (bool, error) {
	groups, err := db.GetUserGroups(id)
	/* This should never happen cause it should never be called without a
	 * database being active. *
	if err != nil {
		return false, err
	}

	for _, g := range groups {
		if g == group.Title {
			return true, nil
		}
	}

	return false, nil
}
*/
