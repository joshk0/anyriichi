package logref

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// vectorExpect mirrors the JSON shape in packages/shared/testdata/log-refs.json.
type vectorExpect struct {
	Platform string      `json:"platform"`
	Error    string      `json:"error"`
	ID       string      `json:"id"`
	Seat     *int        `json:"seat"`
	Lobby    string      `json:"lobby"`
	RuleCode string      `json:"ruleCode"`
	Rule     *TenhouRule `json:"rule"`
	Ranked   bool        `json:"ranked"`
	Server   string      `json:"server"`
	UUID     string      `json:"uuid"`
	Account  *string     `json:"account"`
}

func TestSharedVectors(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "packages", "shared", "testdata", "log-refs.json"))
	if err != nil {
		t.Fatal(err)
	}
	var v struct {
		Cases []struct {
			Input  string        `json:"input"`
			Expect *vectorExpect `json:"expect"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(raw, &v); err != nil {
		t.Fatal(err)
	}
	if len(v.Cases) == 0 {
		t.Fatal("no vectors")
	}
	for _, c := range v.Cases {
		t.Run(c.Input, func(t *testing.T) {
			got := Parse(c.Input)
			if c.Expect == nil {
				if got != nil {
					t.Fatalf("expected no match, got %+v", got)
				}
				return
			}
			if got == nil {
				t.Fatal("expected a match, got nil")
			}
			want := Ref{Platform: c.Expect.Platform, Error: c.Expect.Error, ID: c.Expect.ID, Seat: c.Expect.Seat,
				Lobby: c.Expect.Lobby, RuleCode: c.Expect.RuleCode, Rule: c.Expect.Rule, Ranked: c.Expect.Ranked,
				Server: c.Expect.Server, UUID: c.Expect.UUID, Account: c.Expect.Account}
			gj, _ := json.Marshal(got)
			wj, _ := json.Marshal(want)
			if string(gj) != string(wj) {
				t.Fatalf("\n got  %s\n want %s", gj, wj)
			}
		})
	}
}
