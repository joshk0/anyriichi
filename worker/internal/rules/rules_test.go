package rules

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

const shared = "../../../packages/shared"

func TestDataMatchesShared(t *testing.T) {
	for _, f := range []string{"tenhou.json", "majsoul.json"} {
		want, err := os.ReadFile(filepath.Join(shared, "rules", f))
		if err != nil {
			t.Fatal(err)
		}
		got, err := os.ReadFile(filepath.Join("data", f))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, want) {
			t.Errorf("data/%s differs from packages/shared/rules/%s; run `make rules-sync`", f, f)
		}
	}
}

type vectors struct {
	ResultPt []struct {
		Platform  string  `json:"platform"`
		Players   int     `json:"players"`
		Score     int     `json:"score"`
		Placement int     `json:"placement"`
		Expect    float64 `json:"expect"`
	} `json:"resultPt"`
	TenhouGames []struct {
		Name      string `json:"name"`
		Rank      int    `json:"rank"`
		Pt        int    `json:"pt"`
		Players   int    `json:"players"`
		Level     string `json:"level"`
		Length    string `json:"length"`
		Placement int    `json:"placement"`
		Expect    struct {
			Delta int     `json:"delta"`
			Rank  int     `json:"rank"`
			Pt    int     `json:"pt"`
			Event *string `json:"event"`
		} `json:"expect"`
	} `json:"tenhouGames"`
	MajsoulGames []struct {
		Name      string `json:"name"`
		LevelID   int    `json:"levelId"`
		Points    int    `json:"points"`
		Mode      int    `json:"mode"`
		Score     int    `json:"score"`
		Placement int    `json:"placement"`
		Expect    struct {
			Delta   int     `json:"delta"`
			LevelID int     `json:"levelId"`
			Points  int     `json:"points"`
			Event   *string `json:"event"`
		} `json:"expect"`
	} `json:"majsoulGames"`
	TenhouSupported []struct {
		LogID  string `json:"logId"`
		Expect bool   `json:"expect"`
	} `json:"tenhouSupported"`
}

func load(t *testing.T) vectors {
	b, err := os.ReadFile(filepath.Join(shared, "testdata", "rules.json"))
	if err != nil {
		t.Fatal(err)
	}
	var v vectors
	if err := json.Unmarshal(b, &v); err != nil {
		t.Fatal(err)
	}
	return v
}

func ev(s *string) RankEvent {
	if s == nil {
		return NoEvent
	}
	return RankEvent(*s)
}

func TestVectors(t *testing.T) {
	v := load(t)
	for _, c := range v.ResultPt {
		got, err := ResultPt(c.Platform, c.Players, c.Score, c.Placement)
		if err != nil || got != c.Expect {
			t.Errorf("ResultPt(%s,%d,%d,%d) = %v, %v; want %v", c.Platform, c.Players, c.Score, c.Placement, got, err, c.Expect)
		}
	}
	for _, c := range v.TenhouGames {
		got, err := ApplyTenhouGame(TenhouRank{Rank: c.Rank, Pt: c.Pt}, TenhouGame{Players: c.Players, Level: c.Level, Length: c.Length, Placement: c.Placement})
		if err != nil || got.Delta != c.Expect.Delta || got.After.Rank != c.Expect.Rank || got.After.Pt != c.Expect.Pt || got.Event != ev(c.Expect.Event) {
			t.Errorf("%s: got %+v, %v; want %+v", c.Name, got, err, c.Expect)
		}
	}
	for _, c := range v.MajsoulGames {
		got, err := ApplyMajsoulGame(MajsoulRank{LevelID: c.LevelID, Points: c.Points}, MajsoulGame{Mode: c.Mode, Score: c.Score, Placement: c.Placement})
		if err != nil || got.Delta != c.Expect.Delta || got.After.LevelID != c.Expect.LevelID || got.After.Points != c.Expect.Points || got.Event != ev(c.Expect.Event) {
			t.Errorf("%s: got %+v, %v; want %+v", c.Name, got, err, c.Expect)
		}
	}
	for _, c := range v.TenhouSupported {
		if got := TenhouLogSupported(c.LogID); got != c.Expect {
			t.Errorf("TenhouLogSupported(%s) = %v, want %v", c.LogID, got, c.Expect)
		}
	}
}

func TestNamesAndErrors(t *testing.T) {
	if got := TenhouRankName(16, "ja"); got != "七段" {
		t.Errorf("got %q", got)
	}
	if got := MajsoulLevelName(10402, "en"); got != "Master 2" {
		t.Errorf("got %q", got)
	}
	if got := MajsoulLevelName(20703, "ja"); got != "魂天 3" {
		t.Errorf("got %q", got)
	}
	if _, err := ResultPt("tenhou", 3, 30000, 4); err == nil {
		t.Error("want error for placement 4 in 3p")
	}
	if _, err := MajsoulRankDelta(10401, MajsoulGame{Mode: 13, Score: 30000, Placement: 1}); err == nil {
		t.Error("want error for unsupported mode")
	}
}
