package services

import "time"

type State int

const (
	StateRoot State = iota

	StatePlanMenu
	StatePlanAddingAwaitDesc
	StatePlanAddingAwaitEventTime
	StatePlanAddingAwaitRemindTime

	StateSettingsMenu
	StateSettingsCity
	StateSettingsPartner
)

type Session struct {
	State      State
	TempDesc   string
	TempEvent  time.Time
	TempRemind time.Time
	TempPage   int
	TempPlanID int64
}

type SessionManager struct {
	sessions map[int64]*Session
}

func NewSessionManager() *SessionManager {
	return &SessionManager{sessions: make(map[int64]*Session)}
}

func (m *SessionManager) Get(chatID int64) *Session {
	if s, ok := m.sessions[chatID]; ok {
		return s
	}
	s := &Session{State: StateRoot}
	m.sessions[chatID] = s
	return s
}

func (m *SessionManager) Reset(chatID int64) {
	delete(m.sessions, chatID)
}

func (m *SessionManager) IsPlanState(chatID int64) bool {
	state := m.sessions[chatID].State
	return state >= StatePlanMenu && state <= StatePlanAddingAwaitRemindTime
}

func (m *SessionManager) IsSettingsState(chatID int64) bool {
	state := m.sessions[chatID].State
	return state >= StateSettingsMenu && state <= StateSettingsPartner
}
