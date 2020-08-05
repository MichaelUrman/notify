package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/text/message"

	"github.com/MichaelUrman/notify/internal/event"
)

func LoadEvent(ctx context.Context) (*event.Detail, error) {
	lang := message.MatchLanguage(Actions.Input("lang"), "en")
	pr := message.NewPrinter(lang)

	status := Actions.Input("job-status")
	if status != "" {
		return ReportWorkflowStatus(ctx, pr, status)
	}
	return ParseWorkflow(ctx, pr)
}

type TestEnv struct {
	RunID        string
	EventName    string
	WorkflowName string
	EventPath    string
	JobStatus    string
	Lang         string
}

func LoadTestEvent(ctx context.Context, env TestEnv) (*event.Detail, error) {
	os.Setenv("GITHUB_ACTIONS", "true")
	os.Setenv("GITHUB_RUN_ID", env.RunID)
	os.Setenv("GITHUB_WORKFLOW", env.WorkflowName)
	os.Setenv("GITHUB_EVENT_NAME", env.EventName)
	os.Setenv("GITHUB_EVENT_PATH", env.EventPath)
	os.Setenv("INPUT_JOB-STATUS", env.JobStatus)
	os.Setenv("INPUT_LANG", env.Lang)

	return LoadEvent(ctx)
}

type eventer interface {
	Event(*message.Printer) *event.Detail
}

// Note that the following structures are intentionally incomplete.
// We don't care about most of what GitHub webhooks provide.
//
// Reference: https://docs.github.com/en/developers/webhooks-and-events/webhook-events-and-payloads

type Common struct {
	Action string
	Sender struct {
		Login     string
		AvatarURL string `json:"avatar_url"`
	}
	Repository struct {
		FullName      string `json:"full_name"`
		DefaultBranch string `json:"default_branch"`
		URL           string `json:"html_url"`
	}
	Organization struct {
		Login string
	}
	Ref        string
	HeadCommit struct {
		ID  string
		URL string
	} `json:"head_commit"`
}

type CheckRun struct {
	Common
	CheckRun struct {
		Name   string
		Output struct {
			Title   string
			Summary string
		}
		CheckSuite struct {
			App struct {
				Name      string
				AvatarURL string `json:"avatar_url"`
			}
			HeadBranch string `json:"head_branch"`
		} `json:"check_suite"`
	} `json:"check_run"`
}

func (ev CheckRun) Event(p *message.Printer) *event.Detail {
	switch ev.Action {
	case "completed":
		username := ev.CheckRun.CheckSuite.App.Name
		if strings.HasPrefix(ev.CheckRun.Name, username) {
			username = ev.CheckRun.Name
		}
		head := md(branch(ev.CheckRun.CheckSuite.HeadBranch))
		return fillEvent(p, ev.Common, event.Detail{
			Summary:  strings.TrimPrefix(fmt.Sprintf("%s: %s", head, md(ev.CheckRun.Output.Title)), ": "),
			Username: username,
			Avatar:   ev.CheckRun.CheckSuite.App.AvatarURL,
			Text:     strings.TrimSpace(fmt.Sprintf("%#+s %#s", head, md(ev.CheckRun.Output.Title))),
			Body:     ev.CheckRun.Output.Summary,
		})
	}
	return nil
}

type Create struct {
	Common
	RefType string `json:"ref_type"`
}

func (ev Create) Event(p *message.Printer) *event.Detail {
	switch ev.RefType {
	case "tag":
		tagger := md(ev.Sender.Login)
		tagName := md(tag(ev.Ref))
		return fillEvent(p, ev.Common, event.Detail{
			Summary: p.Sprintf(message.Key(createTag, "%s tagged %s"), tagger, tagName),
			Text:    p.Sprintf(msgUserCreatedTag, tagger, tagName),
			Action:  []event.Action{{URL: ev.Repository.URL + "/tree/" + ev.Ref}},
		})
	default:
		return nil // ignore branches; only interesting at push.
	}
}

type Delete struct {
	Common
	RefType string `json:"ref_type"`
}

func (ev Delete) Event(p *message.Printer) *event.Detail {
	username := md(ev.Sender.Login)
	switch ev.RefType {
	case "branch":
		branchName := md(branch(ev.Ref))
		return fillEvent(p, ev.Common, event.Detail{
			Summary: p.Sprintf(message.Key(deleteBranch, "%s deleted %s"), username, branchName),
			Text:    p.Sprintf(msgUserDeletedBranch, username, branchName),
		})
	case "tag":
		tagName := md(tag(ev.Ref))
		return fillEvent(p, ev.Common, event.Detail{
			Summary: p.Sprintf(message.Key(deleteTag, "%s deleted %s"), username, tagName),
			Text:    p.Sprintf(msgUserDeletedTag, username, tagName),
		})
	}
	return nil
}

type PullRequest struct {
	Common
	PullRequest struct {
		Number int
		Title  string
		Draft  bool
		Head   struct {
			Ref string
		}
		Base struct {
			Ref string
		}
		URL    string `json:"html_url"`
		Body   string
		Merged bool
	} `json:"pull_request"`
}

func (ev PullRequest) Event(p *message.Printer) *event.Detail {
	head := branch(ev.PullRequest.Head.Ref)
	base := branch(ev.PullRequest.Base.Ref)

	switch ev.Action {
	case "opened", "closed":
		username := md(ev.Sender.Login)
		verb := ev.Action + "||pr"
		if ev.PullRequest.Draft {
			verb += "|draft"
		}
		if ev.PullRequest.Merged {
			verb += "|merged"
		}
		var title string
		if head != "" && base != "" {
			title = p.Sprintf(msgUserVerbedPRBranch, username, verb, ev.PullRequest.Number, md(head), md(base))
		} else {
			title = p.Sprintf(msgUserVerbedPRTitle, username, verb, ev.PullRequest.Number, md(ev.PullRequest.Title))
		}
		summaryVerb := verb + "|summary"
		return fillEvent(p, ev.Common, event.Detail{
			Summary: p.Sprintf(message.Key(msgVerbedPR, "%s %m #%#d"), username, summaryVerb, ev.PullRequest.Number),
			Text:    title,
			Action:  []event.Action{{Name: p.Sprintf(viewPR, ev.PullRequest.Number), URL: ev.PullRequest.URL}},
			Body:    ev.PullRequest.Body,
		})
	case "reviewed":
	}
	return nil
}

type PullRequestReview struct {
	Common
	PullRequest struct {
		Number int
		Title  string
		URL    string `json:"html_url"`
	} `json:"pull_request"`
	Review struct {
		State string
		URL   string `json:"html_url"`
	}
}

func (ev *PullRequestReview) Event(p *message.Printer) *event.Detail {
	username := md(ev.Sender.Login)
	verb := ev.Review.State + "||review"
	switch ev.Action {
	case "submitted":
		return fillEvent(p, ev.Common, event.Detail{
			Summary: p.Sprintf(message.Key(msgReviewedPR, "%s reviewed #%#d"), username, ev.PullRequest.Number),
			Text:    p.Sprintf(msgUserSubmittedReview, username, ev.PullRequest.Number),
			Body:    p.Sprintf(msgUserReviewState, username, verb, ev.PullRequest.Number, md(ev.PullRequest.Title), ev.PullRequest.URL),
			Action:  []event.Action{{Name: p.Sprint(viewReview), URL: ev.Review.URL}},
		})
	case "edited":
		return fillEvent(p, ev.Common, event.Detail{
			Summary: p.Sprintf(message.Key(msgEditedReview, "%s edited #%#d review"), username, ev.PullRequest.Number),
			Text:    p.Sprintf(msgUserEditedReview, username, ev.PullRequest.Number),
			Body:    p.Sprintf(msgUserReviewState, username, verb, ev.PullRequest.Number, md(ev.PullRequest.Title), ev.PullRequest.URL),
			Action:  []event.Action{{Name: p.Sprint(viewReview), URL: ev.Review.URL}},
		})
	}
	return nil
}

type PullRequestReviewComment struct{}

func (ev PullRequestReviewComment) Event(p *message.Printer) *event.Detail {
	return nil
}

type Push struct {
	Common
	Before, After string
	CompareURL    string `json:"compare_url"`
	Commits       []struct {
		ID      string
		URL     string
		Message string
		Author  struct {
			Name string
		}
		Distinct bool
	}
	Pusher struct {
		Name string
	}
	Forced bool
}

func (ev Push) Event(p *message.Printer) *event.Detail {
	if len(ev.Commits) == 0 {
		return nil
	}

	head := branch(ev.Ref)
	branchName := md(head)
	pusherName := md(ev.Sender.Login)
	pushType := branchPushed
	if ev.Forced {
		pushType = branchForced
	}
	summary := p.Sprintf(message.Key(branchPushSummary, "%s %m %s"), pusherName, pushType, branchName)
	text := p.Sprintf(message.Key(branchPushText, "%#+s %m %d commits to %#+m"), pusherName, pushType, len(ev.Commits), branchName)

	var commits []event.Fact
	var view []event.Action
	for _, commit := range ev.Commits {
		msg := msgNewCommitMessageLink
		if !commit.Distinct {
			msg = msgRepeatCommitMessageLink
		}

		// message is maximum of 60 characters and first line
		message := commit.Message
		if len(message) > 60 {
			message = message[:60]
		}
		shortMessage := md(strings.Split(message, "\n")[0])
		commits = append(commits, event.Fact{Name: commit.ID[:9], Value: p.Sprintf(msg, shortMessage, commit.URL)})
	}
	if len(commits) > 1 {
		view = append(view, event.Action{Name: p.Sprint(viewPush), URL: strings.ReplaceAll(ev.CompareURL, "^", "%5E")})
	}
	if head != ev.Repository.DefaultBranch && ev.Repository.DefaultBranch != "" {
		base := ev.Repository.DefaultBranch
		view = append(view, event.Action{Name: p.Sprintf(msgCompareBaseToBranch, base, head), URL: fmt.Sprintf("%s/compare/%s", ev.Repository.URL, head)})
	}

	return fillEvent(p, ev.Common, event.Detail{Summary: summary, Text: text, Fact: commits, Action: view})
}

type JobStatus struct {
	Common
	JobName   string
	JobStatus string
	JobURL    string
}

func (ev JobStatus) Event(p *message.Printer) *event.Detail {
	jobName := md(ev.JobName)
	jobStatus := ev.JobStatus + "||job"
	symbol := ev.JobStatus + "||job|sym"

	refName := md(strings.TrimPrefix(strings.TrimPrefix(ev.Ref, "refs/tags/"), "refs/heads/"))
	commitLinkMarkdown := ""
	if ev.HeadCommit.ID != "" && ev.HeadCommit.URL != "" {
		commitLinkMarkdown = fmt.Sprintf(`[%s](%s)`, ev.HeadCommit.ID[:9], ev.HeadCommit.URL)
	}

	refOrSha := refName
	if refName == "" {
		refOrSha = md(ev.HeadCommit.ID[:9])
	}

	return fillEvent(p, ev.Common, event.Detail{
		Username: string(jobName),
		Summary:  p.Sprintf(message.Key(msgWorkflowStatusSummary, "%s %m for %s"), jobName, jobStatus, refOrSha),
		Text:     p.Sprintf(message.Key(msgWorkflowStatus, "%m %#s %m for %#+s"), symbol, jobName, jobStatus, refOrSha),
		Body:     p.Sprintf(message.Key(msgWorkflowDetail, "%m Workflow %#+s %m for %+s commit %s"), symbol, jobName, jobStatus, refName, commitLinkMarkdown),
	})
}

type md string

var _ fmt.Formatter = md("")

func (s md) Format(f fmt.State, c rune) {
	switch c {
	case 's', 'v': // escape
		if s != "" && f.Flag('+') {
			f.Write([]byte{'*', '*'})
		}
		if f.Flag('#') {
			for _, c := range []byte(s) {
				if strings.ContainsRune(`()[]{}_\!.#*+-`+"`", rune(c)) {
					f.Write([]byte{'\\', c})
				} else {
					f.Write([]byte{c})
				}
			}
		} else {
			f.Write([]byte(s))
		}
		if s != "" && f.Flag('+') {
			f.Write([]byte{'*', '*'})
		}
	}
}

func branch(ref string) string {
	return strings.TrimPrefix(ref, "refs/heads/")
}
func tag(ref string) string {
	return strings.TrimPrefix(ref, "refs/tags/")
}

func ReportWorkflowStatus(ctx context.Context, p *message.Printer, status string) (*event.Detail, error) {
	// https://docs.github.com/en/actions/configuring-and-managing-workflows/using-environment-variables
	if os.Getenv("GITHUB_ACTIONS") != "true" {
		return nil, fmt.Errorf("not in GitHub Actions")
	}

	return ReportFile(ctx, p, os.Getenv("GITHUB_WORKFLOW"), status, os.Getenv("GITHUB_RUN_ID"), os.Getenv("GITHUB_EVENT_PATH"))
}

func ReportFile(ctx context.Context, p *message.Printer, workflow, status, runID, payloadPath string) (*event.Detail, error) {
	r, err := os.Open(payloadPath)
	if err != nil {
		return nil, fmt.Errorf("missing payload: %w", err)
	}
	defer r.Close()

	return Report(ctx, p, workflow, status, runID, r)
}

func Report(ctx context.Context, p *message.Printer, workflow, status, runID string, payload io.Reader) (*event.Detail, error) {
	sum, err := parse(ctx, "_job_status", payload)
	if err != nil {
		return nil, err
	}

	job := sum.(*JobStatus)
	job.JobName = workflow
	job.JobStatus = status
	job.JobURL = job.Repository.URL + "/actions/runs/" + runID

	return sum.Event(p), err
}

func ParseWorkflow(ctx context.Context, p *message.Printer) (*event.Detail, error) {
	// https://docs.github.com/en/actions/configuring-and-managing-workflows/using-environment-variables
	if os.Getenv("GITHUB_ACTIONS") != "true" {
		return nil, fmt.Errorf("not in GitHub Actions")
	}
	return ParseFile(ctx, p, os.Getenv("GITHUB_EVENT_NAME"), os.Getenv("GITHUB_EVENT_PATH"))
}

func ParseFile(ctx context.Context, p *message.Printer, event string, payloadPath string) (*event.Detail, error) {
	r, err := os.Open(payloadPath)
	if err != nil {
		return nil, fmt.Errorf("missing payload: %w", err)
	}
	defer r.Close()
	e, err := parse(ctx, event, r)
	if err != nil {
		return nil, err
	}
	return e.Event(p), nil
}

func parse(ctx context.Context, event string, payload io.Reader) (eventer, error) {
	json := json.NewDecoder(payload)

	var sum eventer

	switch event {
	case "check_run":
		sum = &CheckRun{}
	case "create":
		sum = &Create{}
	case "delete":
		sum = &Delete{}
	case "pull_request":
		sum = &PullRequest{}
	case "pull_request_review":
		sum = &PullRequestReview{}
	case "pull_request_review_comment":
		sum = &PullRequestReviewComment{}
	case "push":
		sum = &Push{}
	case "_job_status":
		sum = &JobStatus{}
	default:
		return nil, fmt.Errorf("unsupported event: %v", event)
	}
	if err := json.Decode(sum); err != nil {
		return nil, fmt.Errorf("decoding webhook: %w", err)
	}
	return sum, nil
}

func fillEvent(p *message.Printer, c Common, ev event.Detail) *event.Detail {
	first := func(a, b string) string {
		if a == "" {
			a = b
		}
		return a
	}
	ev.ThemeColor = first(ev.ThemeColor, themeColor)
	ev.Repository = first(ev.Repository, c.Repository.FullName)
	ev.Username = first(ev.Username, c.Sender.Login)
	ev.Avatar = first(ev.Avatar, c.Sender.AvatarURL)

	if strings.Contains(ev.Text, ev.Username) {
		ev.Username = ""
	}

	for i, a := range ev.Action {
		if a.Name == "" {
			ev.Action[i].Name = p.Sprint(viewOnGithub)
		}
	}

	return &ev
}
