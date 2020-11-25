package session

import (
	"testing"
	"time"
)

type FlashMessage struct {
	Type    int
	Message string
}

func TestFlashes(t *testing.T) {
	var flashes []interface{}

	session := &Session{
		values: make(map[interface{}]interface{}),
		name:   "session-key",
		isNew:  true,
		dirty:  false,
	}

	// Get a flash.
	flashes = session.Flashes()
	if len(flashes) != 0 {
		t.Errorf("Expected empty flashes; Got %v", flashes)
	}
	// Add some flashes.
	session.AddFlash("foo")
	session.AddFlash("bar")
	// Custom key.
	session.AddFlash("baz", "custom_key")

	// Check all saved values.
	flashes = session.Flashes()
	if len(flashes) != 2 {
		t.Fatalf("Expected flashes; Got %v", flashes)
	}
	if flashes[0] != "foo" || flashes[1] != "bar" {
		t.Errorf("Expected foo,bar; Got %v", flashes)
	}
	flashes = session.Flashes()
	if len(flashes) != 0 {
		t.Errorf("Expected dumped flashes; Got %v", flashes)
	}
	// Custom key.
	flashes = session.Flashes("custom_key")
	if len(flashes) != 1 {
		t.Errorf("Expected flashes; Got %v", flashes)
	} else if flashes[0] != "baz" {
		t.Errorf("Expected baz; Got %v", flashes)
	}
	flashes = session.Flashes("custom_key")
	if len(flashes) != 0 {
		t.Errorf("Expected dumped flashes; Got %v", flashes)
	}

	// Get a flash.
	flashes = session.Flashes()
	if len(flashes) != 0 {
		t.Errorf("Expected empty flashes; Got %v", flashes)
	}
	// Add some flashes.
	session.AddFlash(&FlashMessage{42, "foo"})

	// Check all saved values.
	flashes = session.Flashes()
	if len(flashes) != 1 {
		t.Fatalf("Expected flashes; Got %v", flashes)
	}
	custom := flashes[0].(*FlashMessage)
	if custom.Type != 42 || custom.Message != "foo" {
		t.Errorf("Expected %#v, got %#v", FlashMessage{42, "foo"}, custom)
	}
}

func TestExpiration(t *testing.T) {
	s := &Session{
		values: make(map[interface{}]interface{}),
		name:   "session-key",
		options: &Options{
			IdleTimeout: 900 * time.Second,
			AbsoluteTimeout: 1800 * time.Second,
		},
		isNew: true,
		dirty: false,
	}

	//test scenario for a just created session
	s.lastAccessed = time.Now()
	s.values[createdTimeKey] = time.Now()

	if s.isExpired() {
		t.Errorf("just created session should not be expired")
	}

	//test scenario for a idle timeout
	s.lastAccessed = time.Now().Add(-901 * time.Second)
	s.values[createdTimeKey] = time.Now().Add(-901 * time.Second)

	if !s.isExpired() {
		t.Errorf("last accessed at %v, with idle timeout %v should be expired", s.lastAccessed, s.options.IdleTimeout)
	}

	//test scenario for a absolute timeout
	s.lastAccessed = time.Now().Add(-450 * time.Second)
	s.values[createdTimeKey] = time.Now().Add(-1801 * time.Second)

	if !s.isExpired() {
		t.Errorf("created at %v, with abs timeout %v should be expired", s.values[createdTimeKey], s.options.AbsoluteTimeout)
	}

	//test scenario for a not expired session
	s.lastAccessed = time.Now().Add(-450 * time.Second)
	s.values[createdTimeKey] = time.Now().Add(-1700 * time.Second)

	if s.isExpired() {
		t.Errorf("created at %v, last accessed at %v should be valid", s.values[createdTimeKey], s.lastAccessed)
	}
}