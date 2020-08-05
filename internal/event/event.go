package event

import "context"

type Submitter interface {
	Submit(context.Context, string) error
}

type Detail struct {
	Summary    string
	ThemeColor string
	Repository string
	Username   string
	Avatar     string
	Title      string // avoid - it's huge

	Text string
	Body string

	Action []Action
	Fact   []Fact
}

type Fact struct{ Name, Value string }
type Action struct{ Name, URL string }
