package teams

import (
	"context"

	"github.com/MichaelUrman/notify/internal/event"
	"github.com/MichaelUrman/notify/internal/notifier"
)

type Request struct {
	Type            string    `json:"@type"`
	Context         string    `json:"@context"`
	Summary         string    `json:"summary,omitempty"`
	ThemeColor      string    `json:"themeColor,omitempty"`
	Title           string    `json:"title,omitempty"`
	Sections        []Section `json:"sections,omitempty"`
	PotentialAction []Action  `json:"potentialAction,omitempty"`
}

type Section struct {
	Image    string `json:"activityImage,omitempty"`
	Title    string `json:"activityTitle,omitempty"`
	Subtitle string `json:"activitySubtitle,omitempty"`
	Text     string `json:"activityText,omitempty"`
	Body     string `json:"text,omitempty"`
	Facts    []Fact `json:"facts,omitempty"`
}

type Fact struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Action struct {
	Type    string   `json:"@type"`
	Name    string   `json:"name"`
	Targets []Target `json:"targets,omitempty"`
}

type Target struct {
	OS  string `json:"os"`
	URI string `json:"uri"`
}

func (r Request) Submit(ctx context.Context, url string) error {
	return notifier.PostJSON(ctx, url, r)
}

func BuildSubmitter(ctx context.Context, d *event.Detail) event.Submitter {
	return Build(d)
}

func Build(d *event.Detail) *Request {
	if d == nil {
		return nil
	}
	req := Request{
		Type:       "MessageCard",
		Context:    "https://schema.org/extensions",
		ThemeColor: d.ThemeColor,
		Summary:    d.Summary,
		Title:      d.Title,
		Sections: []Section{
			{
				Image:    d.Avatar,
				Title:    d.Repository,
				Subtitle: d.Username,
				Text:     d.Text,
				Body:     d.Body,
			},
		},
	}

	sect0 := &req.Sections[0]
	for _, f := range d.Fact {
		sect0.Facts = append(sect0.Facts, Fact{f.Name, f.Value})
	}

	for _, a := range d.Action {
		req.PotentialAction = append(req.PotentialAction, Action{"OpenUri", a.Name, []Target{{"default", a.URL}}})
	}

	return &req
}
