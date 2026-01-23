package analyzer

import (
	"context"

	"github.com/HT4w5/nyaago/internal/rulelist"
	"github.com/HT4w5/nyaago/pkg/dto"
)

type Analyzer interface {
	Start(ctx context.Context) error   // Start analyzer internal goroutines
	Process(request dto.Request) error // Called when processing request
	Report(tx *rulelist.Tx) error      // Called when generating a ruleset
}
