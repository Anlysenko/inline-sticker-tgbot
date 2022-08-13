package main

import (
	"errors"
	"fmt"
	"sync"
)

const (
	Def = StateType("")
	Nop = EventType("nop")
)

type StateType string
type EventType string

type EventContext any
type Actions interface {
	Execute(eventCtx EventContext) EventType
}

type State struct {
	Action Actions
	Event  Events
}

type States map[StateType]State
type Events map[EventType]StateType

type StateMachine struct {
	Previous StateType
	Current  StateType
	States   States

	mu sync.Mutex
}

func (s *StateMachine) getNextState(event EventType) (StateType, error) {
	state, ok := s.States[s.Current]
	if !ok {
		return Def, errors.New("missing state")
	}

	next, ok := state.Event[event]
	if !ok {
		return Def, errors.New("missing event")
	}

	return next, nil
}

func (s *StateMachine) SendEvent(event EventType, eventCtx EventContext) (StateType, EventType, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	nextState, err := s.getNextState(event)
	if err != nil {
		return Def, Nop, fmt.Errorf("can't get next state via SendEvent: %w", err)
	}

	state, ok := s.States[nextState]
	if !ok || state.Action == nil {
		return Def, Nop, errors.New("configuration error")
	}

	s.Previous = s.Current
	s.Current = nextState

	nextEvent := state.Action.Execute(eventCtx)
	if nextEvent == Nop {
		return Def, Nop, nil
	}

	return s.Current, nextEvent, nil
}

const (
	addingSticker            = StateType("addingSticker")
	sendingToAddSticker      = StateType("sendingToAddSticker")
	sendingToAddDescription  = StateType("sendingToAddDescription")
	deletingSticker          = StateType("deletingSticker")
	sendingToDeleteSticker   = StateType("sendingToDeleteSticker")
	showingDescription       = StateType("showingDescription")
	sendingToShowDescription = StateType("sendingToShowDescription")

	addSticker            = EventType("addSticker")
	sendToAddSticker      = EventType("sendToAddSticker")
	sendToAddDescription  = EventType("sendToAddDescription")
	deleteSticker         = EventType("deleteSticker")
	sendToDeleteSticker   = EventType("sendToDeleteSticker")
	showDescription       = EventType("showDescription")
	sendToShowDescription = EventType("sendToShowDescription")
)

func processStickerFSM(current StateType) *StateMachine {
	return &StateMachine{
		States: States{
			Def: State{
				Event: Events{
					addSticker:      addingSticker,
					deleteSticker:   deletingSticker,
					showDescription: showingDescription,
				},
			},

			addingSticker: State{
				Action: &addingStickerAction{},
				Event: Events{
					sendToAddSticker: sendingToAddSticker,
				},
			},
			sendingToAddSticker: State{
				Action: &sendingStickerAction{},
				Event: Events{
					sendToAddSticker:     sendingToAddSticker,
					sendToAddDescription: sendingToAddDescription,
				},
			},
			sendingToAddDescription: State{
				Action: &sendingDescriptionAction{},
				Event: Events{
					sendToAddDescription: sendingToAddDescription,
					Nop:                  Def,
				},
			},

			deletingSticker: State{
				Action: &deletingStickerAction{},
				Event: Events{
					sendToDeleteSticker: sendingToDeleteSticker,
				},
			},
			sendingToDeleteSticker: State{
				Action: &sendingToDeleteStickerAction{},
				Event: Events{
					sendToDeleteSticker: sendingToDeleteSticker,
					Nop:                 Def,
				},
			},

			showingDescription: State{
				Action: &showingDescriptionAction{},
				Event: Events{
					sendToShowDescription: sendingToShowDescription,
				},
			},
			sendingToShowDescription: State{
				Action: &sendingToShowDescriptionAction{},
				Event: Events{
					sendToShowDescription: sendingToShowDescription,
					Nop:                   Def,
				},
			},
		},
		Current: current,
	}
}
