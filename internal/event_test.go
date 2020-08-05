package main_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/MichaelUrman/notify/internal/event"
	"github.com/MichaelUrman/notify/internal/github"
	"github.com/MichaelUrman/notify/internal/teams"
	"github.com/google/go-cmp/cmp"
)

func TestGithub(t *testing.T) {
	testLoader(t, ".github.json", loadGithub)
}

func testLoader(t *testing.T, suffix string, loader func(*testing.T, string) *event.Detail) {
	if err := filepath.Walk("testdata", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			t.Log(err)
			return nil
		}

		t.Run(filepath.Join(filepath.Base(filepath.Dir(path)), strings.TrimSuffix(info.Name(), suffix)), func(t *testing.T) {
			if !info.IsDir() && strings.HasSuffix(path, suffix) {
				detail := loader(t, path)
				output := func(input, newSuffix string) string { return strings.ReplaceAll(input, suffix, newSuffix) }
				t.Run("event", func(t *testing.T) { testEvent(t, detail, path, output(path, ".event.json")) })
				t.Run("teams", func(t *testing.T) { testTeams(t, detail, path, output(path, ".teams.json")) })
			}
		})

		return nil
	}); err != nil {
		t.Error(err)
	}
}

func loadGithub(t *testing.T, input string) *event.Detail {
	detail, err := github.LoadTestEvent(context.Background(), github.TestEnv{
		EventName: filepath.Base(filepath.Dir(input)),
		EventPath: input,
	})
	if err != nil {
		t.Fatal("loading payload", err)
	}
	return detail
}

func decode(t *testing.T, output string, data interface{}) {
	want, err := os.Open(output)
	if err != nil {
		if os.IsNotExist(err) {
			t.Skipf("no test: %v", err)
		} else {
			t.Fatalf("Error reading test: %v", err)
		}
	}

	dec := json.NewDecoder(want)
	dec.DisallowUnknownFields()
	if err := dec.Decode(data); err != nil {
		t.Fatal("decoding test:", err)
	}
}

func testEvent(t *testing.T, detail *event.Detail, input, output string) {
	cases := struct {
		Want   *event.Detail `json:"WORKFLOW"`
		Pass   *event.Detail `json:"PASSED" status:"success"`
		Fail   *event.Detail `json:"FAILED" status:"failure"`
		Cancel *event.Detail `json:"CANCEL" status:"cancelled"`
		Skip   *event.Detail `json:"SKIPPED" status:"skipped"`
	}{}

	decode(t, output, &cases)
	compare(t, input, cases, func(detail *event.Detail) interface{} { return detail })
}

func testTeams(t *testing.T, detail *event.Detail, input, output string) {
	cases := struct {
		Want   *teams.Request `json:"WORKFLOW"`
		Pass   *teams.Request `json:"PASSED" status:"success"`
		Fail   *teams.Request `json:"FAILED" status:"failure"`
		Cancel *teams.Request `json:"CANCEL" status:"cancelled"`
		Skip   *teams.Request `json:"SKIPPED" status:"skipped"`
	}{}

	decode(t, output, &cases)
	compare(t, input, cases, func(detail *event.Detail) interface{} { return teams.Build(detail) })
}

func compare(t *testing.T, input string, cases interface{}, build func(*event.Detail) interface{}) {
	v := reflect.ValueOf(cases)
	for i := 0; i < v.NumField(); i++ {
		want := v.Field(i)
		if !want.IsZero() {
			status := v.Type().Field(i).Tag.Get("status")
			t.Run(status, func(t *testing.T) {
				detail, err := github.LoadTestEvent(context.Background(), github.TestEnv{
					WorkflowName: "WorkflowName",
					JobStatus:    status,
					RunID:        "12345",
					EventName:    filepath.Base(filepath.Dir(input)),
					EventPath:    input,
				})
				if err != nil {
					t.Fatal("loading status payload", err)
				}

				got := build(detail)
				if diff := cmp.Diff(want.Interface(), got); diff != "" {
					t.Errorf("Request (-got +want):\n%v", diff)
				}
			})
		}
	}
}
