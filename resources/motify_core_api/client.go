// It's auto-generated file. It's not recommended to modify it.
package adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/mailru/easyjson"
	"github.com/sergei-svistunov/gorpc/transport/cache"
	"io/ioutil"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"time"
)

type IBalancer interface {
	Next() (string, error)
}

type Callbacks struct {
	OnStart                func(ctx context.Context, req *http.Request) context.Context
	OnPrepareRequest       func(ctx context.Context, req *http.Request, data interface{}) context.Context
	OnResponseUnmarshaling func(ctx context.Context, req *http.Request, response *http.Response, result []byte)
	OnSuccess              func(ctx context.Context, req *http.Request, data interface{})
	OnError                func(ctx context.Context, req *http.Request, err error) error
	OnPanic                func(ctx context.Context, req *http.Request, r interface{}, trace []byte) error
	OnFinish               func(ctx context.Context, req *http.Request, startTime time.Time)
}

type MotifyCoreAPIGoRPC struct {
	client      *http.Client
	serviceName string
	balancer    IBalancer
	callbacks   Callbacks
	cache       cache.ICache
}

func (api *MotifyCoreAPIGoRPC) SetCache(c cache.ICache) *MotifyCoreAPIGoRPC {
	api.cache = c
	return api
}

func NewMotifyCoreAPIGoRPC(client *http.Client, balancer IBalancer, callbacks Callbacks) *MotifyCoreAPIGoRPC {
	if client == nil {
		client = http.DefaultClient
	}
	return &MotifyCoreAPIGoRPC{
		//		client: &http.Client{
		//			Transport: &http.Transport{
		//				//DisableCompression: true,
		//				MaxIdleConnsPerHost: 20,
		//			},
		//			Timeout: apiTimeout,
		//		},
		serviceName: "MotifyCoreAPIGoRPC",
		balancer:    balancer,
		callbacks:   callbacks,
		client:      client,
	}
}

func (api *MotifyCoreAPIGoRPC) AgentCreateV1(ctx context.Context, options AgentCreateV1Args) (*AgentCreateV1Res, error) {
	var result *AgentCreateV1Res
	var entry = cache.CacheEntry{Body: &result}
	err := api.setWithCache(ctx, "/agent/create/v1/", options, &entry, _AgentCreateV1ErrorsMapping)
	if result, ok := entry.Body.(**AgentCreateV1Res); ok {
		return *result, err
	}
	return result, err
}

func (api *MotifyCoreAPIGoRPC) AgentListV1(ctx context.Context, options AgentListV1Args) (*AgentListV1Res, error) {
	var result *AgentListV1Res
	var entry = cache.CacheEntry{Body: &result}
	err := api.setWithCache(ctx, "/agent/list/v1/", options, &entry, nil)
	if result, ok := entry.Body.(**AgentListV1Res); ok {
		return *result, err
	}
	return result, err
}

func (api *MotifyCoreAPIGoRPC) AgentUpdateV1(ctx context.Context, options AgentUpdateV1Args) (*AgentUpdateV1Res, error) {
	var result *AgentUpdateV1Res
	var entry = cache.CacheEntry{Body: &result}
	err := api.setWithCache(ctx, "/agent/update/v1/", options, &entry, _AgentUpdateV1ErrorsMapping)
	if result, ok := entry.Body.(**AgentUpdateV1Res); ok {
		return *result, err
	}
	return result, err
}

func (api *MotifyCoreAPIGoRPC) EmployeeCreateV1(ctx context.Context, options EmployeeCreateV1Args) (*EmployeeCreateV1Res, error) {
	var result *EmployeeCreateV1Res
	var entry = cache.CacheEntry{Body: &result}
	err := api.setWithCache(ctx, "/employee/create/v1/", options, &entry, _EmployeeCreateV1ErrorsMapping)
	if result, ok := entry.Body.(**EmployeeCreateV1Res); ok {
		return *result, err
	}
	return result, err
}

func (api *MotifyCoreAPIGoRPC) EmployeeDetailsV1(ctx context.Context, options EmployeeDetailsV1Args) (*EmployeeDetailsV1Res, error) {
	var result *EmployeeDetailsV1Res
	var entry = cache.CacheEntry{Body: &result}
	err := api.setWithCache(ctx, "/employee/details/v1/", options, &entry, _EmployeeDetailsV1ErrorsMapping)
	if result, ok := entry.Body.(**EmployeeDetailsV1Res); ok {
		return *result, err
	}
	return result, err
}

func (api *MotifyCoreAPIGoRPC) EmployeeInviteV1(ctx context.Context, options EmployeeInviteV1Args) (*EmployeeInviteV1Res, error) {
	var result *EmployeeInviteV1Res
	var entry = cache.CacheEntry{Body: &result}
	err := api.setWithCache(ctx, "/employee/invite/v1/", options, &entry, _EmployeeInviteV1ErrorsMapping)
	if result, ok := entry.Body.(**EmployeeInviteV1Res); ok {
		return *result, err
	}
	return result, err
}

func (api *MotifyCoreAPIGoRPC) EmployeeListV1(ctx context.Context, options EmployeeListV1Args) (*EmployeeListV1Res, error) {
	var result *EmployeeListV1Res
	var entry = cache.CacheEntry{Body: &result}
	err := api.setWithCache(ctx, "/employee/list/v1/", options, &entry, nil)
	if result, ok := entry.Body.(**EmployeeListV1Res); ok {
		return *result, err
	}
	return result, err
}

func (api *MotifyCoreAPIGoRPC) EmployeeUpdateV1(ctx context.Context, options EmployeeUpdateV1Args) (*EmployeeUpdateV1Res, error) {
	var result *EmployeeUpdateV1Res
	var entry = cache.CacheEntry{Body: &result}
	err := api.setWithCache(ctx, "/employee/update/v1/", options, &entry, _EmployeeUpdateV1ErrorsMapping)
	if result, ok := entry.Body.(**EmployeeUpdateV1Res); ok {
		return *result, err
	}
	return result, err
}

func (api *MotifyCoreAPIGoRPC) IntegrationCheckV1(ctx context.Context, options IntegrationCheckV1Args) (*IntegrationCheckV1Res, error) {
	var result *IntegrationCheckV1Res
	var entry = cache.CacheEntry{Body: &result}
	err := api.setWithCache(ctx, "/integration/check/v1/", options, &entry, _IntegrationCheckV1ErrorsMapping)
	if result, ok := entry.Body.(**IntegrationCheckV1Res); ok {
		return *result, err
	}
	return result, err
}

func (api *MotifyCoreAPIGoRPC) IntegrationCreateV1(ctx context.Context, options IntegrationCreateV1Args) (*IntegrationCreateV1Res, error) {
	var result *IntegrationCreateV1Res
	var entry = cache.CacheEntry{Body: &result}
	err := api.setWithCache(ctx, "/integration/create/v1/", options, &entry, _IntegrationCreateV1ErrorsMapping)
	if result, ok := entry.Body.(**IntegrationCreateV1Res); ok {
		return *result, err
	}
	return result, err
}

func (api *MotifyCoreAPIGoRPC) PayslipCreateV1(ctx context.Context, options PayslipCreateV1Args) (*PayslipCreateV1Res, error) {
	var result *PayslipCreateV1Res
	var entry = cache.CacheEntry{Body: &result}
	err := api.setWithCache(ctx, "/payslip/create/v1/", options, &entry, _PayslipCreateV1ErrorsMapping)
	if result, ok := entry.Body.(**PayslipCreateV1Res); ok {
		return *result, err
	}
	return result, err
}

func (api *MotifyCoreAPIGoRPC) PayslipDetailsV1(ctx context.Context, options PayslipDetailsV1Args) (*PayslipDetailsV1Res, error) {
	var result *PayslipDetailsV1Res
	var entry = cache.CacheEntry{Body: &result}
	err := api.setWithCache(ctx, "/payslip/details/v1/", options, &entry, _PayslipDetailsV1ErrorsMapping)
	if result, ok := entry.Body.(**PayslipDetailsV1Res); ok {
		return *result, err
	}
	return result, err
}

func (api *MotifyCoreAPIGoRPC) PayslipListV1(ctx context.Context, options PayslipListV1Args) (*PayslipListV1Res, error) {
	var result *PayslipListV1Res
	var entry = cache.CacheEntry{Body: &result}
	err := api.setWithCache(ctx, "/payslip/list/v1/", options, &entry, nil)
	if result, ok := entry.Body.(**PayslipListV1Res); ok {
		return *result, err
	}
	return result, err
}

func (api *MotifyCoreAPIGoRPC) PayslipListByEmployeeV1(ctx context.Context, options PayslipListByEmployeeV1Args) (*PayslipListByEmployeeV1Res, error) {
	var result *PayslipListByEmployeeV1Res
	var entry = cache.CacheEntry{Body: &result}
	err := api.setWithCache(ctx, "/payslip/listByEmployee/v1/", options, &entry, nil)
	if result, ok := entry.Body.(**PayslipListByEmployeeV1Res); ok {
		return *result, err
	}
	return result, err
}

func (api *MotifyCoreAPIGoRPC) SettingCreateV1(ctx context.Context, options SettingCreateV1Args) (*SettingCreateV1Res, error) {
	var result *SettingCreateV1Res
	var entry = cache.CacheEntry{Body: &result}
	err := api.setWithCache(ctx, "/setting/create/v1/", options, &entry, _SettingCreateV1ErrorsMapping)
	if result, ok := entry.Body.(**SettingCreateV1Res); ok {
		return *result, err
	}
	return result, err
}

func (api *MotifyCoreAPIGoRPC) SettingListV1(ctx context.Context, options SettingListV1Args) (*SettingListV1Res, error) {
	var result *SettingListV1Res
	var entry = cache.CacheEntry{Body: &result}
	err := api.setWithCache(ctx, "/setting/list/v1/", options, &entry, nil)
	if result, ok := entry.Body.(**SettingListV1Res); ok {
		return *result, err
	}
	return result, err
}

func (api *MotifyCoreAPIGoRPC) SettingUpdateV1(ctx context.Context, options SettingUpdateV1Args) (*SettingUpdateV1Res, error) {
	var result *SettingUpdateV1Res
	var entry = cache.CacheEntry{Body: &result}
	err := api.setWithCache(ctx, "/setting/update/v1/", options, &entry, _SettingUpdateV1ErrorsMapping)
	if result, ok := entry.Body.(**SettingUpdateV1Res); ok {
		return *result, err
	}
	return result, err
}

func (api *MotifyCoreAPIGoRPC) UserCreateV1(ctx context.Context, options UserCreateV1Args) (*UserCreateV1Res, error) {
	var result *UserCreateV1Res
	var entry = cache.CacheEntry{Body: &result}
	err := api.setWithCache(ctx, "/user/create/v1/", options, &entry, _UserCreateV1ErrorsMapping)
	if result, ok := entry.Body.(**UserCreateV1Res); ok {
		return *result, err
	}
	return result, err
}

func (api *MotifyCoreAPIGoRPC) UserLoginV1(ctx context.Context, options UserLoginV1Args) (*UserLoginV1Res, error) {
	var result *UserLoginV1Res
	var entry = cache.CacheEntry{Body: &result}
	err := api.setWithCache(ctx, "/user/login/v1/", options, &entry, _UserLoginV1ErrorsMapping)
	if result, ok := entry.Body.(**UserLoginV1Res); ok {
		return *result, err
	}
	return result, err
}

func (api *MotifyCoreAPIGoRPC) UserRemindV1(ctx context.Context, options UserRemindV1Args) (*UserRemindV1Res, error) {
	var result *UserRemindV1Res
	var entry = cache.CacheEntry{Body: &result}
	err := api.setWithCache(ctx, "/user/remind/v1/", options, &entry, _UserRemindV1ErrorsMapping)
	if result, ok := entry.Body.(**UserRemindV1Res); ok {
		return *result, err
	}
	return result, err
}

func (api *MotifyCoreAPIGoRPC) UserUpdateV1(ctx context.Context, options UserUpdateV1Args) (*UserUpdateV1Res, error) {
	var result *UserUpdateV1Res
	var entry = cache.CacheEntry{Body: &result}
	err := api.setWithCache(ctx, "/user/update/v1/", options, &entry, _UserUpdateV1ErrorsMapping)
	if result, ok := entry.Body.(**UserUpdateV1Res); ok {
		return *result, err
	}
	return result, err
}

// easyjson:json
type AgentCreateV1Args struct {
	IntegrationFK uint64  `json:"fk_integration"`
	Name          *string `json:"name,omitempty"`
	CompanyID     string  `json:"company_id"`
	Description   *string `json:"description,omitempty"`
	Logo          *string `json:"logo,omitempty"`
	Background    *string `json:"bg_image,omitempty"`
	Phone         *string `json:"phone,omitempty"`
	Email         *string `json:"email,omitempty"`
	Address       *string `json:"address,omitempty"`
	Site          *string `json:"site,omitempty"`
}

// easyjson:json
type AgentCreateV1Res struct {
	Agent *AgentCreateAgent `json:"agent"`
}

type AgentCreateAgent struct {
	ID            uint64 `json:"id_agent"`
	IntegrationFK uint64 `json:"fk_integration"`
	Name          string `json:"name"`
	CompanyID     string `json:"company_id"`
	Description   string `json:"description"`
	Logo          string `json:"logo"`
	Background    string `json:"bg_image"`
	Phone         string `json:"phone"`
	Email         string `json:"email"`
	Address       string `json:"address"`
	Site          string `json:"site"`
	UpdatedAt     string `json:"updated_at"`
	CreatedAt     string `json:"created_at"`
}

type AgentCreateV1Errors int

const (
	AgentCreateV1Errors_CREATE_FAILED = iota
	AgentCreateV1Errors_AGENT_NOT_CREATED
)

var _AgentCreateV1ErrorsMapping = map[string]int{
	"CREATE_FAILED":     AgentCreateV1Errors_CREATE_FAILED,
	"AGENT_NOT_CREATED": AgentCreateV1Errors_AGENT_NOT_CREATED,
}

// easyjson:json
type AgentListV1Args struct {
	UserID uint64  `json:"user_id"`
	Limit  *uint64 `json:"limit,omitempty"`
	Offset *uint64 `json:"offset,omitempty"`
}

// easyjson:json
type AgentListV1Res struct {
	List []AgentListListItem `json:"list"`
}

type AgentListListItem struct {
	Agent    AgentListAgent    `json:"agent"`
	Employee AgentListEmployee `json:"employee"`
}

type AgentListAgent struct {
	ID            uint64 `json:"id_agent"`
	IntegrationFK uint64 `json:"fk_integration"`
	Name          string `json:"name"`
	CompanyID     string `json:"company_id"`
	Description   string `json:"description"`
	Logo          string `json:"logo"`
	Background    string `json:"bg_image"`
	Phone         string `json:"phone"`
	Email         string `json:"email"`
	Address       string `json:"address"`
	Site          string `json:"site"`
	UpdatedAt     string `json:"updated_at"`
	CreatedAt     string `json:"created_at"`
}

type AgentListEmployee struct {
	ID                 uint64  `json:"id_employee"`
	AgentFK            uint64  `json:"fk_agent"`
	UserFK             *uint64 `json:"fk_user"`
	Code               string  `json:"employee_code"`
	Name               string  `json:"name"`
	Role               string  `json:"role"`
	Email              string  `json:"email"`
	HireDate           string  `json:"hire_date"`
	NumberOfDepandants uint    `json:"number_of_dependants"`
	GrossBaseSalary    float64 `json:"gross_base_salary"`
	UpdatedAt          string  `json:"updated_at"`
	CreatedAt          string  `json:"created_at"`
}

// easyjson:json
type AgentUpdateV1Args struct {
	ID            uint64  `json:"id_agent"`
	IntegrationFK uint64  `json:"fk_integration"`
	Name          *string `json:"name,omitempty"`
	CompanyID     *string `json:"company_id,omitempty"`
	Description   *string `json:"description,omitempty"`
	Logo          *string `json:"logo,omitempty"`
	Background    *string `json:"bg_image,omitempty"`
	Phone         *string `json:"phone,omitempty"`
	Email         *string `json:"email,omitempty"`
	Address       *string `json:"address,omitempty"`
	Site          *string `json:"site,omitempty"`
}

// easyjson:json
type AgentUpdateV1Res struct {
	Agent *AgentUpdateAgent `json:"agent"`
}

type AgentUpdateAgent struct {
	ID            uint64 `json:"id_agent"`
	IntegrationFK uint64 `json:"fk_integration"`
	Name          string `json:"name"`
	CompanyID     string `json:"company_id"`
	Description   string `json:"description"`
	Logo          string `json:"logo"`
	Background    string `json:"bg_image"`
	Phone         string `json:"phone"`
	Email         string `json:"email"`
	Address       string `json:"address"`
	Site          string `json:"site"`
	UpdatedAt     string `json:"updated_at"`
	CreatedAt     string `json:"created_at"`
}

type AgentUpdateV1Errors int

const (
	AgentUpdateV1Errors_AGENT_NOT_FOUND = iota
	AgentUpdateV1Errors_UPDATE_FAILED
	AgentUpdateV1Errors_AGENT_NOT_UPDATED
)

var _AgentUpdateV1ErrorsMapping = map[string]int{
	"AGENT_NOT_FOUND":   AgentUpdateV1Errors_AGENT_NOT_FOUND,
	"UPDATE_FAILED":     AgentUpdateV1Errors_UPDATE_FAILED,
	"AGENT_NOT_UPDATED": AgentUpdateV1Errors_AGENT_NOT_UPDATED,
}

// easyjson:json
type EmployeeCreateV1Args struct {
	AgentFK            uint64   `json:"fk_agent"`
	UserFK             *uint64  `json:"fk_user,omitempty"`
	Code               *string  `json:"employee_code,omitempty"`
	Name               string   `json:"name"`
	Role               *string  `json:"role,omitempty"`
	Email              *string  `json:"email,omitempty"`
	HireDate           *string  `json:"hire_date,omitempty"`
	NumberOfDepandants *uint    `json:"number_of_dependants,omitempty"`
	GrossBaseSalary    *float64 `json:"gross_base_salary,omitempty"`
}

// easyjson:json
type EmployeeCreateV1Res struct {
	Agent    *EmployeeCreateAgent    `json:"agent"`
	Employee *EmployeeCreateEmployee `json:"employee"`
	User     *EmployeeCreateUser     `json:"user"`
}

type EmployeeCreateAgent struct {
	ID            uint64 `json:"id_agent"`
	IntegrationFK uint64 `json:"fk_integration"`
	Name          string `json:"name"`
	CompanyID     string `json:"company_id"`
	Description   string `json:"description"`
	Logo          string `json:"logo"`
	Background    string `json:"bg_image"`
	Phone         string `json:"phone"`
	Email         string `json:"email"`
	Address       string `json:"address"`
	Site          string `json:"site"`
	UpdatedAt     string `json:"updated_at"`
	CreatedAt     string `json:"created_at"`
}

type EmployeeCreateEmployee struct {
	ID                 uint64  `json:"id_employee"`
	AgentFK            uint64  `json:"fk_agent"`
	UserFK             *uint64 `json:"fk_user"`
	Code               string  `json:"employee_code"`
	Name               string  `json:"name"`
	Role               string  `json:"role"`
	Email              string  `json:"email"`
	HireDate           string  `json:"hire_date"`
	NumberOfDepandants uint    `json:"number_of_dependants"`
	GrossBaseSalary    float64 `json:"gross_base_salary"`
	UpdatedAt          string  `json:"updated_at"`
	CreatedAt          string  `json:"created_at"`
}

type EmployeeCreateUser struct {
	ID          uint64 `json:"id_user"`
	Name        string `json:"name"`
	Short       string `json:"p_description"`
	Description string `json:"description"`
	Avatar      string `json:"avatar"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	UpdatedAt   string `json:"updated_at"`
	CreatedAt   string `json:"created_at"`
}

type EmployeeCreateV1Errors int

const (
	EmployeeCreateV1Errors_AGENT_NOT_FOUND = iota
	EmployeeCreateV1Errors_USER_NOT_FOUND
	EmployeeCreateV1Errors_CREATE_FAILED
	EmployeeCreateV1Errors_EMPLOYEE_NOT_CREATED
	EmployeeCreateV1Errors_EMPLOYEE_ALREADY_EXISTS
)

var _EmployeeCreateV1ErrorsMapping = map[string]int{
	"AGENT_NOT_FOUND":         EmployeeCreateV1Errors_AGENT_NOT_FOUND,
	"USER_NOT_FOUND":          EmployeeCreateV1Errors_USER_NOT_FOUND,
	"CREATE_FAILED":           EmployeeCreateV1Errors_CREATE_FAILED,
	"EMPLOYEE_NOT_CREATED":    EmployeeCreateV1Errors_EMPLOYEE_NOT_CREATED,
	"EMPLOYEE_ALREADY_EXISTS": EmployeeCreateV1Errors_EMPLOYEE_ALREADY_EXISTS,
}

// easyjson:json
type EmployeeDetailsV1Args struct {
	ID            *uint64 `json:"id_employee,omitempty"`
	AgentFK       *uint64 `json:"fk_agent,omitempty"`
	UserFK        *uint64 `json:"fk_user,omitempty"`
	IntegrationFK *uint64 `json:"fk_integraion,omitempty"`
	CompanyID     *string `json:"company_id,omitempty"`
	Code          *string `json:"employee_code,omitempty"`
}

// easyjson:json
type EmployeeDetailsV1Res struct {
	Agent    *EmployeeDetailsAgent    `json:"agent"`
	Employee *EmployeeDetailsEmployee `json:"employee"`
	Payslips []EmployeeDetailsPayslip `json:"payslips"`
}

type EmployeeDetailsAgent struct {
	ID            uint64 `json:"id_agent"`
	IntegrationFK uint64 `json:"fk_integration"`
	Name          string `json:"name"`
	CompanyID     string `json:"company_id"`
	Description   string `json:"description"`
	Logo          string `json:"logo"`
	Background    string `json:"bg_image"`
	Phone         string `json:"phone"`
	Email         string `json:"email"`
	Address       string `json:"address"`
	Site          string `json:"site"`
	UpdatedAt     string `json:"updated_at"`
	CreatedAt     string `json:"created_at"`
}

type EmployeeDetailsEmployee struct {
	ID                 uint64  `json:"id_employee"`
	AgentFK            uint64  `json:"fk_agent"`
	UserFK             *uint64 `json:"fk_user"`
	Code               string  `json:"employee_code"`
	Name               string  `json:"name"`
	Role               string  `json:"role"`
	Email              string  `json:"email"`
	HireDate           string  `json:"hire_date"`
	NumberOfDepandants uint    `json:"number_of_dependants"`
	GrossBaseSalary    float64 `json:"gross_base_salary"`
	UpdatedAt          string  `json:"updated_at"`
	CreatedAt          string  `json:"created_at"`
}

type EmployeeDetailsPayslip struct {
	ID         uint64  `json:"id_payslip"`
	EmployeeFK uint64  `json:"fk_employee"`
	Title      string  `json:"title"`
	Currency   string  `json:"currency"`
	Amount     float64 `json:"amount"`
	UpdatedAt  string  `json:"updated_at"`
	CreatedAt  string  `json:"created_at"`
}

type EmployeeDetailsV1Errors int

const (
	EmployeeDetailsV1Errors_MISSED_REQUIRED_FIELDS = iota
	EmployeeDetailsV1Errors_AGENT_NOT_FOUND
	EmployeeDetailsV1Errors_EMPLOYEE_NOT_FOUND
)

var _EmployeeDetailsV1ErrorsMapping = map[string]int{
	"MISSED_REQUIRED_FIELDS": EmployeeDetailsV1Errors_MISSED_REQUIRED_FIELDS,
	"AGENT_NOT_FOUND":        EmployeeDetailsV1Errors_AGENT_NOT_FOUND,
	"EMPLOYEE_NOT_FOUND":     EmployeeDetailsV1Errors_EMPLOYEE_NOT_FOUND,
}

// easyjson:json
type EmployeeInviteV1Args struct {
	ID    uint64  `json:"id_employee"`
	Email *string `json:"email,omitempty"`
}

// easyjson:json
type EmployeeInviteV1Res struct {
	Result   string                  `json:"result"`
	Code     string                  `json:"magic_code"`
	Employee *EmployeeInviteEmployee `json:"employee"`
}

type EmployeeInviteEmployee struct {
	ID                 uint64  `json:"id_employee"`
	AgentFK            uint64  `json:"fk_agent"`
	UserFK             *uint64 `json:"fk_user"`
	Code               string  `json:"employee_code"`
	Name               string  `json:"name"`
	Role               string  `json:"role"`
	Email              string  `json:"email"`
	HireDate           string  `json:"hire_date"`
	NumberOfDepandants uint    `json:"number_of_dependants"`
	GrossBaseSalary    float64 `json:"gross_base_salary"`
	UpdatedAt          string  `json:"updated_at"`
	CreatedAt          string  `json:"created_at"`
}

type EmployeeInviteV1Errors int

const (
	EmployeeInviteV1Errors_MISSED_REQUIRED_FIELDS = iota
	EmployeeInviteV1Errors_AGENT_NOT_FOUND
	EmployeeInviteV1Errors_EMPLOYEE_NOT_FOUND
)

var _EmployeeInviteV1ErrorsMapping = map[string]int{
	"MISSED_REQUIRED_FIELDS": EmployeeInviteV1Errors_MISSED_REQUIRED_FIELDS,
	"AGENT_NOT_FOUND":        EmployeeInviteV1Errors_AGENT_NOT_FOUND,
	"EMPLOYEE_NOT_FOUND":     EmployeeInviteV1Errors_EMPLOYEE_NOT_FOUND,
}

// easyjson:json
type EmployeeListV1Args struct {
	AgentID uint64  `json:"agent_id"`
	Limit   *uint64 `json:"limit,omitempty"`
	Offset  *uint64 `json:"offset,omitempty"`
}

// easyjson:json
type EmployeeListV1Res struct {
	List []EmployeeListListItem `json:"list"`
}

type EmployeeListListItem struct {
	Employee EmployeeListEmployee `json:"employee"`
}

type EmployeeListEmployee struct {
	ID                 uint64  `json:"id_employee"`
	AgentFK            uint64  `json:"fk_agent"`
	UserFK             *uint64 `json:"fk_user"`
	Code               string  `json:"employee_code"`
	Name               string  `json:"name"`
	Role               string  `json:"role"`
	Email              string  `json:"email"`
	HireDate           string  `json:"hire_date"`
	NumberOfDepandants uint    `json:"number_of_dependants"`
	GrossBaseSalary    float64 `json:"gross_base_salary"`
	UpdatedAt          string  `json:"updated_at"`
	CreatedAt          string  `json:"created_at"`
}

// easyjson:json
type EmployeeUpdateV1Args struct {
	ID                 *uint64  `json:"id_employee,omitempty"`
	AgentFK            *uint64  `json:"fk_agent,omitempty"`
	UserFK             *uint64  `json:"fk_user,omitempty"`
	Code               *string  `json:"employee_code,omitempty"`
	Name               *string  `json:"name,omitempty"`
	Role               *string  `json:"role,omitempty"`
	Email              *string  `json:"email,omitempty"`
	HireDate           *string  `json:"hire_date,omitempty"`
	NumberOfDepandants *uint    `json:"number_of_dependants,omitempty"`
	GrossBaseSalary    *float64 `json:"gross_base_salary,omitempty"`
}

// easyjson:json
type EmployeeUpdateV1Res struct {
	Agent    *EmployeeUpdateAgent    `json:"agent"`
	Employee *EmployeeUpdateEmployee `json:"employee"`
	User     *EmployeeUpdateUser     `json:"user"`
}

type EmployeeUpdateAgent struct {
	ID            uint64 `json:"id_agent"`
	IntegrationFK uint64 `json:"fk_integration"`
	Name          string `json:"name"`
	CompanyID     string `json:"company_id"`
	Description   string `json:"description"`
	Logo          string `json:"logo"`
	Background    string `json:"bg_image"`
	Phone         string `json:"phone"`
	Email         string `json:"email"`
	Address       string `json:"address"`
	Site          string `json:"site"`
	UpdatedAt     string `json:"updated_at"`
	CreatedAt     string `json:"created_at"`
}

type EmployeeUpdateEmployee struct {
	ID                 uint64  `json:"id_employee"`
	AgentFK            uint64  `json:"fk_agent"`
	UserFK             *uint64 `json:"fk_user"`
	Code               string  `json:"employee_code"`
	Name               string  `json:"name"`
	Role               string  `json:"role"`
	Email              string  `json:"email"`
	HireDate           string  `json:"hire_date"`
	NumberOfDepandants uint    `json:"number_of_dependants"`
	GrossBaseSalary    float64 `json:"gross_base_salary"`
	UpdatedAt          string  `json:"updated_at"`
	CreatedAt          string  `json:"created_at"`
}

type EmployeeUpdateUser struct {
	ID          uint64 `json:"id_user"`
	Name        string `json:"name"`
	Short       string `json:"p_description"`
	Description string `json:"description"`
	Avatar      string `json:"avatar"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	UpdatedAt   string `json:"updated_at"`
	CreatedAt   string `json:"created_at"`
}

type EmployeeUpdateV1Errors int

const (
	EmployeeUpdateV1Errors_NOT_ENOUGH_PARAMS = iota
	EmployeeUpdateV1Errors_AGENT_NOT_FOUND
	EmployeeUpdateV1Errors_EMPLOYEE_NOT_FOUND
	EmployeeUpdateV1Errors_USER_NOT_FOUND
	EmployeeUpdateV1Errors_UPDATE_FAILED
	EmployeeUpdateV1Errors_EMPLOYEE_NOT_UPDATED
	EmployeeUpdateV1Errors_EMPLOYEE_ALREADY_EXISTS
)

var _EmployeeUpdateV1ErrorsMapping = map[string]int{
	"NOT_ENOUGH_PARAMS":       EmployeeUpdateV1Errors_NOT_ENOUGH_PARAMS,
	"AGENT_NOT_FOUND":         EmployeeUpdateV1Errors_AGENT_NOT_FOUND,
	"EMPLOYEE_NOT_FOUND":      EmployeeUpdateV1Errors_EMPLOYEE_NOT_FOUND,
	"USER_NOT_FOUND":          EmployeeUpdateV1Errors_USER_NOT_FOUND,
	"UPDATE_FAILED":           EmployeeUpdateV1Errors_UPDATE_FAILED,
	"EMPLOYEE_NOT_UPDATED":    EmployeeUpdateV1Errors_EMPLOYEE_NOT_UPDATED,
	"EMPLOYEE_ALREADY_EXISTS": EmployeeUpdateV1Errors_EMPLOYEE_ALREADY_EXISTS,
}

// easyjson:json
type IntegrationCheckV1Args struct {
	Hash string `json:"hash"`
}

// easyjson:json
type IntegrationCheckV1Res struct {
	Integration *IntegrationCheckIntegration `json:"integration"`
}

type IntegrationCheckIntegration struct {
	ID        uint64 `json:"id_integration"`
	Hash      string `json:"hash"`
	UpdatedAt string `json:"updated_at"`
	CreatedAt string `json:"created_at"`
}

type IntegrationCheckV1Errors int

const (
	IntegrationCheckV1Errors_INTEGRATION_NOT_FOUND = iota
)

var _IntegrationCheckV1ErrorsMapping = map[string]int{
	"INTEGRATION_NOT_FOUND": IntegrationCheckV1Errors_INTEGRATION_NOT_FOUND,
}

// easyjson:json
type IntegrationCreateV1Args struct {
	Hash string `json:"hash"`
}

// easyjson:json
type IntegrationCreateV1Res struct {
	Integration *IntegrationCreateIntegration `json:"integration"`
}

type IntegrationCreateIntegration struct {
	ID        uint64 `json:"id_integration"`
	Hash      string `json:"hash"`
	UpdatedAt string `json:"updated_at"`
	CreatedAt string `json:"created_at"`
}

type IntegrationCreateV1Errors int

const (
	IntegrationCreateV1Errors_CREATE_FAILED = iota
	IntegrationCreateV1Errors_DUBLICATE_HASH
	IntegrationCreateV1Errors_INTEGRATION_NOT_CREATED
)

var _IntegrationCreateV1ErrorsMapping = map[string]int{
	"CREATE_FAILED":           IntegrationCreateV1Errors_CREATE_FAILED,
	"DUBLICATE_HASH":          IntegrationCreateV1Errors_DUBLICATE_HASH,
	"INTEGRATION_NOT_CREATED": IntegrationCreateV1Errors_INTEGRATION_NOT_CREATED,
}

// easyjson:json
type PayslipCreateV1Args struct {
	Payslip PayslipCreatePayslipArgs `json:"payslip"`
}

type PayslipCreatePayslipArgs struct {
	EmployeeFK uint64  `json:"fk_employee"`
	Title      string  `json:"title"`
	Currency   string  `json:"currency"`
	Amount     float64 `json:"amount"`
	Data       string  `json:"data"`
}

// easyjson:json
type PayslipCreateV1Res struct {
	Employee *PayslipCreateEmployee `json:"agent"`
	Payslip  *PayslipCreatePayslip  `json:"payslip"`
}

type PayslipCreateEmployee struct {
	ID                 uint64  `json:"id_employee"`
	AgentFK            uint64  `json:"fk_agent"`
	UserFK             *uint64 `json:"fk_user"`
	Code               string  `json:"employee_code"`
	Name               string  `json:"name"`
	Role               string  `json:"role"`
	Email              string  `json:"email"`
	HireDate           string  `json:"hire_date"`
	NumberOfDepandants uint    `json:"number_of_dependants"`
	GrossBaseSalary    float64 `json:"gross_base_salary"`
	UpdatedAt          string  `json:"updated_at"`
	CreatedAt          string  `json:"created_at"`
}

type PayslipCreatePayslip struct {
	ID         uint64  `json:"id_payslip"`
	EmployeeFK uint64  `json:"fk_employee"`
	Title      string  `json:"title"`
	Currency   string  `json:"currency"`
	Amount     float64 `json:"amount"`
	UpdatedAt  string  `json:"updated_at"`
	CreatedAt  string  `json:"created_at"`
}

type PayslipCreateV1Errors int

const (
	PayslipCreateV1Errors_EMPLOYEE_NOT_FOUND = iota
	PayslipCreateV1Errors_CREATE_FAILED
	PayslipCreateV1Errors_PAYSLIP_NOT_CREATED
)

var _PayslipCreateV1ErrorsMapping = map[string]int{
	"EMPLOYEE_NOT_FOUND":  PayslipCreateV1Errors_EMPLOYEE_NOT_FOUND,
	"CREATE_FAILED":       PayslipCreateV1Errors_CREATE_FAILED,
	"PAYSLIP_NOT_CREATED": PayslipCreateV1Errors_PAYSLIP_NOT_CREATED,
}

// easyjson:json
type PayslipDetailsV1Args struct {
	ID uint64 `json:"payslip_id"`
}

// easyjson:json
type PayslipDetailsV1Res struct {
	Agent    *PayslipDetailsAgent    `json:"agent"`
	Employee *PayslipDetailsEmployee `json:"employee"`
	Payslip  PayslipDetailsPayslip   `json:"payslip"`
}

type PayslipDetailsAgent struct {
	ID            uint64 `json:"id_agent"`
	IntegrationFK uint64 `json:"fk_integration"`
	Name          string `json:"name"`
	CompanyID     string `json:"company_id"`
	Description   string `json:"description"`
	Logo          string `json:"logo"`
	Phone         string `json:"phone"`
	Email         string `json:"email"`
	Address       string `json:"address"`
	Site          string `json:"site"`
	UpdatedAt     string `json:"updated_at"`
	CreatedAt     string `json:"created_at"`
}

type PayslipDetailsEmployee struct {
	ID                 uint64  `json:"id_employee"`
	AgentFK            uint64  `json:"fk_agent"`
	UserFK             *uint64 `json:"fk_user"`
	Code               string  `json:"employee_code"`
	Name               string  `json:"name"`
	Role               string  `json:"role"`
	Email              string  `json:"email"`
	HireDate           string  `json:"hire_date"`
	NumberOfDepandants uint    `json:"number_of_dependants"`
	GrossBaseSalary    float64 `json:"gross_base_salary"`
	UpdatedAt          string  `json:"updated_at"`
	CreatedAt          string  `json:"created_at"`
}

type PayslipDetailsPayslip struct {
	ID         uint64  `json:"id_payslip"`
	EmployeeFK uint64  `json:"fk_employee"`
	Title      string  `json:"title"`
	Currency   string  `json:"currency"`
	Amount     float64 `json:"amount"`
	UpdatedAt  string  `json:"updated_at"`
	CreatedAt  string  `json:"created_at"`
	Data       string  `json:"data"`
}

type PayslipDetailsV1Errors int

const (
	PayslipDetailsV1Errors_AGENT_NOT_FOUND = iota
	PayslipDetailsV1Errors_EMPLOYEE_NOT_FOUND
	PayslipDetailsV1Errors_PAYSLIP_NOT_FOUND
)

var _PayslipDetailsV1ErrorsMapping = map[string]int{
	"AGENT_NOT_FOUND":    PayslipDetailsV1Errors_AGENT_NOT_FOUND,
	"EMPLOYEE_NOT_FOUND": PayslipDetailsV1Errors_EMPLOYEE_NOT_FOUND,
	"PAYSLIP_NOT_FOUND":  PayslipDetailsV1Errors_PAYSLIP_NOT_FOUND,
}

// easyjson:json
type PayslipListV1Args struct {
	UserID uint64  `json:"user_id"`
	Limit  *uint64 `json:"limit,omitempty"`
	Offset *uint64 `json:"offset,omitempty"`
}

// easyjson:json
type PayslipListV1Res struct {
	List []PayslipListListItem `json:"list"`
}

type PayslipListListItem struct {
	Agent    PayslipListAgent    `json:"agent"`
	Employee PayslipListEmployee `json:"employee"`
	Payslip  PayslipListPayslip  `json:"payslip"`
}

type PayslipListAgent struct {
	ID            uint64 `json:"id_agent"`
	IntegrationFK uint64 `json:"fk_integration"`
	Name          string `json:"name"`
	CompanyID     string `json:"company_id"`
	Description   string `json:"description"`
	Logo          string `json:"logo"`
	Background    string `json:"bg_image"`
	Phone         string `json:"phone"`
	Email         string `json:"email"`
	Address       string `json:"address"`
	Site          string `json:"site"`
	UpdatedAt     string `json:"updated_at"`
	CreatedAt     string `json:"created_at"`
}

type PayslipListEmployee struct {
	ID                 uint64  `json:"id_employee"`
	AgentFK            uint64  `json:"fk_agent"`
	UserFK             *uint64 `json:"fk_user"`
	Code               string  `json:"employee_code"`
	Name               string  `json:"name"`
	Role               string  `json:"role"`
	Email              string  `json:"email"`
	HireDate           string  `json:"hire_date"`
	NumberOfDepandants uint    `json:"number_of_dependants"`
	GrossBaseSalary    float64 `json:"gross_base_salary"`
	UpdatedAt          string  `json:"updated_at"`
	CreatedAt          string  `json:"created_at"`
}

type PayslipListPayslip struct {
	ID         uint64  `json:"id_payslip"`
	EmployeeFK uint64  `json:"fk_employee"`
	Title      string  `json:"title"`
	Currency   string  `json:"currency"`
	Amount     float64 `json:"amount"`
	UpdatedAt  string  `json:"updated_at"`
	CreatedAt  string  `json:"created_at"`
}

// easyjson:json
type PayslipListByEmployeeV1Args struct {
	EmployeeID uint64  `json:"employee_id"`
	Limit      *uint64 `json:"limit,omitempty"`
	Offset     *uint64 `json:"offset,omitempty"`
}

// easyjson:json
type PayslipListByEmployeeV1Res struct {
	List []PayslipListByEmployeeListItem `json:"list"`
}

type PayslipListByEmployeeListItem struct {
	Payslip PayslipListByEmployeePayslip `json:"payslip"`
}

type PayslipListByEmployeePayslip struct {
	ID         uint64  `json:"id_payslip"`
	EmployeeFK uint64  `json:"fk_employee"`
	Title      string  `json:"title"`
	Currency   string  `json:"currency"`
	Amount     float64 `json:"amount"`
	UpdatedAt  string  `json:"updated_at"`
	CreatedAt  string  `json:"created_at"`
}

// easyjson:json
type SettingCreateV1Args struct {
	AgentFK               uint64  `json:"fk_agent"`
	UserFK                *uint64 `json:"fk_user,omitempty"`
	AgentProcessedFK      *uint64 `json:"fk_agent_processed,omitempty"`
	Role                  string  `json:"role"`
	IsNotificationEnabled bool    `json:"notifications_enabled"`
	IsMainAgent           bool    `json:"is_main_agent"`
}

// easyjson:json
type SettingCreateV1Res struct {
	Agent          *SettingCreateAgent        `json:"agent"`
	AgentProcessed *SettingCreateAgent        `json:"agent_processed"`
	Setting        *SettingCreateAgentSetting `json:"setting"`
	User           *SettingCreateUser         `json:"user"`
}

type SettingCreateAgent struct {
	ID            uint64 `json:"id_agent"`
	IntegrationFK uint64 `json:"fk_integration"`
	Name          string `json:"name"`
	CompanyID     string `json:"company_id"`
	Description   string `json:"description"`
	Logo          string `json:"logo"`
	Background    string `json:"bg_image"`
	Phone         string `json:"phone"`
	Email         string `json:"email"`
	Address       string `json:"address"`
	Site          string `json:"site"`
	UpdatedAt     string `json:"updated_at"`
	CreatedAt     string `json:"created_at"`
}

type SettingCreateAgentSetting struct {
	ID                    uint64  `json:"id_setting"`
	AgentFK               uint64  `json:"fk_agent"`
	AgentProcessedFK      *uint64 `json:"fk_agent_processed"`
	UserFK                *uint64 `json:"fk_user"`
	Role                  string  `json:"role"`
	IsNotificationEnabled bool    `json:"notifications_enabled"`
	IsMainAgent           bool    `json:"is_main_agent"`
	UpdatedAt             string  `json:"updated_at"`
	CreatedAt             string  `json:"created_at"`
}

type SettingCreateUser struct {
	ID          uint64 `json:"id_user"`
	Name        string `json:"name"`
	Short       string `json:"p_description"`
	Description string `json:"description"`
	Avatar      string `json:"avatar"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	UpdatedAt   string `json:"updated_at"`
	CreatedAt   string `json:"created_at"`
}

type SettingCreateV1Errors int

const (
	SettingCreateV1Errors_AGENT_NOT_FOUND = iota
	SettingCreateV1Errors_USER_NOT_FOUND
	SettingCreateV1Errors_CREATE_FAILED
	SettingCreateV1Errors_SETTING_NOT_CREATED
	SettingCreateV1Errors_SETTING_ALREADY_EXISTS
)

var _SettingCreateV1ErrorsMapping = map[string]int{
	"AGENT_NOT_FOUND":        SettingCreateV1Errors_AGENT_NOT_FOUND,
	"USER_NOT_FOUND":         SettingCreateV1Errors_USER_NOT_FOUND,
	"CREATE_FAILED":          SettingCreateV1Errors_CREATE_FAILED,
	"SETTING_NOT_CREATED":    SettingCreateV1Errors_SETTING_NOT_CREATED,
	"SETTING_ALREADY_EXISTS": SettingCreateV1Errors_SETTING_ALREADY_EXISTS,
}

// easyjson:json
type SettingListV1Args struct {
	IntegrationID uint64 `json:"integration_id"`
	UserID        uint64 `json:"user_id"`
}

// easyjson:json
type SettingListV1Res struct {
	List []SettingListListItem `json:"list"`
}

type SettingListListItem struct {
	Agent   *SettingListAgent        `json:"agent"`
	Setting *SettingListAgentSetting `json:"setting"`
}

type SettingListAgent struct {
	ID            uint64 `json:"id_agent"`
	IntegrationFK uint64 `json:"fk_integration"`
	Name          string `json:"name"`
	CompanyID     string `json:"company_id"`
	Description   string `json:"description"`
	Logo          string `json:"logo"`
	Background    string `json:"bg_image"`
	Phone         string `json:"phone"`
	Email         string `json:"email"`
	Address       string `json:"address"`
	Site          string `json:"site"`
	UpdatedAt     string `json:"updated_at"`
	CreatedAt     string `json:"created_at"`
}

type SettingListAgentSetting struct {
	ID                    uint64  `json:"id_setting"`
	AgentFK               uint64  `json:"fk_agent"`
	AgentProcessedFK      *uint64 `json:"fk_agent_processed"`
	UserFK                *uint64 `json:"fk_user"`
	Role                  string  `json:"role"`
	IsNotificationEnabled bool    `json:"notifications_enabled"`
	IsMainAgent           bool    `json:"is_main_agent"`
	UpdatedAt             string  `json:"updated_at"`
	CreatedAt             string  `json:"created_at"`
}

// easyjson:json
type SettingUpdateV1Args struct {
	ID                    *uint64 `json:"id_setting,omitempty"`
	AgentFK               *uint64 `json:"fk_agent,omitempty"`
	UserFK                *uint64 `json:"fk_user,omitempty"`
	AgentProcessedFK      *uint64 `json:"fk_agent_processed,omitempty"`
	Role                  *string `json:"role,omitempty"`
	IsNotificationEnabled *bool   `json:"notifications_enabled,omitempty"`
	IsMainAgent           *bool   `json:"is_main_agent,omitempty"`
}

// easyjson:json
type SettingUpdateV1Res struct {
	Agent          *SettingUpdateAgent        `json:"agent"`
	AgentProcessed *SettingUpdateAgent        `json:"agent_processed"`
	Setting        *SettingUpdateAgentSetting `json:"setting"`
	User           *SettingUpdateUser         `json:"user"`
}

type SettingUpdateAgent struct {
	ID            uint64 `json:"id_agent"`
	IntegrationFK uint64 `json:"fk_integration"`
	Name          string `json:"name"`
	CompanyID     string `json:"company_id"`
	Description   string `json:"description"`
	Logo          string `json:"logo"`
	Background    string `json:"bg_image"`
	Phone         string `json:"phone"`
	Email         string `json:"email"`
	Address       string `json:"address"`
	Site          string `json:"site"`
	UpdatedAt     string `json:"updated_at"`
	CreatedAt     string `json:"created_at"`
}

type SettingUpdateAgentSetting struct {
	ID                    uint64  `json:"id_setting"`
	AgentFK               uint64  `json:"fk_agent"`
	AgentProcessedFK      *uint64 `json:"fk_agent_processed"`
	UserFK                *uint64 `json:"fk_user"`
	Role                  string  `json:"role"`
	IsNotificationEnabled bool    `json:"notifications_enabled"`
	IsMainAgent           bool    `json:"is_main_agent"`
	UpdatedAt             string  `json:"updated_at"`
	CreatedAt             string  `json:"created_at"`
}

type SettingUpdateUser struct {
	ID          uint64 `json:"id_user"`
	Name        string `json:"name"`
	Short       string `json:"p_description"`
	Description string `json:"description"`
	Avatar      string `json:"avatar"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	UpdatedAt   string `json:"updated_at"`
	CreatedAt   string `json:"created_at"`
}

type SettingUpdateV1Errors int

const (
	SettingUpdateV1Errors_NOT_ENOUGH_PARAMS = iota
	SettingUpdateV1Errors_AGENT_NOT_FOUND
	SettingUpdateV1Errors_SETTING_NOT_FOUND
	SettingUpdateV1Errors_USER_NOT_FOUND
	SettingUpdateV1Errors_UPDATE_FAILED
	SettingUpdateV1Errors_SETTING_NOT_UPDATED
	SettingUpdateV1Errors_SETTING_ALREADY_EXISTS
)

var _SettingUpdateV1ErrorsMapping = map[string]int{
	"NOT_ENOUGH_PARAMS":      SettingUpdateV1Errors_NOT_ENOUGH_PARAMS,
	"AGENT_NOT_FOUND":        SettingUpdateV1Errors_AGENT_NOT_FOUND,
	"SETTING_NOT_FOUND":      SettingUpdateV1Errors_SETTING_NOT_FOUND,
	"USER_NOT_FOUND":         SettingUpdateV1Errors_USER_NOT_FOUND,
	"UPDATE_FAILED":          SettingUpdateV1Errors_UPDATE_FAILED,
	"SETTING_NOT_UPDATED":    SettingUpdateV1Errors_SETTING_NOT_UPDATED,
	"SETTING_ALREADY_EXISTS": SettingUpdateV1Errors_SETTING_ALREADY_EXISTS,
}

// easyjson:json
type UserCreateV1Args struct {
	IntegrationFK *uint64 `json:"fk_integration,omitempty"`
	Name          *string `json:"name,omitempty"`
	Short         *string `json:"p_description,omitempty"`
	Description   *string `json:"description,omitempty"`
	Avatar        *string `json:"avatar,omitempty"`
	Phone         *string `json:"phone,omitempty"`
	Email         *string `json:"email,omitempty"`
	Password      string  `json:"password"`
}

// easyjson:json
type UserCreateV1Res struct {
	User *UserCreateUser `json:"user"`
}

type UserCreateUser struct {
	ID            uint64  `json:"id_user"`
	IntegrationFK *uint64 `json:"fk_integration,omitempty"`
	Name          string  `json:"name"`
	Short         string  `json:"p_description"`
	Description   string  `json:"description"`
	Avatar        string  `json:"avatar"`
	Phone         string  `json:"phone"`
	Email         string  `json:"email"`
	UpdatedAt     string  `json:"updated_at"`
	CreatedAt     string  `json:"created_at"`
}

type UserCreateV1Errors int

const (
	UserCreateV1Errors_MISSED_REQUIRED_FIELDS = iota
	UserCreateV1Errors_USER_EXISTS
	UserCreateV1Errors_CREATE_FAILED
	UserCreateV1Errors_USER_NOT_CREATED
)

var _UserCreateV1ErrorsMapping = map[string]int{
	"MISSED_REQUIRED_FIELDS": UserCreateV1Errors_MISSED_REQUIRED_FIELDS,
	"USER_EXISTS":            UserCreateV1Errors_USER_EXISTS,
	"CREATE_FAILED":          UserCreateV1Errors_CREATE_FAILED,
	"USER_NOT_CREATED":       UserCreateV1Errors_USER_NOT_CREATED,
}

// easyjson:json
type UserLoginV1Args struct {
	IntegrationFK *uint64 `json:"fk_integration,omitempty"`
	Login         string  `json:"login"`
	Password      string  `json:"password"`
}

// easyjson:json
type UserLoginV1Res struct {
	User *UserLoginUser `json:"user"`
}

type UserLoginUser struct {
	ID            uint64  `json:"id_user"`
	IntegrationFK *uint64 `json:"fk_integration,omitempty"`
	Name          string  `json:"name"`
	Short         string  `json:"p_description"`
	Description   string  `json:"description"`
	Avatar        string  `json:"avatar"`
	Phone         string  `json:"phone"`
	Email         string  `json:"email"`
	UpdatedAt     string  `json:"updated_at"`
	CreatedAt     string  `json:"created_at"`
}

type UserLoginV1Errors int

const (
	UserLoginV1Errors_LOGIN_FAILED = iota
	UserLoginV1Errors_USER_NOT_FOUND
)

var _UserLoginV1ErrorsMapping = map[string]int{
	"LOGIN_FAILED":   UserLoginV1Errors_LOGIN_FAILED,
	"USER_NOT_FOUND": UserLoginV1Errors_USER_NOT_FOUND,
}

// easyjson:json
type UserRemindV1Args struct {
	IntegrationFK *uint64 `json:"fk_integration,omitempty"`
	Login         string  `json:"login"`
}

// easyjson:json
type UserRemindV1Res struct {
	Result string          `json:"result"`
	Code   string          `json:"magic_code"`
	User   *UserRemindUser `json:"user"`
}

type UserRemindUser struct {
	ID            uint64  `json:"id_user"`
	IntegrationFK *uint64 `json:"fk_integration,omitempty"`
	Name          string  `json:"name"`
	Short         string  `json:"p_description"`
	Description   string  `json:"description"`
	Avatar        string  `json:"avatar"`
	Phone         string  `json:"phone"`
	Email         string  `json:"email"`
	UpdatedAt     string  `json:"updated_at"`
	CreatedAt     string  `json:"created_at"`
}

type UserRemindV1Errors int

const (
	UserRemindV1Errors_EMAIL_NOT_SENDED = iota
	UserRemindV1Errors_USER_NOT_FOUND
)

var _UserRemindV1ErrorsMapping = map[string]int{
	"EMAIL_NOT_SENDED": UserRemindV1Errors_EMAIL_NOT_SENDED,
	"USER_NOT_FOUND":   UserRemindV1Errors_USER_NOT_FOUND,
}

// easyjson:json
type UserUpdateV1Args struct {
	ID            uint64  `json:"id_user"`
	IntegrationFK *uint64 `json:"fk_integration,omitempty"`
	Name          *string `json:"name,omitempty"`
	Short         *string `json:"p_description,omitempty"`
	Description   *string `json:"description,omitempty"`
	Avatar        *string `json:"avatar,omitempty"`
	Phone         *string `json:"phone,omitempty"`
	Email         *string `json:"email,omitempty"`
	Password      *string `json:"password,omitempty"`
}

// easyjson:json
type UserUpdateV1Res struct {
	User *UserUpdateUser `json:"user"`
}

type UserUpdateUser struct {
	ID            uint64  `json:"id_user"`
	IntegrationFK *uint64 `json:"fk_integration,omitempty"`
	Name          string  `json:"name"`
	Short         string  `json:"p_description"`
	Description   string  `json:"description"`
	Avatar        string  `json:"avatar"`
	Phone         string  `json:"phone"`
	Email         string  `json:"email"`
	UpdatedAt     string  `json:"updated_at"`
	CreatedAt     string  `json:"created_at"`
}

type UserUpdateV1Errors int

const (
	UserUpdateV1Errors_USER_NOT_FOUND = iota
	UserUpdateV1Errors_UPDATE_FAILED
	UserUpdateV1Errors_NEW_EMAIL_IS_BUSY
	UserUpdateV1Errors_NEW_PHONE_IS_BUSY
)

var _UserUpdateV1ErrorsMapping = map[string]int{
	"USER_NOT_FOUND":    UserUpdateV1Errors_USER_NOT_FOUND,
	"UPDATE_FAILED":     UserUpdateV1Errors_UPDATE_FAILED,
	"NEW_EMAIL_IS_BUSY": UserUpdateV1Errors_NEW_EMAIL_IS_BUSY,
	"NEW_PHONE_IS_BUSY": UserUpdateV1Errors_NEW_PHONE_IS_BUSY,
}

// TODO: duplicates http_json.httpSessionResponse
// easyjson:json
type httpSessionResponse struct {
	Result string              `json:"result"` //OK or ERROR
	Data   easyjson.RawMessage `json:"data"`
	Error  string              `json:"error"`
}

func unmarshal(data []byte, r interface{}) error {
	if m, ok := r.(easyjson.Unmarshaler); ok {
		return easyjson.Unmarshal(data, m)
	}
	return json.Unmarshal(data, r)
}

func (api *MotifyCoreAPIGoRPC) set(ctx context.Context, path string, data interface{}, buf interface{}, handlerErrors map[string]int) (err error) {
	startTime := time.Now()

	var apiURL string
	var req *http.Request

	if api.callbacks.OnStart != nil {
		ctx = api.callbacks.OnStart(ctx, req)
	}

	defer func() {
		if api.callbacks.OnFinish != nil {
			api.callbacks.OnFinish(ctx, req, startTime)
		}

		if r := recover(); r != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			n := runtime.Stack(buf, false)
			trace := buf[:n]

			err = fmt.Errorf("panic while calling %q service: %v", api.serviceName, r)
			if api.callbacks.OnPanic != nil {
				err = api.callbacks.OnPanic(ctx, req, r, trace)
			}
		}
	}()

	apiURL, err = api.balancer.Next()
	if err != nil {
		err = fmt.Errorf("could not locate service '%s': %v", api.serviceName, err)
		if api.callbacks.OnError != nil {
			err = api.callbacks.OnError(ctx, req, err)
		}
		return err
	}

	b := bytes.NewBuffer(nil)
	if m, ok := data.(easyjson.Marshaler); ok {
		_, err = easyjson.MarshalToWriter(m, b)
	} else {
		encoder := json.NewEncoder(b)
		err = encoder.Encode(data)
	}
	if err != nil {
		err = fmt.Errorf("could not marshal data %+v: %v", data, err)
		if api.callbacks.OnError != nil {
			err = api.callbacks.OnError(ctx, req, err)
		}
		return err
	}

	req, err = http.NewRequest("POST", createRawURL(apiURL, path, nil), b)
	if err != nil {
		if api.callbacks.OnError != nil {
			err = api.callbacks.OnError(ctx, req, err)
		}
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if api.callbacks.OnPrepareRequest != nil {
		ctx = api.callbacks.OnPrepareRequest(ctx, req, data)
	}

	if err := api.doRequest(ctx, req, buf, handlerErrors); err != nil {
		if api.callbacks.OnError != nil {
			err = api.callbacks.OnError(ctx, req, err)
		}
		return err
	}

	if api.callbacks.OnSuccess != nil {
		api.callbacks.OnSuccess(ctx, req, buf)
	}

	return nil
}

func (api *MotifyCoreAPIGoRPC) setWithCache(ctx context.Context, path string, data interface{}, entry *cache.CacheEntry, handlerErrors map[string]int) error {
	if api.cache != nil && cache.IsTransportCacheEnabled(ctx) {
		cacheKey := getCacheKey(path, data)
		if cacheKey != nil {
			api.cache.Lock(cacheKey)
			defer api.cache.Unlock(cacheKey)
			cacheEntry := api.cache.Get(cacheKey)
			if cacheEntry != nil && cacheEntry.Body != nil {
				*entry = *cacheEntry
				return nil
			}
			if err := api.set(ctx, path, data, entry.Body, handlerErrors); err != nil {
				return err
			}
			ttl := cache.TTL(ctx)
			if p, ok := api.cache.(cache.TTLAwareCachePutter); ok && ttl > 0 {
				p.PutWithTTL(cacheKey, entry, ttl)
			} else {
				api.cache.Put(cacheKey, entry)
			}
			return nil
		}
	}
	return api.set(ctx, path, data, entry.Body, handlerErrors)
}

func createRawURL(url, path string, values url.Values) string {
	var buf bytes.Buffer
	buf.WriteString(strings.TrimRight(url, "/"))
	//buf.WriteRune('/')
	//buf.WriteString(strings.TrimLeft(path, "/"))
	// path must contain leading /
	buf.WriteString(path)
	if len(values) > 0 {
		buf.WriteRune('?')
		buf.WriteString(values.Encode())
	}
	return buf.String()
}

func (api *MotifyCoreAPIGoRPC) doRequest(ctx context.Context, request *http.Request, buf interface{}, handlerErrors map[string]int) error {
	return HTTPDo(ctx, api.client, request, func(response *http.Response, err error) error {
		// Run
		if err != nil {
			return err
		}
		defer response.Body.Close()

		// Handle error
		if response.StatusCode != http.StatusOK {
			switch response.StatusCode {
			// TODO separate error types for different status codes (and different callbacks)
			/*
			   case http.StatusForbidden:
			   case http.StatusBadGateway:
			   case http.StatusBadRequest:
			*/
			default:
				return fmt.Errorf("Request %q failed. Server returns status code %d", request.URL.RequestURI(), response.StatusCode)
			}
		}

		// Read response
		result, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return err
		}

		if api.callbacks.OnResponseUnmarshaling != nil {
			api.callbacks.OnResponseUnmarshaling(ctx, request, response, result)
		}

		var mainResp httpSessionResponse
		if err := unmarshal(result, &mainResp); err != nil {
			return fmt.Errorf("request %q failed to decode response %q: %v", request.URL.RequestURI(), string(result), err)
		}
		if mainResp.Result == "OK" {
			if err := unmarshal(mainResp.Data, buf); err != nil {
				return fmt.Errorf("request %q failed to decode response data %+v: %v", request.URL.RequestURI(), mainResp.Data, err)
			}
			return nil
		}

		if mainResp.Result == "ERROR" {
			errCode, ok := handlerErrors[mainResp.Error]
			if ok {
				return &ServiceError{
					Code:    errCode,
					Message: mainResp.Error,
				}
			}
		}

		return fmt.Errorf("request %q returned incorrect response %q", request.URL.RequestURI(), string(result))
	})
}

// HTTPDo is taken and adapted from https://blog.golang.org/context
func HTTPDo(ctx context.Context, client *http.Client, req *http.Request, f func(*http.Response, error) error) error {
	c := make(chan error, 1)
	go func() { c <- f(client.Do(req)) }()
	select {
	case <-ctx.Done():
		if tr, ok := client.Transport.(canceler); ok {
			tr.CancelRequest(req)
			<-c // Wait for f to return.
		}
		return ctx.Err()
	case err := <-c:
		return err
	}
}

type canceler interface {
	CancelRequest(*http.Request)
}

// ServiceError uses to separate critical and non-critical errors which returns in external service response.
// For this type of error we shouldn't use 500 error counter for librato
type ServiceError struct {
	Code    int
	Message string
}

// Error method for implementing common error interface
func (err *ServiceError) Error() string {
	return err.Message
}

func getCacheKey(route string, params interface{}) []byte {
	buf := bytes.NewBufferString(route)
	var err error
	if m, ok := params.(easyjson.Marshaler); ok {
		_, err = easyjson.MarshalToWriter(m, buf)
	} else {
		encoder := json.NewEncoder(buf)
		err = encoder.Encode(params)
	}
	if err != nil {
		return nil
	}
	return buf.Bytes()
}
