package rules

import "testing"

func hou(placement int, anchor *int) ReplayGame {
	return ReplayGame{Players: 4, Level: "houou", Length: "south", Placement: placement, AnchorRank: anchor}
}

func val(p *int) any {
	if p == nil {
		return nil
	}
	return *p
}

func evt(p *RankEvent) any {
	if p == nil {
		return nil
	}
	return string(*p)
}

func TestReplayUnknownUntilAnchor(t *testing.T) {
	steps, err := ReplayTenhou([]ReplayGame{hou(1, nil), hou(4, nil)})
	if err != nil {
		t.Fatal(err)
	}
	for i, s := range steps {
		if s.RankBefore != nil || s.RankAfter != nil || s.Event != nil {
			t.Errorf("game %d: expected unknown, got %+v", i, s)
		}
	}
}

func TestReplayAnchorRange(t *testing.T) {
	// 7段 (16): promote 2800, south 鳳凰 1st +90. Range after: [90, 2799] stays, [2800, 2889] promotes.
	steps, err := ReplayTenhou([]ReplayGame{hou(1, ip(16))})
	if err != nil {
		t.Fatal(err)
	}
	s := steps[0]
	if val(s.RankBefore) != 16 || val(s.Delta) != 90 {
		t.Fatalf("got %+v", s)
	}
	if s.RankAfter != nil || s.Event != nil { // ambiguous: stayed or promoted
		t.Fatalf("expected ambiguity, got after=%v event=%v", val(s.RankAfter), evt(s.Event))
	}
}

func TestReplayPromotionResolvedByLaterAnchor(t *testing.T) {
	// Anchor 7段, then a 1st place; the next log shows 8段 → the 1st place was the promotion, and pt
	// is exact (8段 start 1600) from then on.
	steps, err := ReplayTenhou([]ReplayGame{hou(1, ip(16)), hou(2, ip(17)), hou(4, nil)})
	if err != nil {
		t.Fatal(err)
	}
	if evt(steps[0].Event) != "promotion" || val(steps[0].RankAfter) != 17 || val(steps[0].PtMin) != 1600 {
		t.Fatalf("game 0: %+v", steps[0])
	}
	// 2nd place south 鳳凰 = +45 → 1645.
	if val(steps[1].PtMin) != 1645 || val(steps[1].PtMax) != 1645 || evt(steps[1].Event) != "" {
		t.Fatalf("game 1: %+v", steps[1])
	}
	// 8段 last place south = −150 → 1495, exact.
	if val(steps[2].Delta) != -150 || val(steps[2].PtMin) != 1495 || val(steps[2].PtMax) != 1495 {
		t.Fatalf("game 2: %+v", steps[2])
	}
}

func TestReplayStayResolvedByLaterAnchor(t *testing.T) {
	steps, err := ReplayTenhou([]ReplayGame{hou(1, ip(16)), hou(3, ip(16))})
	if err != nil {
		t.Fatal(err)
	}
	if evt(steps[0].Event) != "" || val(steps[0].RankAfter) != 16 || val(steps[0].PtMin) != 90 || val(steps[0].PtMax) != 2799 {
		t.Fatalf("game 0: %+v", steps[0])
	}
	if val(steps[1].PtMin) != 90 || val(steps[1].PtMax) != 2799 {
		t.Fatalf("game 1: %+v", steps[1])
	}
}

func TestReplayCertainDemotion(t *testing.T) {
	// Promote 4段→… use 1段 (10): start 200 after promotion from 1級; anchor at 1級 with a known path
	// is hard to build, so: anchor 1段, range [0,399]; three south 鳳凰 lasts (−45 each) can't all
	// avoid demotion only if pt < 135. Instead pin pt via a promotion: 1級 (9) promote at 100.
	games := []ReplayGame{
		{Players: 4, Level: "joukyuu", Length: "south", Placement: 1, AnchorRank: ip(9)}, // +60
		{Players: 4, Level: "joukyuu", Length: "south", Placement: 1, AnchorRank: ip(10)}, // 1st was the promotion → 200, +60 → 260
		{Players: 4, Level: "joukyuu", Length: "south", Placement: 4},                     // −45 → 215
		{Players: 4, Level: "joukyuu", Length: "south", Placement: 4},                     // 170
		{Players: 4, Level: "joukyuu", Length: "south", Placement: 4},                     // 125
		{Players: 4, Level: "joukyuu", Length: "south", Placement: 4},                     // 80
		{Players: 4, Level: "joukyuu", Length: "south", Placement: 4},                     // 35
		{Players: 4, Level: "joukyuu", Length: "south", Placement: 4},                     // −10 → demoted to 1級 (0)
	}
	steps, err := ReplayTenhou(games)
	if err != nil {
		t.Fatal(err)
	}
	if evt(steps[0].Event) != "promotion" {
		t.Fatalf("game 0: %+v", steps[0])
	}
	if val(steps[1].PtMin) != 260 {
		t.Fatalf("game 1: %+v", steps[1])
	}
	last := steps[7]
	if evt(last.Event) != "demotion" || val(last.RankAfter) != 9 || val(last.PtMin) != 0 {
		t.Fatalf("game 7: before=%v after=%v pt=%v event=%v", val(last.RankBefore), val(last.RankAfter), val(last.PtMin), evt(last.Event))
	}
}

func TestReplayKyuFloor(t *testing.T) {
	// 3級 (7) can't lose points; 2級 (8) loses 15 on a south last and floors at 0.
	games := []ReplayGame{
		{Players: 4, Level: "ippan", Length: "south", Placement: 4, AnchorRank: ip(7)},
	}
	steps, _ := ReplayTenhou(games)
	if val(steps[0].Delta) != 0 || val(steps[0].PtMin) != 0 || val(steps[0].PtMax) != 99 {
		t.Fatalf("got %+v", steps[0])
	}
}

func TestReplayUnreconcilableAnchorRestarts(t *testing.T) {
	// Anchor 7段, then an anchor at 5段 (games outside our window): restart at 5段.
	steps, err := ReplayTenhou([]ReplayGame{hou(3, ip(16)), hou(3, ip(14))})
	if err != nil {
		t.Fatal(err)
	}
	if val(steps[0].RankAfter) != 16 || val(steps[1].RankBefore) != 14 || val(steps[1].PtMin) != 0 || val(steps[1].PtMax) != 1999 {
		t.Fatalf("got %+v / %+v", steps[0], steps[1])
	}
}

func TestReplayTenhoui(t *testing.T) {
	steps, _ := ReplayTenhou([]ReplayGame{hou(4, ip(20)), hou(1, nil)})
	for i, s := range steps {
		if val(s.RankAfter) != 20 || val(s.Delta) != 0 || evt(s.Event) != "" {
			t.Fatalf("game %d: %+v", i, s)
		}
	}
}
