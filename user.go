package irckit

import (
	"net"
	"strings"
	"sync"

	"github.com/sorcix/irc"
)

// NewUser creates a *User, wrapping a connection with metadata we need for our server.
func NewUser(c Conn) *User {
	return &User{
		Conn:     c,
		Host:     "*",
		Channels: map[Channel]struct{}{},
	}
}

// NewUserNet creates a *User from a net.Conn connection.
func NewUserNet(c net.Conn) *User {
	return NewUser(&conn{
		Conn:    c,
		Encoder: irc.NewEncoder(c),
		Decoder: irc.NewDecoder(c),
	})
}

type User struct {
	Conn

	sync.RWMutex
	Nick string // From NICK command
	User string // From USER command
	Real string // From USER command
	Host string

	Channels map[Channel]struct{}
}

func (u *User) ID() string {
	return strings.ToLower(u.Nick)
}

func (u *User) Prefix() *irc.Prefix {
	return &irc.Prefix{
		Name: u.Nick,
		User: u.User,
		Host: u.Host,
	}
}

func (u *User) String() string {
	return u.Prefix().String()
}

func (u *User) VisibleTo() []*User {
	seen := map[*User]struct{}{}
	seen[u] = struct{}{}

	num := 0
	for ch := range u.Channels {
		// Don't include self
		num += ch.Len()
	}

	// Pre-allocate
	users := make([]*User, 0, num)
	if num == 0 {
		return users
	}

	// Get all unique users
	for ch := range u.Channels {
		for _, other := range ch.Users() {
			if _, dupe := seen[other]; dupe {
				continue
			}
			seen[other] = struct{}{}
			// TODO: Check visibility (once it's implemented)
			users = append(users, other)
		}
	}
	return users
}

// Encode and send each msg until an error occurs, then returns.
func (user *User) Encode(msgs ...*irc.Message) (err error) {
	for _, msg := range msgs {
		logger.Debugf("-> %s", msg)
		err := user.Conn.Encode(msg)
		if err != nil {
			return err
		}
	}
	return nil
}

// Decode will receive and return a decoded message, or an error.
func (user *User) Decode() (*irc.Message, error) {
	msg, err := user.Conn.Decode()
	logger.Debugf("<- %s", msg)
	return msg, err
}