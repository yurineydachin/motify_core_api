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

func (api *MotifyCoreAPIGoRPC) EmployeeUpdateV1(ctx context.Context, options EmployeeUpdateV1Args) (*EmployeeUpdateV1Res, error) {
	var result *EmployeeUpdateV1Res
	var entry = cache.CacheEntry{Body: &result}
	err := api.setWithCache(ctx, "/employee/update/v1/", options, &entry, _EmployeeUpdateV1ErrorsMapping)
	if result, ok := entry.Body.(**EmployeeUpdateV1Res); ok {
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

func (api *MotifyCoreAPIGoRPC) PayslipSetV1(ctx context.Context, options PayslipSetV1Args) (string, error) {
	var result string
	var entry = cache.CacheEntry{Body: &result}
	err := api.setWithCache(ctx, "/payslip/set/v1/", options, &entry, nil)
	if result, ok := entry.Body.(*string); ok {
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
	Name        *string `json:"name,omitempty"`
	CompanyID   string  `json:"company_id"`
	Description *string `json:"description,omitempty"`
	Logo        *string `json:"logo,omitempty"`
	Background  *string `json:"bg_image,omitempty"`
	Phone       *string `json:"phone,omitempty"`
	Email       *string `json:"email,omitempty"`
	Address     *string `json:"address,omitempty"`
	Site        *string `json:"site,omitempty"`
}

// easyjson:json
type AgentCreateV1Res struct {
	Agent *AgentCreateAgent `json:"agent"`
}

type AgentCreateAgent struct {
	ID          uint64 `json:"id_agent"`
	Name        string `json:"name"`
	CompanyID   string `json:"company_id"`
	Description string `json:"description"`
	Logo        string `json:"Logo"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	Address     string `json:"address"`
	Site        string `json:"site"`
	UpdatedAt   string `json:"updated_at"`
	CreatedAt   string `json:"created_at"`
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
type AgentUpdateV1Args struct {
	ID          uint64  `json:"id_agent"`
	Name        *string `json:"name,omitempty"`
	CompanyID   *string `json:"company_id,omitempty"`
	Description *string `json:"description,omitempty"`
	Logo        *string `json:"logo,omitempty"`
	Background  *string `json:"bg_image,omitempty"`
	Phone       *string `json:"phone,omitempty"`
	Email       *string `json:"email,omitempty"`
	Address     *string `json:"address,omitempty"`
	Site        *string `json:"site,omitempty"`
}

// easyjson:json
type AgentUpdateV1Res struct {
	Agent *AgentUpdateAgent `json:"agent"`
}

type AgentUpdateAgent struct {
	ID          uint64 `json:"id_agent"`
	Name        string `json:"name"`
	CompanyID   string `json:"company_id"`
	Description string `json:"description"`
	Logo        string `json:"Logo"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	Address     string `json:"address"`
	Site        string `json:"site"`
	UpdatedAt   string `json:"updated_at"`
	CreatedAt   string `json:"created_at"`
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
	ID          uint64 `json:"id_agent"`
	Name        string `json:"name"`
	CompanyID   string `json:"company_id"`
	Description string `json:"description"`
	Logo        string `json:"Logo"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	Address     string `json:"address"`
	Site        string `json:"site"`
	UpdatedAt   string `json:"updated_at"`
	CreatedAt   string `json:"created_at"`
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
	Awatar      string `json:"awatar"`
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
	ID          uint64 `json:"id_agent"`
	Name        string `json:"name"`
	CompanyID   string `json:"company_id"`
	Description string `json:"description"`
	Logo        string `json:"Logo"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	Address     string `json:"address"`
	Site        string `json:"site"`
	UpdatedAt   string `json:"updated_at"`
	CreatedAt   string `json:"created_at"`
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
	Awatar      string `json:"awatar"`
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
type PayslipCreateV1Args struct {
	EmployeeFK uint64                   `json:"fk_employee"`
	Payslip    PayslipCreatePayslipData `json:"payslip"`
}

type PayslipCreatePayslipData struct {
	Currency string  `json:"currency"`
	Amount   float64 `json:"amount"`
	Data     string  `json:"data"`
}

// easyjson:json
type PayslipCreateV1Res struct {
	Employee *PayslipCreateEmployee `json:"agent"`
	Payslip  *PayslipCreatePayslip  `json:"agent"`
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
	Currency   string  `json:"currency"`
	Amount     float64 `json:"amount"`
	Data       []uint8 `json:"data"`
	UpdateAt   string  `json:"updated_at"`
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
type PayslipSetV1Args struct {
	ID               *uint64               `json:"payslip_id,omitempty"`
	Employee         PayslipSetPerson      `json:"employee"`
	ProcessedByUser  *PayslipSetPerson     `json:"processed_by_user,omitempty"`
	Agent            PayslipSetCompany     `json:"agent"`
	ProcessedByAgent *PayslipSetCompany    `json:"processed_by_agent,omitempty"`
	Data             PayslipSetPayslipData `json:"data"`
}

type PayslipSetPerson struct {
	ID          *uint64            `json:"id,omitempty"`
	Name        string             `json:"name"`
	RoleDesc    string             `json:"p_description"`
	Description string             `json:"description"`
	Contacts    PayslipSetContacts `json:"contacts"`
	Avatar      string             `json:"awatar"`
}

type PayslipSetContacts struct {
	Address string `json:"address"`
	Phone   string `json:"phone"`
	Email   string `json:"email"`
	Site    string `json:"site"`
}

type PayslipSetCompany struct {
	ID          *uint64            `json:"id,omitempty"`
	Name        string             `json:"name"`
	RoleDesc    string             `json:"c_description"`
	Description string             `json:"description"`
	Contacts    PayslipSetContacts `json:"contacts"`
	BGImage     string             `json:"bg_image"`
	Logo        string             `json:"logo"`
}

type PayslipSetPayslipData struct {
	Currency string              `json:"currency"`
	Payslip  PayslipSetOperation `json:"payslip"`
	Details  PayslipSetOperation `json:"details"`
}

type PayslipSetOperation struct {
	Amount  *float64               `json:"amount,omitempty"`
	Float   *float64               `json:"float,omitempty"`
	Int     *int64                 `json:"int,omitempty"`
	Text    *string                `json:"text,omitempty"`
	Title   *string                `json:"title,omitempty"`
	Details *[]PayslipSetOperation `json:"details,omitempty"`
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
	ID          uint64 `json:"id_agent"`
	Name        string `json:"name"`
	CompanyID   string `json:"company_id"`
	Description string `json:"description"`
	Logo        string `json:"Logo"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	Address     string `json:"address"`
	Site        string `json:"site"`
	UpdatedAt   string `json:"updated_at"`
	CreatedAt   string `json:"created_at"`
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
	Awatar      string `json:"awatar"`
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
	ID          uint64 `json:"id_agent"`
	Name        string `json:"name"`
	CompanyID   string `json:"company_id"`
	Description string `json:"description"`
	Logo        string `json:"Logo"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	Address     string `json:"address"`
	Site        string `json:"site"`
	UpdatedAt   string `json:"updated_at"`
	CreatedAt   string `json:"created_at"`
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
	Awatar      string `json:"awatar"`
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
	Name        string  `json:"name"`
	Short       string  `json:"p_description"`
	Description string  `json:"description"`
	Awatar      string  `json:"awatar"`
	Phone       string  `json:"phone"`
	Email       string  `json:"email"`
	Password    *string `json:"password,omitempty"`
}

// easyjson:json
type UserCreateV1Res struct {
	Token string          `json:"token"`
	User  *UserCreateUser `json:"user"`
}

type UserCreateUser struct {
	ID          uint64 `json:"id_user"`
	Name        string `json:"name"`
	Short       string `json:"p_description"`
	Description string `json:"description"`
	Awatar      string `json:"awatar"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	UpdatedAt   string `json:"updated_at"`
	CreatedAt   string `json:"created_at"`
}

type UserCreateV1Errors int

const (
	UserCreateV1Errors_USER_EXISTS = iota
	UserCreateV1Errors_CREATE_FAILED
	UserCreateV1Errors_USER_NOT_CREATED
)

var _UserCreateV1ErrorsMapping = map[string]int{
	"USER_EXISTS":      UserCreateV1Errors_USER_EXISTS,
	"CREATE_FAILED":    UserCreateV1Errors_CREATE_FAILED,
	"USER_NOT_CREATED": UserCreateV1Errors_USER_NOT_CREATED,
}

// easyjson:json
type UserLoginV1Args struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// easyjson:json
type UserLoginV1Res struct {
	Token string         `json:"token"`
	User  *UserLoginUser `json:"user"`
}

type UserLoginUser struct {
	ID          uint64 `json:"id_user"`
	Name        string `json:"name"`
	Short       string `json:"p_description"`
	Description string `json:"description"`
	Awatar      string `json:"awatar"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	UpdatedAt   string `json:"updated_at"`
	CreatedAt   string `json:"created_at"`
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
type UserUpdateV1Args struct {
	ID          uint64  `json:"id_user"`
	Name        *string `json:"name,omitempty"`
	Short       *string `json:"p_description,omitempty"`
	Description *string `json:"description,omitempty"`
	Awatar      *string `json:"awatar,omitempty"`
	Phone       *string `json:"phone,omitempty"`
	Email       *string `json:"email,omitempty"`
	Password    *string `json:"password,omitempty"`
}

// easyjson:json
type UserUpdateV1Res struct {
	Token string          `json:"token"`
	User  *UserUpdateUser `json:"user"`
}

type UserUpdateUser struct {
	ID          uint64 `json:"id_user"`
	Name        string `json:"name"`
	Short       string `json:"p_description"`
	Description string `json:"description"`
	Awatar      string `json:"awatar"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	UpdatedAt   string `json:"updated_at"`
	CreatedAt   string `json:"created_at"`
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
