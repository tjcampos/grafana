package middleware

import (
	"time"

	"github.com/Unknwon/macaron"
	"github.com/macaron-contrib/session"
	_ "github.com/macaron-contrib/session/mysql"
	_ "github.com/macaron-contrib/session/postgres"
	_ "github.com/macaron-contrib/session/redis"
)

const (
	SESS_KEY_USERID    = "uid"
	SESS_KEY_FAVORITES = "favorites"
)

var sessionManager *session.Manager
var sessionOptions session.Options

func startSessionGC() {
	sessionManager.GC()
	time.AfterFunc(time.Duration(sessionOptions.Gclifetime)*time.Second, startSessionGC)
}

func Sessioner(options session.Options) macaron.Handler {
	var err error
	sessionOptions = options
	sessionManager, err = session.NewManager(options.Provider, options)
	if err != nil {
		panic(err)
	}

	go startSessionGC()

	return func(ctx *Context) {
		ctx.Next()

		if err = ctx.Session.Release(); err != nil {
			panic("session(release): " + err.Error())
		}
	}
}

func GetSession() SessionStore {
	return &SessionWrapper{manager: sessionManager}
}

type SessionStore interface {
	// Set sets value to given key in session.
	Set(interface{}, interface{}) error
	// Get gets value by given key in session.
	Get(interface{}) interface{}
	// ID returns current session ID.
	ID() string
	// Release releases session resource and save data to provider.
	Release() error
	// Destory deletes a session.
	Destory(*Context) error
	// init
	Start(*Context) error
}

type SessionWrapper struct {
	session session.RawStore
	manager *session.Manager
}

func (s *SessionWrapper) Start(c *Context) error {
	var err error
	s.session, err = s.manager.Start(c.Context)
	return err
}

func (s *SessionWrapper) Set(k interface{}, v interface{}) error {
	if s.session != nil {
		return s.session.Set(k, v)
	}
	return nil
}

func (s *SessionWrapper) Get(k interface{}) interface{} {
	if s.session != nil {
		return s.session.Get(k)
	}
	return nil
}

func (s *SessionWrapper) ID() string {
	if s.session != nil {
		return s.session.ID()
	}
	return ""
}

func (s *SessionWrapper) Release() error {
	if s.session != nil {
		return s.session.Release()
	}
	return nil
}

func (s *SessionWrapper) Destory(c *Context) error {
	if s.session != nil {
		return s.manager.Destory(c.Context)
	}
	return nil
}
