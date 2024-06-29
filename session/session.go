package session

import (
	"context"
	"encoding/gob"
	"fmt"
	"net/http"

	"github.com/gofrs/uuid/v5"
	"github.com/gorilla/sessions"
)

var ctxKey contextKey = "session"

type contextKey string

func init() {
	// TODO: Look for a better place
	gob.Register(uuid.UUID{})
}

type Session interface {
	Register(w http.ResponseWriter, r *http.Request)
}

func New(secret, name string, options ...Option) Session {
	store := sessions.NewCookieStore([]byte(secret))

	// Run the options on the store
	for _, option := range options {
		option(store)
	}

	return &session{
		name:  name,
		store: store,
	}
}

type session struct {
	name  string
	store *sessions.CookieStore
}

func (s *session) Register(w http.ResponseWriter, r *http.Request) {
	session := s.session(r)

	r = r.WithContext(context.WithValue(r.Context(), ctxKey, session))

	w = &saver{
		w:     w,
		req:   r,
		store: session,
	}

	type valueSetter interface {
		Set(key string, value any)
	}

	// Look for a valuer in the context and set the values for flash
	// and session so that they can be used in other components of the request.
	if vlr, ok := r.Context().Value("valuer").(valueSetter); ok {
		vlr.Set("flash", flashHelper(session))
		vlr.Set("session", func() *sessions.Session { return session })

		return
	}

	fmt.Println("no valuer in context")
}

func (s *session) session(r *http.Request) *sessions.Session {
	session, _ := s.store.Get(r, s.name)
	return session
}
