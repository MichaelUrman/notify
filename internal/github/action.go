package github

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

// Thanks go to github.com/sethvargo/go-githubactions

var Actions = env{}

type env struct{}

func (e env) Secret(i string) string {
	s := e.Input(i)
	if s != "" {
		e.Mask(s)
	}
	return s
}

func (env) Input(i string) string {
	return strings.TrimSpace(os.Getenv("INPUT_" + strings.ReplaceAll(strings.ToUpper(i), " ", "_")))
}

func (env) Mask(value string)                        { println("::add-mask::" + value) }
func (env) Group(name string)                        { println("::group::" + name) }
func (env) EndGroup()                                { println("::endgroup::\n") }
func (env) Debugf(format string, a ...interface{})   { println("::debug::" + fmt.Sprintf(format, a...)) }
func (env) Warnf(format string, a ...interface{})    { println("::warn::" + fmt.Sprintf(format, a...)) }
func (env) Errorf(format string, a ...interface{})   { println("::error::" + fmt.Sprintf(format, a...)) }
func (e env) Fatalf(format string, a ...interface{}) { e.Errorf(format, a...); os.Exit(1) }

func (e env) Dump(title, fn string) {
	e.Group(title)
	payload, err := ioutil.ReadFile(os.Getenv("GITHUB_EVENT_PATH"))
	if err != nil {
		e.Debugf("reading payload: %v", err)
	} else {
		e.Debugf("%s", payload)
		e.EndGroup()
	}
}
