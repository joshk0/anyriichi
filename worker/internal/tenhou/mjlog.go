package tenhou

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

// Log is the parsed subset of an mjlog (<mjloggm ver="2.3">) we store. Tag reference:
// Euophrys/tenhou-log README (see THANKS.md).
type Log struct {
	RuleCode string // GO type as 4-hex
	Lobby    int    // GO lobby (0 = ranked)
	Players  int
	Seats    []Seat // by seat (0 = initial dealer)
	Hands    []Hand
}

type Seat struct {
	Name       string
	Dan        int     // tenhou rank index at game start (0 新人 … 20 天鳳位)
	Rate       float64 // R at game start
	FinalScore int     // points
	PtTenths   int     // published ±pt × 10
	Placement  int     // 1-based; ties broken by seat order (closer to initial dealer ranks higher)
}

type Hand struct {
	Round        int // 0 = East 1 …
	Honba        int
	RiichiSticks int
	Dealer       int
	Outcome      string // ron | tsumo | draw | abortive
	Players      []HandPlayer
}

type HandPlayer struct {
	Won          bool
	WonBy        string // ron | tsumo
	DealtIn      bool
	Riichi       bool
	Called       bool
	TenpaiAtDraw *bool
	Han, Fu      int
	WinValue     int
	DealInValue  int
	Yaku         []int
	ScoreDelta   int
}

// abortive draw types (RYUUKYOKU type attribute); exhaustive draws have no type or nm (nagashi mangan).
var abortive = map[string]bool{"yao9": true, "reach4": true, "ron3": true, "kan4": true, "kaze4": true}

// ParseMjlog parses an mjlog XML document.
func ParseMjlog(r io.Reader) (*Log, error) {
	dec := xml.NewDecoder(r)
	l := &Log{}
	var cur *Hand
	haveUN, haveGO, done := false, false, false
	finish := func() {
		if cur != nil {
			l.Hands = append(l.Hands, *cur)
			cur = nil
		}
	}
	for {
		tok, err := dec.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("mjlog: %w", err)
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		a := attrs(se)
		switch se.Name.Local {
		case "GO":
			t, err := strconv.Atoi(a["type"])
			if err != nil {
				return nil, fmt.Errorf("mjlog: GO type %q", a["type"])
			}
			l.RuleCode = fmt.Sprintf("%04x", t)
			l.Lobby, _ = strconv.Atoi(a["lobby"])
			if t&0x10 != 0 {
				l.Players = 3
			} else {
				l.Players = 4
			}
			haveGO = true
		case "UN":
			if haveUN || a["dan"] == "" { // later UN tags are reconnects
				continue
			}
			if !haveGO {
				return nil, errors.New("mjlog: UN before GO")
			}
			dans := strings.Split(a["dan"], ",")
			rates := strings.Split(a["rate"], ",")
			for i := 0; i < l.Players; i++ {
				name, err := url.PathUnescape(a[fmt.Sprintf("n%d", i)])
				if err != nil {
					return nil, fmt.Errorf("mjlog: name n%d: %w", i, err)
				}
				s := Seat{Name: name}
				if i < len(dans) {
					s.Dan, _ = strconv.Atoi(dans[i])
				}
				if i < len(rates) {
					s.Rate, _ = strconv.ParseFloat(rates[i], 64)
				}
				l.Seats = append(l.Seats, s)
			}
			haveUN = true
		case "INIT":
			if !haveUN {
				return nil, errors.New("mjlog: INIT before UN")
			}
			finish()
			seed := ints(a["seed"])
			if len(seed) < 3 {
				return nil, fmt.Errorf("mjlog: INIT seed %q", a["seed"])
			}
			dealer, _ := strconv.Atoi(a["oya"])
			cur = &Hand{Round: seed[0], Honba: seed[1], RiichiSticks: seed[2], Dealer: dealer,
				Players: make([]HandPlayer, l.Players)}
		case "REACH":
			if cur != nil && a["step"] == "1" {
				if w, ok := seat(a["who"], l.Players); ok {
					cur.Players[w].Riichi = true
				}
			}
		case "N":
			if cur != nil {
				m, _ := strconv.Atoi(a["m"])
				if w, ok := seat(a["who"], l.Players); ok && openCall(m) {
					cur.Players[w].Called = true
				}
			}
		case "AGARI":
			if cur == nil {
				return nil, errors.New("mjlog: AGARI outside a hand")
			}
			who, ok1 := seat(a["who"], l.Players)
			from, ok2 := seat(a["fromWho"], l.Players)
			if !ok1 || !ok2 {
				return nil, fmt.Errorf("mjlog: AGARI who=%q fromWho=%q", a["who"], a["fromWho"])
			}
			ten := ints(a["ten"])
			p := &cur.Players[who]
			p.Won = true
			if who == from {
				p.WonBy, cur.Outcome = "tsumo", "tsumo"
			} else {
				p.WonBy, cur.Outcome = "ron", "ron"
				cur.Players[from].DealtIn = true
				if len(ten) > 1 {
					cur.Players[from].DealInValue += ten[1]
				}
			}
			if len(ten) > 1 {
				p.Fu, p.WinValue = ten[0], ten[1]
			}
			yaku := ints(a["yaku"]) // pairs: id, han
			for i := 0; i+1 < len(yaku); i += 2 {
				if yaku[i+1] > 0 { // dora entries are listed even with 0 han
					p.Yaku = append(p.Yaku, yaku[i])
					p.Han += yaku[i+1]
				}
			}
			for _, y := range ints(a["yakuman"]) {
				p.Yaku = append(p.Yaku, y)
				p.Han += 13
			}
			addScoreDeltas(cur, a["sc"])
			if a["owari"] != "" {
				if err := l.owari(a["owari"]); err != nil {
					return nil, err
				}
				done = true
			}
		case "RYUUKYOKU":
			if cur == nil {
				return nil, errors.New("mjlog: RYUUKYOKU outside a hand")
			}
			if abortive[a["type"]] {
				cur.Outcome = "abortive"
			} else {
				cur.Outcome = "draw"
				for i := range cur.Players {
					tenpai := a[fmt.Sprintf("hai%d", i)] != ""
					cur.Players[i].TenpaiAtDraw = &tenpai
				}
			}
			addScoreDeltas(cur, a["sc"])
			if a["owari"] != "" {
				if err := l.owari(a["owari"]); err != nil {
					return nil, err
				}
				done = true
			}
		}
	}
	finish()
	if !haveUN {
		return nil, errors.New("mjlog: no players")
	}
	if !done {
		return nil, errors.New("mjlog: game did not finish (no owari)")
	}
	return l, nil
}

// owari = "score0,pt0,score1,pt1,…" by seat, scores in hundreds.
func (l *Log) owari(s string) error {
	parts := strings.Split(s, ",")
	if len(parts) < 2*l.Players {
		return fmt.Errorf("mjlog: owari %q", s)
	}
	for i := 0; i < l.Players; i++ {
		sc, err1 := strconv.Atoi(parts[2*i])
		pt, err2 := strconv.ParseFloat(parts[2*i+1], 64)
		if err1 != nil || err2 != nil {
			return fmt.Errorf("mjlog: owari %q", s)
		}
		l.Seats[i].FinalScore = sc * 100
		l.Seats[i].PtTenths = tenths(strconv.FormatFloat(pt, 'f', 1, 64))
	}
	order := make([]int, l.Players)
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(a, b int) bool { return l.Seats[order[a]].FinalScore > l.Seats[order[b]].FinalScore })
	for place, s := range order {
		l.Seats[s].Placement = place + 1
	}
	return nil
}

// ByPlacement returns the players in placement order as PlayerResults (for MatchKey).
func (l *Log) ByPlacement() []PlayerResult {
	out := make([]PlayerResult, l.Players)
	for _, s := range l.Seats {
		out[s.Placement-1] = PlayerResult{Name: s.Name, PtTenths: s.PtTenths}
	}
	return out
}

// sc = "before0,delta0,before1,delta1,…" in hundreds.
func addScoreDeltas(h *Hand, sc string) {
	v := ints(sc)
	for i := range h.Players {
		if 2*i+1 < len(v) {
			h.Players[i].ScoreDelta += v[2*i+1] * 100
		}
	}
}

// openCall reports whether an N meld code is an open call (chi, pon, added kan, open kan) rather
// than a closed kan or a 3p kita.
func openCall(m int) bool {
	switch {
	case m&0x04 != 0: // chi
		return true
	case m&0x18 != 0: // pon (0x08) / added kan (0x10)
		return true
	case m&0x20 != 0: // kita (nukidora)
		return false
	default: // kan: from-who bits 0 = closed
		return m&0x03 != 0
	}
}

func seat(s string, players int) (int, bool) {
	n, err := strconv.Atoi(s)
	return n, err == nil && n >= 0 && n < players
}

func attrs(se xml.StartElement) map[string]string {
	m := make(map[string]string, len(se.Attr))
	for _, a := range se.Attr {
		m[a.Name.Local] = a.Value
	}
	return m
}

func ints(s string) []int {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil {
			return out
		}
		out = append(out, n)
	}
	return out
}

// ParseMjlogBytes is a convenience wrapper.
func ParseMjlogBytes(b []byte) (*Log, error) { return ParseMjlog(bytes.NewReader(b)) }
