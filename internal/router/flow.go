package router

import (
	"errors"
	"fmt"

	"github.com/HT4w5/nyaago/internal/config"
	"github.com/HT4w5/nyaago/pkg/dto"
)

var (
	errExitFlow = errors.New("exit flow")
)

type flow interface {
	Run(request *dto.Request) error
}

func makeFlow(cfg config.FlowConfig) (flow, error) {
	var f flow
	// Build subflow
	subflow := make([]flow, 0)
	for _, v := range cfg.Subflow {
		f, err := makeFlow(v)
		if err != nil {
			return nil, err
		}
		subflow = append(subflow, f)
	}
	switch cfg.Action {
	case "sequence":
		f = &sequence{
			subflow: subflow,
		}
	case "match":
		p := cfg.Params.(config.MatchParams)
		matchers := make([]matcher, len(p.Matchers))
		for _, v := range p.Matchers {
			matchers = append(matchers, makeMatcher(v))
		}
		f = &match{
			matchers: matchers,
			subflow:  subflow,
		}
	default:
		return nil, fmt.Errorf("bad action %s", cfg.Action)
	}

	return f, nil
}

// --- Begin "sequence" ---
type sequence struct {
	subflow []flow
}

func (f *sequence) Run(request *dto.Request) error {
	return runSubflow(f.subflow, request)
}

// --- End "sequence" ---

// --- Begin "match" ---

// Match is a flow that runs its subflow if request satisfies one of its conditions (matchers)
type match struct {
	matchers []matcher
	subflow  []flow
}

type matcher struct {
	filters []func(*dto.Request) bool
}

// Only returns true if all filters are satisfied
func (m matcher) Match(request *dto.Request) bool {
	for _, v := range m.filters {
		if !v(request) {
			return false
		}
	}
	return true
}

type filter struct {
	fn       func(*dto.Request) bool
	priority int
}

const (
	filterPrioHigh = iota // Match faster checks first for better performance
	filterPrioMedium
	filterPrioLow
)

func makeMatcher(cfg config.MatcherConfig) matcher {
	var active []filter

	// filterPrioHigh
	if cfg.Status != nil {
		val := *cfg.Status
		active = append(active, filter{priority: filterPrioHigh, fn: func(r *dto.Request) bool {
			return r.Status == val
		}})
	}
	if cfg.Method != nil {
		val := *cfg.Method
		active = append(active, filter{priority: filterPrioHigh, fn: func(r *dto.Request) bool {
			return r.Method == val
		}})
	}
	if cfg.SentMin != nil {
		val := *cfg.SentMin
		active = append(active, filter{priority: filterPrioHigh, fn: func(r *dto.Request) bool {
			return r.Sent >= val
		}})
	}
	if cfg.SentMax != nil {
		val := *cfg.SentMax
		active = append(active, filter{priority: filterPrioHigh, fn: func(r *dto.Request) bool {
			return r.Sent <= val
		}})
	}

	// filterPrioMedium
	if cfg.Client != nil {
		val := *cfg.Client
		active = append(active, filter{priority: filterPrioMedium, fn: func(r *dto.Request) bool {
			return val.Contains(r.Client)
		}})
	}
	if cfg.Server != nil {
		val := *cfg.Server
		active = append(active, filter{priority: filterPrioMedium, fn: func(r *dto.Request) bool {
			return val.Contains(r.Server)
		}})
	}

	// filterPrioLow
	if cfg.URL != nil {
		re := cfg.URL
		active = append(active, filter{priority: filterPrioLow, fn: func(r *dto.Request) bool {
			return re.MatchString(r.URL)
		}})
	}
	if cfg.Host != nil {
		re := cfg.Host
		active = append(active, filter{priority: filterPrioLow, fn: func(r *dto.Request) bool {
			return re.MatchString(r.Host)
		}})
	}
	if cfg.Agent != nil {
		re := cfg.Agent
		active = append(active, filter{priority: filterPrioLow, fn: func(r *dto.Request) bool {
			return re.MatchString(r.Agent)
		}})
	}

	// Sort by priority and extract functions
	final := make([]func(*dto.Request) bool, 0, len(active))
	for p := filterPrioHigh; p <= filterPrioLow; p++ {
		for _, item := range active {
			if item.priority == p {
				final = append(final, item.fn)
			}
		}
	}

	return matcher{
		filters: final,
	}
}

func (f *match) Run(request *dto.Request) error {
	// Check conditions
	match := false
	for _, v := range f.matchers {
		if v.Match(request) {
			match = true
			break
		}
	}
	if !match {
		return nil
	}

	return runSubflow(f.subflow, request)
}

// --- End "match" ---

func runSubflow(subflow []flow, request *dto.Request) error {
	for _, v := range subflow {
		err := v.Run(request)
		if err != nil {
			return err
		}
	}
	return nil
}
