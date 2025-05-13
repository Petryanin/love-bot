package services

import "time"

type State int

const (
    StateMenu State = iota
    StateAddingAwaitDesc
    StateAddingAwaitEventTime
    StateAddingAwaitRemindTime
)

type Session struct {
    State       State
    TempDesc    string
    TempEvent   time.Time
    TempRemind  time.Time
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
    s := &Session{State: StateMenu}
    m.sessions[chatID] = s
    return s
}

func (m *SessionManager) Reset(chatID int64) {
    delete(m.sessions, chatID)
}
