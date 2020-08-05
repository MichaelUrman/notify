package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/MichaelUrman/notify/internal/event"
)

const hookUrlInput = "hookurl"

type EventLoader func(context.Context) (*event.Detail, error)
type EventPreparer func(context.Context, *event.Detail) event.Submitter
type Environment interface {
	Dump(string, string)
	Debugf(string, ...interface{})
	Fatalf(string, ...interface{})
	Secret(string) string
}

func Main(env Environment, load EventLoader, prepare EventPreparer) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	defer func() {
		if err != nil {
			env.Dump("payload", os.Getenv("GITHUB_EVENT_PATH"))
			env.Fatalf("handling event: %v", err)
		} else {
			env.Debugf("No message sent")
		}
	}()

	detail, err := load(ctx)

	if err != nil || detail == nil {
		return err
	}
	if detail == nil {
		return nil
	}

	url := env.Secret(hookUrlInput)
	if url == "" {
		return fmt.Errorf("missing input %q", hookUrlInput)
	}

	req := prepare(ctx, detail)
	return req.Submit(ctx, url)
}

// PostJSON encodes req to JSON, and Posts it.
func PostJSON(ctx context.Context, url string, data interface{}) error {

	body := bytes.NewBuffer(nil)
	enc := json.NewEncoder(body)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(data); err != nil {
		return fmt.Errorf("encoding JSON: %w", err)
	}

	return Post(ctx, nil, url, body)
}

func Post(ctx context.Context, cli *http.Client, url string, body io.Reader) error {
	if cli == nil {
		cli = http.DefaultClient
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	req.Header.Add("Content-Type", "application/json")
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	resp, err := cli.Do(req)
	if err != nil {
		return fmt.Errorf("posting request: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		response, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("reading response: %w", err)
		}
		if err = resp.Body.Close(); err != nil {
			return fmt.Errorf("closing response: %w", err)
		}
		return fmt.Errorf("webhook failed (%v): %s", resp.StatusCode, response)
	}

	return nil
}
