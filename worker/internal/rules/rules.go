// Package rules implements ranked-game rules: normalized ±pt (D-016) and rank-point progression for
// Tenhou and Mahjong Soul. It is a port of packages/shared/src/rules.ts; data files in data/ are
// copies of packages/shared/rules/*.json (TestDataMatchesShared enforces this) and both
// implementations run packages/shared/testdata/rules.json.
package rules

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"time"
)

//go:embed data/tenhou.json
var tenhouJSON []byte

//go:embed data/majsoul.json
var majsoulJSON []byte

type RankEvent string

const (
	NoEvent   RankEvent = ""
	Promotion RankEvent = "promotion"
	Demotion  RankEvent = "demotion"
)

type scoring struct {
	StartPoints  int   `json:"startPoints"`
	ReturnPoints int   `json:"returnPoints"`
	Uma          []int `json:"uma"`
}

type byLength struct {
	East  int `json:"east"`
	South int `json:"south"`
}

type tenhouRank struct {
	ID        int      `json:"id"`
	JA        string   `json:"ja"`
	EN        string   `json:"en"`
	StartPt   int      `json:"startPt"`
	PromotePt *int     `json:"promotePt"`
	LastPlace byLength `json:"lastPlace"`
	CanDemote bool     `json:"canDemote"`
}

var tenhou struct {
	SupportedFrom   time.Time                              `json:"supportedFrom"`
	Scoring         map[string]scoring                     `json:"scoring"`
	PlacementPoints map[string]map[string]map[string][]int `json:"-"`
	RawPlacement    map[string]json.RawMessage             `json:"placementPoints"`
	Ranks           []tenhouRank                           `json:"ranks"`
}

type majsoulMode struct {
	Room           string `json:"room"`
	Players        int    `json:"players"`
	Length         string `json:"length"`
	PlacementBonus []int  `json:"placementBonus"`
	PenaltyTable   string `json:"penaltyTable"`
	KontenDelta    []int  `json:"kontenDelta"`
}

var majsoul struct {
	Scoring map[string]scoring     `json:"scoring"`
	Modes   map[string]majsoulMode `json:"modes"`
	Majors  []struct {
		Major int    `json:"major"`
		JA    string `json:"ja"`
		EN    string `json:"en"`
	} `json:"majors"`
	LevelMaxPoints   []int            `json:"levelMaxPoints"`
	LastPlacePenalty map[string][]int `json:"lastPlacePenalty"`
	Konten           struct {
		Major     int `json:"major"`
		MaxPoints int `json:"maxPoints"`
		TopMinor  int `json:"topMinor"`
	} `json:"konten"`
}

func init() {
	if err := json.Unmarshal(tenhouJSON, &tenhou); err != nil {
		panic(fmt.Sprintf("rules: tenhou.json: %v", err))
	}
	tenhou.PlacementPoints = map[string]map[string]map[string][]int{}
	for players, raw := range tenhou.RawPlacement {
		if players == "$comment" {
			continue
		}
		var m map[string]map[string][]int
		if err := json.Unmarshal(raw, &m); err != nil {
			panic(fmt.Sprintf("rules: tenhou placementPoints[%s]: %v", players, err))
		}
		tenhou.PlacementPoints[players] = m
	}
	if err := json.Unmarshal(majsoulJSON, &majsoul); err != nil {
		panic(fmt.Sprintf("rules: majsoul.json: %v", err))
	}
}

func checkPlacement(placement, players int) error {
	if placement < 1 || placement > players {
		return fmt.Errorf("invalid placement %d for %d players", placement, players)
	}
	return nil
}

func scoringFor(platform string, players int) (scoring, error) {
	var m map[string]scoring
	switch platform {
	case "tenhou":
		m = tenhou.Scoring
	case "majsoul":
		m = majsoul.Scoring
	default:
		return scoring{}, fmt.Errorf("unknown platform %q", platform)
	}
	s, ok := m[fmt.Sprint(players)]
	if !ok {
		return scoring{}, fmt.Errorf("no scoring for %s %dp", platform, players)
	}
	return s, nil
}

// ResultPt is the normalized ±pt for one player: (score − return)/1000 + uma[placement], plus oka to
// 1st, rounded to 0.1. Placement is 1-based and must be the platform's own (ties broken by seat).
func ResultPt(platform string, players, score, placement int) (float64, error) {
	if err := checkPlacement(placement, players); err != nil {
		return 0, err
	}
	s, err := scoringFor(platform, players)
	if err != nil {
		return 0, err
	}
	oka := 0
	if placement == 1 {
		oka = (s.ReturnPoints - s.StartPoints) * players / 100
	}
	tenths := int(math.Round(float64(score-s.ReturnPoints)/100)) + s.Uma[placement-1]*10 + oka
	return float64(tenths) / 10, nil
}

// ---------------------------------------------------------------- Tenhou

type TenhouRank struct {
	Rank int // index into ranks (0 = 新人 … 19 = 十段, 20 = 天鳳位)
	Pt   int
}

type TenhouGame struct {
	Players   int
	Level     string // ippan | joukyuu | tokujou | houou
	Length    string // east | south
	Placement int
}

type TenhouChange struct {
	Before, After TenhouRank
	Delta         int
	Event         RankEvent
}

func tenhouRankAt(rank int) (tenhouRank, error) {
	if rank < 0 || rank >= len(tenhou.Ranks) {
		return tenhouRank{}, fmt.Errorf("invalid tenhou rank %d", rank)
	}
	return tenhou.Ranks[rank], nil
}

// TenhouRankName returns the rank's display name in "en" or "ja".
func TenhouRankName(rank int, lang string) string {
	r, err := tenhouRankAt(rank)
	if err != nil {
		return ""
	}
	if lang == "ja" {
		return r.JA
	}
	return r.EN
}

// TenhouRankDelta is the rank-point change for one ranked Tenhou game.
func TenhouRankDelta(rank int, g TenhouGame) (int, error) {
	if err := checkPlacement(g.Placement, g.Players); err != nil {
		return 0, err
	}
	r, err := tenhouRankAt(rank)
	if err != nil {
		return 0, err
	}
	if g.Placement == g.Players {
		if g.Length == "south" {
			return r.LastPlace.South, nil
		}
		return r.LastPlace.East, nil
	}
	table, ok := tenhou.PlacementPoints[fmt.Sprint(g.Players)][g.Level][g.Length]
	if !ok {
		return 0, fmt.Errorf("no tenhou placement points for %dp %s %s", g.Players, g.Level, g.Length)
	}
	return table[g.Placement-1], nil
}

// ApplyTenhouGame applies one game: promotion at ≥ promotePt, demotion below 0 (dan only); kyu ranks
// floor at 0.
func ApplyTenhouGame(s TenhouRank, g TenhouGame) (TenhouChange, error) {
	r, err := tenhouRankAt(s.Rank)
	if err != nil {
		return TenhouChange{}, err
	}
	if r.PromotePt == nil { // 天鳳位
		return TenhouChange{Before: s, After: s}, nil
	}
	delta, err := TenhouRankDelta(s.Rank, g)
	if err != nil {
		return TenhouChange{}, err
	}
	pt := s.Pt + delta
	c := TenhouChange{Before: s, Delta: delta, After: TenhouRank{Rank: s.Rank, Pt: pt}}
	switch {
	case pt >= *r.PromotePt:
		c.After = TenhouRank{Rank: s.Rank + 1, Pt: tenhou.Ranks[s.Rank+1].StartPt}
		c.Event = Promotion
	case pt < 0 && r.CanDemote:
		c.After = TenhouRank{Rank: s.Rank - 1, Pt: tenhou.Ranks[s.Rank-1].StartPt}
		c.Event = Demotion
	case pt < 0:
		c.After.Pt = 0
	}
	return c, nil
}

var tenhouLogStart = regexp.MustCompile(`^(\d{10})gm-`)
var jst = time.FixedZone("JST", 9*3600)

// TenhouSupportedFrom is when the modelled Tenhou rules took effect; older logs are declined (D-046).
func TenhouSupportedFrom() time.Time { return tenhou.SupportedFrom }

// TenhouLogSupported reports whether a Tenhou log ID (starting YYYYMMDDHH, JST) is on or after
// TenhouSupportedFrom.
func TenhouLogSupported(logID string) bool {
	m := tenhouLogStart.FindStringSubmatch(logID)
	if m == nil {
		return false
	}
	t, err := time.ParseInLocation("2006010215", m[1], jst)
	if err != nil {
		return false
	}
	return !t.Before(tenhou.SupportedFrom)
}

// ---------------------------------------------------------------- Mahjong Soul

// MajsoulRank uses Amae Koromo-style level ids: players digit (1 = 4p, 2 = 3p) × 10000 + major × 100 + minor.
type MajsoulRank struct {
	LevelID int
	Points  int
}

type MajsoulGame struct {
	Mode      int // room mode id, see data/majsoul.json
	Score     int // final score
	Placement int
}

type MajsoulChange struct {
	Before, After MajsoulRank
	Delta         int
	Event         RankEvent
}

const kontenFallbackLevel = 503 // Saint 3: Konten playing outside Throne rooms scores like this

type level struct{ p, major, minor int }

func parseLevel(id int) level {
	real := id % 10000
	return level{p: id / 10000, major: real / 100, minor: real % 100}
}

func (l level) id() int { return l.p*10000 + l.major*100 + l.minor }

func oldKontenMajor() int { return majsoul.Konten.Major - 1 }

func (l level) konten() bool { return l.major >= oldKontenMajor() }

func (l level) maxPoints() (int, error) {
	if l.konten() {
		if l.minor == majsoul.Konten.TopMinor {
			return 0, nil
		}
		return majsoul.Konten.MaxPoints, nil
	}
	i := (l.major-1)*3 + l.minor - 1
	if l.major < 1 || l.minor < 1 || l.minor > 3 || i >= len(majsoul.LevelMaxPoints) {
		return 0, fmt.Errorf("invalid majsoul level %d", l.id())
	}
	return majsoul.LevelMaxPoints[i], nil
}

func (l level) startPoints() int {
	if l.major == 1 {
		return 0
	}
	m, _ := l.maxPoints()
	return m / 2
}

func (l level) versionAdjusted() level {
	if l.major == oldKontenMajor() {
		return level{p: l.p, major: majsoul.Konten.Major, minor: 1}
	}
	return l
}

func (l level) versionAdjustedPoints(points int) int {
	if l.major == oldKontenMajor() {
		return int(math.Ceil(float64(points)/100))*10 + 200
	}
	return points
}

func (l level) next() level {
	a := l.versionAdjusted()
	major, minor := a.major, a.minor+1
	if minor > 3 && !a.konten() {
		major++
		minor = 1
	}
	if major == oldKontenMajor() {
		major = majsoul.Konten.Major
	}
	return level{p: a.p, major: major, minor: minor}
}

func (l level) previous() level {
	if l.major == 1 && l.minor == 1 {
		return l
	}
	a := l.versionAdjusted()
	major, minor := a.major, a.minor-1
	if minor < 1 {
		major--
		minor = 3
	}
	if major == oldKontenMajor() {
		major = oldKontenMajor() - 1
	}
	return level{p: a.p, major: major, minor: minor}
}

// MajsoulLevelName returns e.g. "Master 2" / "雀豪 2".
func MajsoulLevelName(id int, lang string) string {
	l := parseLevel(id)
	for _, m := range majsoul.Majors {
		if m.Major != l.major {
			continue
		}
		name := m.EN
		if lang == "ja" {
			name = m.JA
		}
		if l.major == oldKontenMajor() {
			return name
		}
		return fmt.Sprintf("%s %d", name, l.minor)
	}
	return ""
}

func majsoulModeFor(mode int) (majsoulMode, error) {
	m, ok := majsoul.Modes[fmt.Sprint(mode)]
	if !ok {
		return majsoulMode{}, fmt.Errorf("unsupported majsoul mode %d", mode)
	}
	return m, nil
}

// MajsoulRankDelta is the rank-point change for one ranked Mahjong Soul game (last place includes
// the level penalty).
func MajsoulRankDelta(levelID int, g MajsoulGame) (int, error) {
	l := parseLevel(levelID)
	mode, err := majsoulModeFor(g.Mode)
	if err != nil {
		return 0, err
	}
	if err := checkPlacement(g.Placement, mode.Players); err != nil {
		return 0, err
	}
	i := g.Placement - 1
	if l.konten() {
		if mode.KontenDelta != nil {
			return mode.KontenDelta[i], nil
		}
		l = parseLevel(l.p*10000 + kontenFallbackLevel)
	}
	s, err := scoringFor("majsoul", mode.Players)
	if err != nil {
		return 0, err
	}
	delta := ceilDiv(g.Score-s.StartPoints+s.Uma[i]*1000, 1000) + mode.PlacementBonus[i]
	if g.Placement == mode.Players {
		pi := (l.major-1)*3 + l.minor - 1
		pen := majsoul.LastPlacePenalty[mode.PenaltyTable]
		if pi < 0 || pi >= len(pen) {
			return 0, fmt.Errorf("invalid majsoul level %d", levelID)
		}
		delta -= pen[pi]
	}
	return delta, nil
}

// ApplyMajsoulGame applies one game, mirroring Mahjong Soul / Amae Koromo level adjustment.
func ApplyMajsoulGame(s MajsoulRank, g MajsoulGame) (MajsoulChange, error) {
	delta, err := MajsoulRankDelta(s.LevelID, g)
	if err != nil {
		return MajsoulChange{}, err
	}
	orig := parseLevel(s.LevelID)
	points := orig.versionAdjustedPoints(s.Points + delta)
	l := orig.versionAdjusted()
	max, err := l.maxPoints()
	if err != nil {
		return MajsoulChange{}, err
	}
	ev := NoEvent
	switch {
	case max != 0 && points >= max:
		l = l.next()
		points = l.startPoints()
		ev = Promotion
	case points < 0:
		if max == 0 || l.major == 1 || (l.major == 2 && l.minor == 1) {
			points = 0
		} else {
			l = l.previous()
			points = l.startPoints()
			ev = Demotion
		}
	}
	return MajsoulChange{Before: s, After: MajsoulRank{LevelID: l.id(), Points: points}, Delta: delta, Event: ev}, nil
}

// ceilDiv is ceil(a/b) for b > 0.
func ceilDiv(a, b int) int {
	q := a / b
	if a%b != 0 && a > 0 {
		q++
	}
	return q
}
