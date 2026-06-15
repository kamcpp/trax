package execpl

import ()

type OutboundSignal struct {
	Type   string            `json:"type"`
	Output map[string]*Value `json:"output"`
	Extra  map[string]string `json:"extra"`
}
