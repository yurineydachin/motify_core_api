package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"motify_core_api/godep_libs/service/config"
	"motify_core_api/godep_libs/service/logger"
	"motify_core_api/godep_libs/service/utils"
	"motify_core_api/godep_libs/service/watcher"
)

const (
	codeCookieName      = `backoffice_code`
	emailHeaderName     = `X-Auth-Email`
	authNotUsedDuration = 3 * 24 * time.Hour
)

var (
	admAuthURL     = "https://my.lazada.com"
	authorizations = make(map[string]*authorization)
	authMutex      sync.RWMutex
	initOnce       sync.Once
	admAuthEnabled bool
)

func init() {
	config.RegisterBool("adm-auth-enabled", "Admin panel authentication enabled", false)
	//	config.RegisterString("service_auth_url", "Service authenticator url", "https://my.lazada.com")
}

func initAuth() {
	initOnce.Do(func() {
		admAuthEnabled, _ = config.GetBool("adm-auth-enabled")
		//		admAuthURL, _ = config.GetString("service_auth_url")
		if admAuthEnabled {
			watcher.WatchForever(cleanAuthorization, 12*time.Hour)
		}
	})
}

func accessHandler(h http.Handler) http.Handler {
	initAuth()
	if !admAuthEnabled {
		return h
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		email := getAuthHeaderValue(r)
		if len(email) == 0 {
			logger.Error(nil, "No email in header '%s'", emailHeaderName)
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func requireAuth(f func(w http.ResponseWriter, r *http.Request, email string)) http.HandlerFunc {
	initAuth()
	return func(w http.ResponseWriter, r *http.Request) {
		f(w, r, getAuthHeaderValue(r))
	}
}

func getAuthHeaderValue(r *http.Request) string {
	if admAuthEnabled {
		return strings.TrimSpace(r.Header.Get(emailHeaderName))
	}
	return "no-email-as-auth-disabled"
}

func authHandler(h http.Handler) http.Handler {
	initAuth()
	if !admAuthEnabled {
		return h
	}
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/login/" || authValid(w, req) {
			h.ServeHTTP(w, req)
		} else {
			redirect(w, req)
		}
	})
}

func authValid(w http.ResponseWriter, r *http.Request) bool {
	if !admAuthEnabled {
		return true
	}

	var auth *authorization
	for _, cookie := range r.Cookies() {
		if cookie.Name == codeCookieName {
			code := cookie.Value
			auth = getAndProlongAuthorization(code)
			if auth != nil {
				break
			}
		}
	}

	if auth == nil {
		return false
	}

	if auth.allowed {
		r.Header.Set(emailHeaderName, auth.email)
	}

	return auth.allowed
}

type answerAllowed struct {
	Valid struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Mobile    string `json:"mobile"`
		Email     string `json:"email"`
		Code      string `json:"code"`
	} `json:"valid"`
}

type answerForbidden struct {
	Valid int8 `json:"valid"`
}

func login(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get(codeCookieName)
	if len(code) == 0 {
		redirect(w, r)
		return
	}

	body, err := utils.DoGetRequest(http.DefaultClient, admAuthURL+`/validate?code=`+code)
	if err != nil {
		logger.Error(nil, "Error on request to validator: %s", err)
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	var profile answerAllowed
	err = json.Unmarshal(body, &profile)
	if err != nil {
		var forbidden answerForbidden
		err := json.Unmarshal(body, &forbidden)
		if err != nil {
			logger.Error(nil, "Error on validator answer unmarshaling: %s", err)
			return
		}

		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	if profile.Valid.Code != code {
		logger.Error(nil, "The code got from '%s' doesn't correspond to the queried one: %s != %s", admAuthURL, profile.Valid.Code, code)
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:   codeCookieName,
		Value:  code,
		Path:   `/`,
		Domain: r.Host,
	})

	http.Redirect(w, r, `/`, http.StatusFound)

	auth := registerAuthorization(code)
	auth.email = profile.Valid.Email
	auth.allowed = true

	return
}

func redirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, fmt.Sprintf("%s?redirect_uri=http://%s/login/", admAuthURL, r.Host), http.StatusFound)
}

func getAndProlongAuthorization(code string) *authorization {
	auth := getAuthorization(code)
	if auth != nil {
		auth.updateAccessTime(code)
	}
	return auth
}

func registerAuthorization(code string) *authorization {
	authMutex.Lock()
	defer authMutex.Unlock()

	if auth, ok := authorizations[code]; ok {
		return auth
	}

	auth := newAuthorization()
	authorizations[code] = auth

	return auth
}

func getAuthorization(code string) *authorization {
	authMutex.RLock()
	defer authMutex.RUnlock()

	return authorizations[code]
}

func deleteAuthorization(code string) {
	authMutex.Lock()
	defer authMutex.Unlock()

	delete(authorizations, code)
}

func cleanAuthorization() {
	codes := getCodes()
	for _, code := range codes {
		auth := getAuthorization(code)
		if auth.isExpired() {
			deleteAuthorization(code)
		}
	}
}

func getCodes() []string {
	authMutex.RLock()
	defer authMutex.RUnlock()

	codes := make([]string, 0, len(authorizations))
	for code := range authorizations {
		codes = append(codes, code)
	}

	return codes
}

type authorization struct {
	allowed        bool
	email          string
	lastAccessTime time.Time
	mutex          sync.RWMutex
	once           sync.Once
}

func newAuthorization() *authorization {
	return &authorization{
		allowed:        false,
		lastAccessTime: time.Now(),
	}
}

func (auth *authorization) updateAccessTime(code string) {
	auth.mutex.Lock()
	defer auth.mutex.Unlock()

	auth.lastAccessTime = time.Now()
}

func (auth *authorization) isExpired() bool {
	auth.mutex.RLock()
	defer auth.mutex.RUnlock()

	return time.Since(auth.lastAccessTime) > authNotUsedDuration
}
