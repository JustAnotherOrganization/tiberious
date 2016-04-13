package handlers

import (
	"testing"

	"tiberious/types"
)

func TestGetGroup(t *testing.T) {
	var g *types.Group
	// No database is enabled, should return nil
	g = GetGroup("#test")
	if g != nil {
		t.Fail()
	}

	// #default exists with or with out a database
	g = GetGroup("#default")
	if g == nil {
		t.Fail()
	}

	if g.Title != "#default" {
		t.Fail()
	}
}

func TestGetNewGroup(t *testing.T) {
	var g *types.Group
	// No database is enabled, should return nil
	g = GetNewGroup("#anything", false)
	if g != nil {
		t.Fail()
	}

	g = GetNewGroup("#test", true)
	if g == nil {
		t.Fail()
	}

	if g.Title != "#test" {
		t.Fail()
	}
}

func TestGetRoom(t *testing.T) {
	var r *types.Room
	/* No database is enabled so only the #default and #test groups exist
	 * test having been created during the previous test. */
	r = GetRoom("#anything", "#anywhere")
	if r != nil {
		t.Fail()
	}

	// Only the room #general exists current...
	r = GetRoom("#default", "#anywhere")
	if r != nil {
		t.Fail()
	}

	r = GetRoom("#default", "#general")
	if r == nil {
		t.Fail()
	}

	if r.Title != "#general" {
		t.Fail()
	}
}

func TestGetNewRoom(t *testing.T) {
	var r *types.Room
	// Only the #general and #test groups exist.
	r = GetNewRoom("#anything", "#anywhere")
	if r != nil {
		t.Fail()
	}

	r = GetNewRoom("#default", "#anywhere")
	if r == nil {
		t.Fail()
	}

	if r.Title != "#anywhere" {
		t.Fail()
	}
}
