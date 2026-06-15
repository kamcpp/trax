package execpl

import ()

type Envelope struct {
	Command *Command    `json:"command"`
	Object  interface{} `json:"object"`
}
