// Package tenhou parses Tenhou's public result files (sc/raw), mjlog game logs, and fetches them
// politely. Formats: docs/research/tenhou-and-nodocchi.md. Usage rules: docs/design/worker/tenhou-ingest.md.
package tenhou

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"html"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// JST is Tenhou's time zone: file names, line times and log ids are all JST.
var JST = time.FixedZone("JST", 9*3600)

// RawFile is one entry of the sc/raw index (list.cgi).
type RawFile struct {
	Name string // e.g. scb2026112603.log.gz
	Size int64
}

var indexEntry = regexp.MustCompile(`\{file:'([^']+)',size:(\d+)\}`)

// ParseIndex parses list.cgi's JSONP body: list([{file:'…',size:…}, …]).
func ParseIndex(body []byte) ([]RawFile, error) {
	ms := indexEntry.FindAllSubmatch(body, -1)
	if len(ms) == 0 {
		return nil, fmt.Errorf("tenhou index: no entries")
	}
	out := make([]RawFile, 0, len(ms))
	for _, m := range ms {
		size, _ := strconv.ParseInt(string(m[2]), 10, 64)
		out = append(out, RawFile{Name: string(m[1]), Size: size})
	}
	return out, nil
}

var hourlyFile = regexp.MustCompile(`^(scb|scc)(\d{10})\.(log|html)\.gz$`)

// RankedHourly reports whether the file is an hourly scb/scc file (the only kinds we ingest, D-033)
// and returns its kind and JST start hour.
func RankedHourly(name string) (kind string, hour time.Time, ok bool) {
	m := hourlyFile.FindStringSubmatch(name)
	if m == nil {
		return "", time.Time{}, false
	}
	t, err := time.ParseInLocation("2006010215", m[2], JST)
	if err != nil {
		return "", time.Time{}, false
	}
	return m[1], t, true
}

// Rule is a decoded Tenhou rule (from a log id's type code or a result line's rule string).
type Rule struct {
	Code    string // 4-hex, e.g. "00a9"
	Players int
	Length  string // east | south
	Lobby   string // ippan | joukyuu | tokujou | houou
	Aka     bool
	Kuitan  bool
	Speed   string // normal | fast | hikari
}

var lobbyByChar = map[string]string{"般": "ippan", "上": "joukyuu", "特": "tokujou", "鳳": "houou"}
var lobbyBits = map[string]int{"ippan": 0x00, "joukyuu": 0x80, "tokujou": 0x20, "houou": 0xa0}

// ParseRuleString decodes a result-line rule such as 四鳳南喰赤－ or 三上東喰赤速. Only ranked
// lobbies (般上特鳳) without shuugi are accepted.
func ParseRuleString(s string) (Rule, error) {
	r := []rune(s)
	if len(r) < 6 {
		return Rule{}, fmt.Errorf("rule %q: too short", s)
	}
	var rule Rule
	switch r[0] {
	case '四':
		rule.Players = 4
	case '三':
		rule.Players = 3
	default:
		return Rule{}, fmt.Errorf("rule %q: players", s)
	}
	lobby, ok := lobbyByChar[string(r[1])]
	if !ok {
		return Rule{}, fmt.Errorf("rule %q: not a ranked lobby", s)
	}
	rule.Lobby = lobby
	switch r[2] {
	case '東':
		rule.Length = "east"
	case '南':
		rule.Length = "south"
	default:
		return Rule{}, fmt.Errorf("rule %q: length", s)
	}
	rule.Kuitan = r[3] == '喰'
	rule.Aka = r[4] == '赤'
	switch r[5] {
	case '速':
		rule.Speed = "fast"
	case '光':
		rule.Speed = "hikari"
	default:
		rule.Speed = "normal"
	}
	if len(r) > 6 {
		return Rule{}, fmt.Errorf("rule %q: shuugi or unknown suffix", s)
	}
	b := 0x01 | lobbyBits[rule.Lobby]
	if !rule.Aka {
		b |= 0x02
	}
	if !rule.Kuitan {
		b |= 0x04
	}
	if rule.Length == "south" {
		b |= 0x08
	}
	if rule.Players == 3 {
		b |= 0x10
	}
	switch rule.Speed {
	case "fast":
		b |= 0x40
	case "hikari":
		b |= 0x40 | 0x1000
	}
	rule.Code = fmt.Sprintf("%04x", b)
	return rule, nil
}

// DecodeRuleCode decodes a 4-hex type code (log id / mjlog GO type).
func DecodeRuleCode(code string) (Rule, error) {
	b64, err := strconv.ParseUint(code, 16, 16)
	if err != nil || len(code) != 4 {
		return Rule{}, fmt.Errorf("rule code %q", code)
	}
	b := int(b64)
	levels := []string{"ippan", "joukyuu", "tokujou", "houou"}
	r := Rule{
		Code:    strings.ToLower(code),
		Players: 4,
		Length:  "east",
		Lobby:   levels[((b&0x20)>>4)|((b&0x80)>>7)],
		Aka:     b&0x02 == 0,
		Kuitan:  b&0x04 == 0,
		Speed:   "normal",
	}
	if b&0x10 != 0 {
		r.Players = 3
	}
	if b&0x08 != 0 {
		r.Length = "south"
	}
	if b&0x40 != 0 {
		r.Speed = "fast"
		if b&0x1000 != 0 {
			r.Speed = "hikari"
		}
	}
	return r, nil
}

// Result is one finished game from an scb/scc line, players in placement order.
type Result struct {
	StartedAt   time.Time
	DurationMin int
	Rule        Rule
	LogID       string // scc only
	Players     []PlayerResult
}

type PlayerResult struct {
	Name     string
	PtTenths int // ±pt × 10, as published (e.g. +48.6 → 486)
}

var (
	logLink    = regexp.MustCompile(`log=(\d{10}gm-[0-9a-f]{4}-\d{4,5}-x?[0-9a-f]{8,16})`)
	playerTok  = regexp.MustCompile(`^(.*)\(([+-]?\d+(?:\.\d)?)\)$`)
	timeColumn = regexp.MustCompile(`^(\d{2}):(\d{2})$`)
)

// ParseLine parses one result line. hour is the file's JST hour (date and hour come from the file
// name; minutes from the line). htmlFile unescapes names (scc .html files). Lines for unranked or
// shuugi rules return an error.
func ParseLine(line string, hour time.Time, htmlFile bool) (Result, error) {
	line = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(line), "<br>"))
	cols := strings.Split(line, " | ")
	if len(cols) < 4 {
		return Result{}, fmt.Errorf("line %q: columns", line)
	}
	tm := timeColumn.FindStringSubmatch(strings.TrimSpace(cols[0]))
	if tm == nil {
		return Result{}, fmt.Errorf("line %q: time", line)
	}
	hh, _ := strconv.Atoi(tm[1])
	mm, _ := strconv.Atoi(tm[2])
	day := time.Date(hour.Year(), hour.Month(), hour.Day(), 0, 0, 0, 0, JST)
	res := Result{StartedAt: day.Add(time.Duration(hh)*time.Hour + time.Duration(mm)*time.Minute)}
	res.DurationMin, _ = strconv.Atoi(strings.TrimSpace(cols[1]))
	rule, err := ParseRuleString(strings.TrimSpace(cols[2]))
	if err != nil {
		return Result{}, err
	}
	res.Rule = rule
	playersCol := cols[len(cols)-1]
	if len(cols) >= 5 {
		if m := logLink.FindStringSubmatch(cols[3]); m != nil {
			res.LogID = m[1]
		}
	}
	for _, tok := range strings.Fields(playersCol) {
		m := playerTok.FindStringSubmatch(tok)
		if m == nil {
			return Result{}, fmt.Errorf("line %q: player %q", line, tok)
		}
		name := m[1]
		if htmlFile {
			name = html.UnescapeString(name)
		}
		res.Players = append(res.Players, PlayerResult{Name: name, PtTenths: tenths(m[2])})
	}
	if len(res.Players) != rule.Players {
		return Result{}, fmt.Errorf("line %q: %d players for %dp rule", line, len(res.Players), rule.Players)
	}
	return res, nil
}

func tenths(s string) int {
	f, _ := strconv.ParseFloat(s, 64)
	if f < 0 {
		return int(f*10 - 0.5)
	}
	return int(f*10 + 0.5)
}

// MatchKey identifies a game across sources without its log id: start hour (JST), rule code, and
// players with ±pt in placement order (docs/design/storage/game-record-model.md).
func MatchKey(startHour time.Time, ruleCode string, players []PlayerResult) string {
	var b strings.Builder
	b.WriteString(startHour.In(JST).Format("2006010215"))
	b.WriteString("|" + ruleCode)
	for _, p := range players {
		fmt.Fprintf(&b, "|%s|%d", p.Name, p.PtTenths)
	}
	sum := md5.Sum([]byte(b.String()))
	return hex.EncodeToString(sum[:])
}

// SortFiles orders raw files chronologically (by hour, scb before scc).
func SortFiles(files []RawFile) {
	sort.SliceStable(files, func(i, j int) bool {
		ki, hi, _ := RankedHourly(files[i].Name)
		kj, hj, _ := RankedHourly(files[j].Name)
		if !hi.Equal(hj) {
			return hi.Before(hj)
		}
		return ki < kj
	})
}
