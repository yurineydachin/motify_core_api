package payslip_set

import (
	"context"
	//"time"

	"github.com/sergei-svistunov/gorpc/transport/cache"
	"godep.lzd.co/service/logger"
)

// v1Args contains a request arguments
type v1Args struct {
	ID               *uint64     `key:"payslip_id" description:"Payslip id"`
	Employee         Person      `key:"employee" description:"Employee"`
	ProcessedByUser  *Person     `key:"processed_by_user" description:"Processed by user"`
	Agent            Company     `key:"agent" description:"Agent id"`
	ProcessedByAgent *Company    `key:"processed_by_agent" description:"Processed by agent"`
	Data             PayslipData `key:"data" description:"Payslip data"`
}

type PayslipData struct {
	Currency string    `key:"currency" description:"Currency"`
	Payslip  Operation `key:"payslip" description:"Payslip items"`
	Details  Operation `key:"details" description:"Details items"`
}

type Operation struct {
	Amount  *float64     `key:"amount" description:"Amount"`
	Float   *float64     `key:"float" description:"Float number"`
	Int     *int64       `key:"int" description:"Integer number"`
	Text    *string      `key:"text" description:"Text"`
	Title   *string      `key:"title" description:"Title"`
	Details *[]Operation `key:"details" description:"Detail page info"`
}

type Person struct {
	ID          *uint64  `key:"id" description:"ID"`
	Name        string   `key:"name" description:"Name"`
	RoleDesc    string   `key:"p_description" description:"Role and Company"`
	Description string   `key:"description" description:"Description"`
	Contacts    Contacts `key:"contacts" description:"Contacts"`
	Avatar      string   `key:"awatar" description:"Avatar url icon"`
}

type Company struct {
	ID          *uint64  `key:"id" description:"ID"`
	Name        string   `key:"name" description:"Name"`
	RoleDesc    string   `key:"c_description" description:"Company short description"`
	Description string   `key:"description" description:"Description"`
	Contacts    Contacts `key:"contacts" description:"Contacts"`
	BGImage     string   `key:"bg_image" description:"Background image url"`
	Logo        string   `key:"logo" description:"Logo url icon"`
}

type Contacts struct {
	Address string `key:"address" description:"Address"`
	Phone   string `key:"phone" description:"Phone"`
	Email   string `key:"email" description:"Email"`
	Site    string `key:"site" description:"Site"`
}

func (*Handler) V1(ctx context.Context, opts *v1Args) (string, error) {
	logger.Debug(ctx, "Payslip SET")
	cache.EnableTransportCache(ctx)
	return "OK", nil
}
