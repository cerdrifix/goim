package middleware

import (
	"log"
	"strings"
)

type Session struct {
	Email    string
	Password string
	Name     string
	Data     map[string]interface{}
}

type MockClient struct {
	Session *Session
}

type MockDB struct {
	Session *Session
}

type SessionHandler interface {
	Apply(*Session) *Session
}

type SessionHandlerFunc func(*Session) *Session

func (f SessionHandlerFunc) Apply(s *Session) *Session {
	return f(s)
}

func getUserSession(s *Session) *Session {
	log.Println("Enter getUserSession")
	return s
}

func checkEmail(next SessionHandler) SessionHandler {
	log.Println("Enter checkEmail")
	return SessionHandlerFunc(func(s *Session) *Session {
		ss := next.Apply(s)

		// Screen email address
		if !strings.Contains(ss.Email, "@") {
			log.Panicln("Email missing '@'")
		}
		return ss
	})
}

func checkPassword(next SessionHandler) SessionHandler {
	log.Println("Enter checkPassword")

	return SessionHandlerFunc(func(s *Session) *Session {
		ss := next.Apply(s)

		// Screen password's length
		if len(ss.Password) < 6 {
			log.Panicln("Password too short!")
		}
		return ss
	})
}

func (db *MockDB) Query(next SessionHandler) SessionHandlerFunc {
	return SessionHandlerFunc(func(s *Session) *Session {
		ss := next.Apply(s)

		// Match email and password to MockDB's
		if ss.Email != db.Session.Email {
			log.Panicln("Email not matched")
		}
		if ss.Password != db.Session.Password {
			log.Panicln("Password not matched")
		}
		return db.Session
	})
}

func New() {

	client := &MockClient{
		Session: &Session{
			Email: "somecooluser@gmail.com",

			// This will make the program panics "Password not matched"
			Password: "bittersecret",
		},
	}

	db := &MockDB{
		Session: &Session{
			Email:    "somecooluser@gmail.com",
			Name:     "Cool User",
			Password: "sweetsecret",
			Data: map[string]interface{}{
				"status":    "sober",
				"score":     1500,
				"followers": 45,
			},
		},
	}

	getSession := SessionHandlerFunc(getUserSession)
	sessionSuite := db.Query(checkPassword(checkEmail(getSession)))
	serverSession := sessionSuite(client.Session)

	log.Println(serverSession)
}
