package payslip_set

type Handler struct{}

func New() *Handler {
	return &Handler{}
}

func (*Handler) Caption() string {
	return "Add and update payslip"
}

func (*Handler) Description() string {
	return "Save changes of payslip"
}
