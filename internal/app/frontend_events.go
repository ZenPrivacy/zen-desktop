package app

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type frontendEvents struct {
	ctx context.Context
}

func newFrontendEvents(ctx context.Context) *frontendEvents {
	return &frontendEvents{ctx: ctx}
}

func (e *frontendEvents) emit(channel string, payload any) {
	runtime.EventsEmit(e.ctx, channel, payload)
}
