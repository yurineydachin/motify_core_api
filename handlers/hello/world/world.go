package world

type Handler struct{}

func New() *Handler {
	return &Handler{}
}

func (*Handler) Caption() string {
	return "Example handler"
}

func (*Handler) Description() string {
	return "Returns 'Hello world' string"
}
