package args

import (
	"errors"

	"github.com/alecthomas/kong"
)

type NonrootPort uint16

func (NonrootPort) BeforeApply(ctx *kong.Context, trace *kong.Path) error {
	p := uint16(ctx.FlagValue(trace.Flag).(NonrootPort))
	if p < 1024 {
		return errors.New("only ports 1024-65535 are valid")
	}

	return nil
}
