package main

import (
	"errors"
	"fmt"
	"sync"
)

type StateType string
type EventType string

type EventContext any

type Actions interface {
	Execute(eventCtx EventContext) EventType
}

type States map[StateType]EventObj
type Events map[EventType]StateType

type EventObj struct {
	action Actions
	event  Events
}

type StateMachine struct {
	currentState StateType
	currentEvent EventType
	states       States

	mu sync.Mutex
}

func (s *StateMachine) getNextState(event EventType) (StateType, error) {
	state, ok := s.states[s.currentState]
	if !ok {
		return def, errors.New("missing state")
	}

	next, ok := state.event[event]
	if !ok {
		return def, errors.New("missing event")
	}

	return next, nil
}

func (s *StateMachine) SendEvent(event EventType, eventCtx EventContext) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	nextState, err := s.getNextState(event)
	if err != nil {
		return fmt.Errorf("can't get next state via SendEvent: %w", err)
	}

	state, ok := s.states[nextState]
	if !ok || state.action == nil {
		return errors.New("configuration error")
	}

	s.currentEvent = state.action.Execute(eventCtx)
	s.currentState = nextState

	return nil
}

// States
const (
	def = StateType("def")

	addingStickerState = StateType("addingSticker")
	addingTagsState    = StateType("addingTags")

	showingTagsState = StateType("showingTags")

	deletingStickerState = StateType("deletingSticker")
)

// Events
const (
	nop = EventType("nop")

	addStickerEvent = EventType("addSticker")
	addTagsEvent    = EventType("addTags")

	showTagsEvent = EventType("showTags")

	deleteStickerEvent = EventType("deleteSticker")
)

func newStickerFSM() *StateMachine {
	return &StateMachine{
		states: States{
			def: EventObj{
				action: &defaultStickerAction{},
				event: Events{
					nop:                def,
					addStickerEvent:    addingStickerState,
					showTagsEvent:      showingTagsState,
					deleteStickerEvent: deletingStickerState,
					addTagsEvent:       addingTagsState,
				},
			},

			addingStickerState: EventObj{
				action: &addingStickerAction{},
				event: Events{
					addStickerEvent: addingStickerState,
					addTagsEvent:    addingTagsState,
					nop:             def,
				},
			},
			addingTagsState: EventObj{
				action: &addingTagsAction{},
				event: Events{
					addTagsEvent: addingTagsState,
					nop:          def,
				},
			},

			showingTagsState: EventObj{
				action: &showingTagsAction{},
				event: Events{
					showTagsEvent: showingTagsState,
					nop:           def,
				},
			},

			deletingStickerState: EventObj{
				action: &deletingStickerAction{},
				event: Events{
					deleteStickerEvent: deletingStickerState,
					nop:                def,
				},
			},
		},
		currentState: def,
	}
}
