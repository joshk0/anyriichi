// Package logref detects Tenhou and Mahjong Soul game-log references in user input.
// It mirrors packages/shared/src/logRef.ts; both are tested against packages/shared/testdata/log-refs.json.
package logref

import (
	"regexp"
	"strconv"
)

var (
	tenhouLogIDRe = regexp.MustCompile(`(?:^|[^0-9A-Za-z])(\d{10}gm-([0-9a-f]{4})-(\d{4,5})-x?[0-9a-f]{8,16})(?:[^0-9A-Za-z]|$)`)
	tenhouSeatRe  = regexp.MustCompile(`[?&]tw=([0-3])(?:[^0-9]|$)`)
	majsoulURLRe  = regexp.MustCompile(`https?://(mahjongsoul\.game\.yo-star\.com|game\.mahjongsoul\.com|game\.maj-soul\.com)/[^\s?#]*\?(?:[^\s#]*&)?paipu=(\d{6}-[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})(?:_a(\d+))?(?:[^0-9A-Za-z-]|$)`)
	majsoulAnyRe  = regexp.MustCompile(`https?://(?:mahjongsoul\.game\.yo-star\.com|game\.mahjongsoul\.com|game\.maj-soul\.com)/[^\s?#]*\?(?:[^\s#]*&)?paipu=`)
)

var majsoulHosts = map[string]string{
	"mahjongsoul.game.yo-star.com": "en",
	"game.mahjongsoul.com":         "jp",
	"game.maj-soul.com":            "cn",
}

// TenhouRule is the decoded 4-hex rule code.
type TenhouRule struct {
	Players int    `json:"players"`
	Length  string `json:"length"`
	Level   string `json:"level"`
	Speed   string `json:"speed"`
	Aka     bool   `json:"aka"`
	Kuitan  bool   `json:"kuitan"`
	Shuugi  bool   `json:"shuugi"`
}

// Ref is a detected log reference. Exactly one of the platform-specific field groups is set.
type Ref struct {
	Platform string `json:"platform,omitempty"`
	Error    string `json:"error,omitempty"`

	// Tenhou
	ID       string      `json:"id,omitempty"`
	Seat     *int        `json:"seat,omitempty"`
	Lobby    string      `json:"lobby,omitempty"`
	RuleCode string      `json:"ruleCode,omitempty"`
	Rule     *TenhouRule `json:"rule,omitempty"`
	Ranked   bool        `json:"ranked,omitempty"`

	// Mahjong Soul
	Server  string  `json:"server,omitempty"`
	UUID    string  `json:"uuid,omitempty"`
	Account *string `json:"account,omitempty"`
}

var levels = [...]string{"ippan", "joukyuu", "tokujou", "houou"}

// DecodeTenhouRule decodes a rule code such as "00a9".
func DecodeTenhouRule(code string) (TenhouRule, error) {
	b, err := strconv.ParseUint(code, 16, 16)
	if err != nil || len(code) != 4 {
		return TenhouRule{}, &strconv.NumError{Func: "DecodeTenhouRule", Num: code, Err: strconv.ErrSyntax}
	}
	r := TenhouRule{Players: 4, Length: "east", Speed: "normal", Aka: b&0x02 == 0, Kuitan: b&0x04 == 0, Shuugi: b&0x600 != 0}
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
	r.Level = levels[((b&0x20)>>4)|((b&0x80)>>7)]
	return r, nil
}

// Parse returns the log reference in input, or nil if none is recognised.
func Parse(input string) *Ref {
	if m := tenhouLogIDRe.FindStringSubmatch(input); m != nil {
		rule, err := DecodeTenhouRule(m[2])
		if err != nil {
			return nil
		}
		ref := &Ref{Platform: "tenhou", ID: m[1], RuleCode: m[2], Lobby: m[3], Rule: &rule, Ranked: m[3] == "0000"}
		if s := tenhouSeatRe.FindStringSubmatch(input); s != nil {
			n, _ := strconv.Atoi(s[1])
			ref.Seat = &n
		}
		return ref
	}
	if m := majsoulURLRe.FindStringSubmatch(input); m != nil {
		ref := &Ref{Platform: "majsoul", Server: majsoulHosts[m[1]], UUID: m[2]}
		if m[3] != "" {
			a := m[3]
			ref.Account = &a
		}
		return ref
	}
	if majsoulAnyRe.MatchString(input) {
		return &Ref{Error: "majsoul_anonymous"}
	}
	return nil
}
