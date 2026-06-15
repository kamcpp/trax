package execpl

import ()

type FIXRequest struct {
	SessionId string `json:"session_id"`
	Content   string `json:"content"`
}

type RESTRequest struct {
	Method       string            `json:"method"`
	Path         string            `json:"path"`
	Headers      map[string]string `json:"headers"`
	BodyEncoding string            `json:"body_endoding"`
	Body         string            `json:"body"`
}

type SessionInfo struct {
	SessionId    string `json:"session_id"`
	AuthProvider string `json:"auth_provider"`
	TokenType    string `json:"token_type"`
	Token        string `json:"token"`
	Identity     string `json:"identity"`
}

type Command struct {
	Id            string            `json:"id"`
	Timestamp     string            `json:"timestamp"`
	Origin        string            `json:"origin"`
	RequestId     string            `json:"request_id"`
	Request       interface{}       `json:"request"`
	ParticipantId string            `json:"participant_id"`
	Session       *SessionInfo      `json:"session_info"`
	Operation     string            `json:"operation"`
	Arguments     map[string]*Value `json:"arguments"`
	IsEffectful   bool              `json:"is_effectful"`
	IsIdempotent  bool              `json:"is_idempotent"`
	IsDryRun      bool              `json:"is_dry_run"`
	Extra         map[string]string `json:"extra"`
}
