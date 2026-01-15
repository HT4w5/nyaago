package analyzer

import (
	"github.com/HT4w5/nyaago/internal/rulelist"
	"github.com/HT4w5/nyaago/pkg/dto"
)

type Analyzer interface {
	Process(request dto.Request) error
	Report(tx *rulelist.Tx) error
}
