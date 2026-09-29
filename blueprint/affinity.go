package blueprint

import (
	_ "embed"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"unicode"
)

type Manifest struct {
	System           string  `json:"system"`
	Blueprint        string  `json:"blueprint"`
	BlueprintVersion string  `json:"blueprint_version"`
	MappingPolicy    string  `json:"mapping_policy"`
	NonDestructive   bool    `json:"non_destructive"`
	EvidenceRule     string  `json:"evidence_rule"`
	Entries          []Entry `json:"entries"`
}

type Entry struct {
	ID           string     `json:"id"`
	Family       string     `json:"family"`
	Description  string     `json:"description"`
	Affinities   []Affinity `json:"affinities"`
	CoverageNote string     `json:"coverage_note,omitempty"`
}

type Affinity struct {
	Owner      string `json:"owner"`
	Capability string `json:"capability"`
	Path       string `json:"path"`
	Status     string `json:"status"`
}

type Match struct {
	EntryID     string
	Family      string
	Description string
	Score       float64
	Affinities  []Affinity
}

type RoutePlan struct {
	Query      string
	Primary    Match
	Supporting []Match
	Preserved  []Affinity
}

var (
	//go:embed atlas_affinity.json
	manifestData []byte
	defaultManifest Manifest
)

func init() {
	if err := json.Unmarshal(manifestData, &defaultManifest); err != nil {
		panic("blueprint: invalid embedded atlas affinity manifest: " + err.Error())
	}
}

func DefaultManifest() Manifest {
	return cloneManifest(defaultManifest)
}

func Resolve(query string, limit int) ([]Match, error) {
	normalized := normalize(query)
	if normalized == "" {
		return nil, errors.New("blueprint query is required")
	}
	if limit <= 0 {
		limit = 8
	}
	qTokens := tokenSet(normalized)
	matches := make([]Match, 0, len(defaultManifest.Entries))
	for _, entry := range defaultManifest.Entries {
		score := affinityScore(qTokens, tokenSet(strings.ToLower(entry.ID+" "+entry.Family+" "+entry.Description)))
		if score == 0 {
			continue
		}
		matches = append(matches, Match{
			EntryID: entry.ID, Family: entry.Family, Description: entry.Description,
			Score: score, Affinities: append([]Affinity(nil), entry.Affinities...),
		})
	}
	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].Score != matches[j].Score {
			return matches[i].Score > matches[j].Score
		}
		return matches[i].EntryID < matches[j].EntryID
	})
	if len(matches) > limit {
		matches = matches[:limit]
	}
	if len(matches) == 0 {
		return nil, errors.New("no registered blueprint affinity matched the query")
	}
	return matches, nil
}

func ResolveExecutable(query string, limit int) ([]Match, error) {
	matches, err := Resolve(query, limit)
	if err != nil {
		return nil, err
	}
	for i := range matches {
		executable := matches[i].Affinities[:0]
		for _, affinity := range matches[i].Affinities {
			switch strings.ToUpper(strings.TrimSpace(affinity.Status)) {
			case "IMPLEMENTED", "LOCAL_EXECUTABLE", "EXECUTABLE", "BUILTIN_RUNTIME", "REAL", "CONNECTED", "CANONICAL_OWNER", "REGISTERED", "DECLARED_EXECUTABLE_BOUNDARY":
				executable = append(executable, affinity)
			}
		}
		matches[i].Affinities = executable
	}
	matches = filterMatchesWithAffinities(matches)
	if len(matches) == 0 {
		return nil, errors.New("no executable blueprint affinity matched the query")
	}
	return matches, nil
}

func filterMatchesWithAffinities(matches []Match) []Match {
	out := matches[:0]
	for _, match := range matches {
		if len(match.Affinities) > 0 {
			out = append(out, match)
		}
	}
	return out
}

func ResolveFamily(family string) (Match, error) {
	want := strings.TrimSpace(strings.ToUpper(family))
	if want == "" {
		return Match{}, errors.New("blueprint family is required")
	}
	for _, entry := range defaultManifest.Entries {
		if strings.EqualFold(entry.Family, want) {
			return Match{
				EntryID: entry.ID, Family: entry.Family, Description: entry.Description,
				Score: 1, Affinities: append([]Affinity(nil), entry.Affinities...),
			}, nil
		}
	}
	return Match{}, errors.New("blueprint family not found: " + family)
}

func Compose(query string, limit int) (RoutePlan, error) {
	matches, err := Resolve(query, limit)
	if err != nil {
		return RoutePlan{Query: query}, err
	}
	primary := matches[0]
	supporting := make([]Match, 0, len(matches)-1)
	seen := map[string]struct{}{}
	preserved := make([]Affinity, 0)
	for _, a := range primary.Affinities {
		key := a.Owner + "/" + a.Capability
		seen[key] = struct{}{}
		preserved = append(preserved, a)
	}
	for _, item := range matches[1:] {
		overlap := false
		for _, a := range item.Affinities {
			key := a.Owner + "/" + a.Capability
			if _, ok := seen[key]; ok {
				overlap = true
				continue
			}
			seen[key] = struct{}{}
			preserved = append(preserved, a)
		}
		if !overlap {
			supporting = append(supporting, item)
		}
	}
	return RoutePlan{Query: query, Primary: primary, Supporting: supporting, Preserved: preserved}, nil
}

func ValidateManifest() error {
	owners := map[string]string{}
	seenEntries := map[string]struct{}{}
	for _, entry := range defaultManifest.Entries {
		if strings.TrimSpace(entry.ID) == "" || strings.TrimSpace(entry.Family) == "" {
			return errors.New("manifest contains entry without id/family")
		}
		if _, duplicate := seenEntries[entry.ID]; duplicate {
			return errors.New("duplicate blueprint entry: " + entry.ID)
		}
		seenEntries[entry.ID] = struct{}{}
		for _, affinity := range entry.Affinities {
			key := affinity.Owner + "/" + affinity.Capability
			if strings.TrimSpace(affinity.Owner) == "" || strings.TrimSpace(affinity.Capability) == "" || strings.TrimSpace(affinity.Path) == "" {
				return errors.New("manifest contains incomplete affinity: " + key)
			}
			if existing, ok := owners[affinity.Capability]; ok && existing != affinity.Owner {
				return errors.New("capability has parallel canonical owners: " + affinity.Capability + " -> " + existing + ", " + affinity.Owner)
			}
			owners[affinity.Capability] = affinity.Owner
		}
	}
	return nil
}

func normalize(value string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(value) {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			b.WriteRune(r)
		} else {
			b.WriteByte(' ')
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

func tokenSet(value string) map[string]struct{} {
	out := make(map[string]struct{})
	for _, token := range strings.Fields(normalize(value)) {
		if len(token) < 3 {
			continue
		}
		out[token] = struct{}{}
	}
	return out
}

func affinityScore(query, descriptor map[string]struct{}) float64 {
	if len(query) == 0 || len(descriptor) == 0 {
		return 0
	}
	hits := 0
	for token := range query {
		if _, ok := descriptor[token]; ok {
			hits++
		}
	}
	return float64(hits) / float64(len(query))
}

func cloneManifest(src Manifest) Manifest {
	out := src
	out.Entries = make([]Entry, 0, len(src.Entries))
	for _, entry := range src.Entries {
		copyEntry := entry
		copyEntry.Affinities = append([]Affinity(nil), entry.Affinities...)
		out.Entries = append(out.Entries, copyEntry)
	}
	return out
}
