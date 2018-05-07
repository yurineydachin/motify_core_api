package resources

type IResource interface {
	Caption() string
	Status() Status
}

type Status struct {
	Header []string    `json:"header,omitempty"`
	Rows   []StatusRow `json:"rows,omitempty"`
}

type StatusRow struct {
	Level string   `json:"level,omitempty"`
	Data  []string `json:"data,omitempty"`
}

const (
	ResourceStatusOK     = "success"
	ResourceStatusFail   = "danger"
	ResourceConnected    = "connected"
	ResourceNotConnected = "not connected"
)
