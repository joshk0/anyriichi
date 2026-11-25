package tenhou

import (
	"os"
	"testing"
	"time"
)

// All fixtures here are synthetic (made-up names and games): the repo is public and Tenhou
// forbids redistributing its result files and logs.

func TestParseIndex(t *testing.T) {
	files, err := ParseIndex([]byte("list([\n{file:'sca20261106.log.gz',size:55496},\n{file:'scb2026112603.log.gz',size:9150},\n{file:'2026/scc20260101.html.gz',size:1}\n]);"))
	if err != nil || len(files) != 3 || files[1].Name != "scb2026112603.log.gz" || files[1].Size != 9150 {
		t.Fatalf("got %+v %v", files, err)
	}
	if _, _, ok := RankedHourly("sca20261106.log.gz"); ok {
		t.Error("sca must not be ingested")
	}
	if _, _, ok := RankedHourly("2026/scc20260101.html.gz"); ok {
		t.Error("archived daily files are not hourly")
	}
	kind, hour, ok := RankedHourly("scc2026112603.html.gz")
	if !ok || kind != "scc" || !hour.Equal(time.Date(2026, 11, 26, 3, 0, 0, 0, JST)) {
		t.Errorf("RankedHourly = %q %v %v", kind, hour, ok)
	}
}

func TestParseRuleString(t *testing.T) {
	for s, code := range map[string]string{
		"四鳳南喰赤－": "00a9", "三鳳南喰赤－": "00b9", "三般南喰赤－": "0019",
		"四般東喰赤－": "0001", "四上東喰赤速": "00c1", "四特南喰赤－": "0029", "四般東－－－": "0007",
	} {
		r, err := ParseRuleString(s)
		if err != nil || r.Code != code {
			t.Errorf("ParseRuleString(%s) = %+v, %v; want code %s", s, r, err, code)
			continue
		}
		d, _ := DecodeRuleCode(code)
		if d != r {
			t.Errorf("DecodeRuleCode(%s) = %+v, want %+v", code, d, r)
		}
	}
	for _, s := range []string{"四若東喰赤－祝５", "四技東喰赤－", "x"} {
		if _, err := ParseRuleString(s); err == nil {
			t.Errorf("ParseRuleString(%s) should fail", s)
		}
	}
}

func TestParseLine(t *testing.T) {
	hour := time.Date(2026, 11, 26, 3, 0, 0, 0, JST)
	r, err := ParseLine("03:07 | 20 | 四特南喰赤－ | Alice(+61.6) ヽ(・∀・)ノ(+3.3) NoName(-23.4) Bob(-41.5)", hour, false)
	if err != nil {
		t.Fatal(err)
	}
	if !r.StartedAt.Equal(hour.Add(7*time.Minute)) || r.DurationMin != 20 || r.Rule.Code != "0029" || r.LogID != "" {
		t.Errorf("got %+v", r)
	}
	want := []PlayerResult{{"Alice", 616}, {"ヽ(・∀・)ノ", 33}, {"NoName", -234}, {"Bob", -415}}
	for i, p := range want {
		if r.Players[i] != p {
			t.Errorf("player %d = %+v, want %+v", i, r.Players[i], p)
		}
	}

	scc, err := ParseLine(`03:07 | 20 | 四特南喰赤－ | <a href="http://tenhou.net/0/?log=2026112603gm-0029-0000-0123abcd">牌譜</a> | Alice(+61.6) ヽ(・∀・)ノ(+3.3) NoName(-23.4) Bob(-41.5)<br>`, hour, true)
	if err != nil || scc.LogID != "2026112603gm-0029-0000-0123abcd" {
		t.Fatalf("scc: %+v %v", scc, err)
	}
	if MatchKey(r.StartedAt.Truncate(time.Hour), r.Rule.Code, r.Players) != MatchKey(hour, scc.Rule.Code, scc.Players) {
		t.Error("scb and scc rows of the same game must share a match key")
	}

	h, err := ParseLine("03:10 | 05 | 三上南喰赤－ | A&amp;B(+95.2) C(-5.0) D(-90.2)", hour, true)
	if err != nil || h.Players[0].Name != "A&B" || h.Rule.Players != 3 {
		t.Errorf("3p html: %+v %v", h, err)
	}
	if _, err := ParseLine("03:10 | 05 | 四般東喰赤－ | A(+1.0) B(-1.0)", hour, false); err == nil {
		t.Error("player count mismatch should fail")
	}
}

func TestParseMjlog(t *testing.T) {
	b, err := os.ReadFile("testdata/game4p.mjlog")
	if err != nil {
		t.Fatal(err)
	}
	l, err := ParseMjlogBytes(b)
	if err != nil {
		t.Fatal(err)
	}
	if l.RuleCode != "00a9" || l.Lobby != 0 || l.Players != 4 || len(l.Hands) != 3 {
		t.Fatalf("header: %+v", l)
	}
	wantSeats := []Seat{
		{Name: "Alice", Dan: 16, Rate: 2100.5, FinalScore: 33700, PtTenths: 467, Placement: 1},
		{Name: "ボブ", Dan: 10, Rate: 1650, FinalScore: 25500, PtTenths: 55, Placement: 2},
		{Name: "Carol", Dan: 20, Rate: 2300.25, FinalScore: 15300, PtTenths: -347, Placement: 4},
		{Name: "Dave", Dan: 8, Rate: 1500, FinalScore: 25500, PtTenths: -175, Placement: 3},
	}
	for i, w := range wantSeats {
		if l.Seats[i] != w {
			t.Errorf("seat %d = %+v, want %+v (ties broken by seat order)", i, l.Seats[i], w)
		}
	}
	h0 := l.Hands[0]
	if h0.Outcome != "ron" || !h0.Players[0].Won || h0.Players[0].WonBy != "ron" || !h0.Players[0].Riichi ||
		h0.Players[0].Han != 3 || h0.Players[0].Fu != 30 || h0.Players[0].WinValue != 7700 ||
		!h0.Players[2].DealtIn || h0.Players[2].DealInValue != 7700 || h0.Players[2].ScoreDelta != -7700 {
		t.Errorf("hand 0: %+v", h0)
	}
	if !h0.Players[1].Called || h0.Players[3].Called != true { // 51786 = pon, 4127 = chi
		t.Errorf("calls: %+v", h0.Players)
	}
	if len(h0.Players[0].Yaku) != 3 { // riichi, tanyao, dora (aka with 0 han dropped)
		t.Errorf("yaku: %v", h0.Players[0].Yaku)
	}
	h1 := l.Hands[1]
	if h1.Outcome != "draw" || h1.Players[2].Called || *h1.Players[0].TenpaiAtDraw != true || *h1.Players[2].TenpaiAtDraw != false {
		t.Errorf("hand 1: %+v", h1) // 2048 = closed kan → not a call
	}
	h2 := l.Hands[2]
	if h2.Outcome != "tsumo" || h2.Honba != 1 || h2.Players[3].WonBy != "tsumo" || h2.Players[1].ScoreDelta != -1000 {
		t.Errorf("hand 2: %+v", h2)
	}
	if got := l.ByPlacement(); got[3].Name != "Carol" {
		t.Errorf("ByPlacement = %+v", got)
	}
}

func TestParseMjlogUnfinished(t *testing.T) {
	_, err := ParseMjlogBytes([]byte(`<mjloggm ver="2.3"><GO type="169" lobby="0"/><UN n0="a" n1="b" n2="c" n3="d" dan="1,1,1,1" rate="1,1,1,1"/><INIT seed="0,0,0,0,0,0" oya="0"/></mjloggm>`))
	if err == nil {
		t.Error("want error for a game without owari")
	}
}
