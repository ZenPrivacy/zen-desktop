package app

type updateEventKind string

const (
	updateChannel                   = "app:update"
	updateAvailable updateEventKind = "updateAvailable"
)

type updateEvent struct {
	Kind updateEventKind `json:"kind"`
}

func (e *frontendEvents) OnUpdateAvailable() {
	e.emit(updateChannel, updateEvent{Kind: updateAvailable})
}
