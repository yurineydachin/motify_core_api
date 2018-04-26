package admin

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	host            = `thehost`
	email           = `dmitriy.lukianchuk@lazada.com`
	allowedTemplate = `{"valid":{"id":5216,"first_name":"Dmitriy","last_name":"Lukianchuk","email":"%s","mobile":"+79261523471","country_head":"Tech Hub","ldap_id":null,"ldap_name":null,"department":"Platform","ldap_password_changed_at":"2015-09-14 10:06:33","code":"%s"}}`
	forbidden       = `{"valid":0}`
)

func TestAuth(t *testing.T) {
	code := `5f3d8389de38866accef9228b14459d7`

	mux := http.NewServeMux()
	mux.HandleFunc("/login/", login)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {})

	validationServer := myLazadaAPI(code)
	defer validationServer.Close()

	initAuth()
	admAuthEnabled = true
	admAuthURL = validationServer.URL

	if codeCookieName != `backoffice_code` {
		t.Fatalf("Unexpected codeCookieName: %s", codeCookieName)
	}

	if emailHeaderName != `X-Auth-Email` {
		t.Fatalf("Unexpected emailHeaderName: %s", emailHeaderName)
	}

	cookie := &http.Cookie{
		Name:   codeCookieName,
		Value:  code,
		Path:   `/`,
		Domain: host,
	}
	{
		rw := httptest.NewRecorder()

		request, err := http.NewRequest(`GET`, `/`, nil)
		request.Host = host
		if err != nil {
			t.Fatal(err)
		}

		authHandler(mux).ServeHTTP(rw, request)

		if rw.Code != http.StatusFound {
			t.Errorf("Status must be: %d, Got: %d", http.StatusFound, rw.Code)
		}

		redirectExpected := fmt.Sprintf(`%s?redirect_uri=http://%s`, validationServer.URL, host+`/login/`)
		redirectGot := rw.HeaderMap.Get(`Location`)
		if redirectGot != redirectExpected {
			t.Errorf("The http client must be redirected to\n%s\texpected\n%s\tgot", redirectExpected, redirectGot)
		}
	}
	{
		rw := httptest.NewRecorder()

		request, err := http.NewRequest(`GET`, `/login/?`+codeCookieName+`=`+code+`wrongCode`, nil)
		if err != nil {
			t.Fatal(err)
		}
		request.Host = host

		authHandler(mux).ServeHTTP(rw, request)

		if rw.Code != http.StatusUnauthorized {
			t.Errorf("Status must be: %d, Got: %d", http.StatusUnauthorized, rw.Code)
		}

		if getAuthorization(code) != nil {
			t.Errorf("Authorization mustn't be saved")
		}
	}
	{
		rw := httptest.NewRecorder()

		request, err := http.NewRequest(`GET`, `/login/?`+codeCookieName+`=`+`WrongCode`, nil)
		if err != nil {
			t.Fatal(err)
		}
		request.Host = host

		authHandler(mux).ServeHTTP(rw, request)

		if rw.Code != http.StatusUnauthorized {
			t.Errorf("Status must be: %d, Got: %d", http.StatusUnauthorized, rw.Code)
		}

		if getAuthorization(code) != nil {
			t.Errorf("Authorization mustn't be saved")
		}
	}
	{
		rw := httptest.NewRecorder()

		request, err := http.NewRequest(`GET`, `/login/?`+codeCookieName+`=`+code, nil)
		if err != nil {
			t.Fatal(err)
		}
		request.Host = host

		authHandler(mux).ServeHTTP(rw, request)

		if rw.Code != http.StatusFound {
			t.Errorf("Status must be: %d, Got: %d", http.StatusFound, rw.Code)
		}

		redirectExpected := `/`
		redirectGot := rw.HeaderMap.Get(`Location`)
		if redirectGot != redirectExpected {
			t.Errorf("The http client must be redirected to\n%s\texpected\n%s\tgot", redirectExpected, redirectGot)
		}

		cookieString := rw.HeaderMap.Get(`Set-Cookie`)
		if cookieString != cookie.String() {
			t.Fatalf("Cookie backoffice_code must be set. Got: '%s'", cookieString)
		}

		auth := getAuthorization(code)
		if auth == nil {
			t.Fatalf("Authorization must be saved")
		}

		if auth.email != email {
			t.Errorf("Email must be: '%s'. Got: '%s'", email, auth.email)
		}

		if !auth.allowed {
			t.Errorf("Auth.allowed isn't true: %t", auth.allowed)
		}
	}
	{
		rw := httptest.NewRecorder()

		request, err := http.NewRequest(`GET`, `/`, nil)
		if err != nil {
			t.Fatal(err)
		}
		request.Host = host
		request.AddCookie(cookie)

		authHandler(mux).ServeHTTP(rw, request)

		if rw.Code != http.StatusOK {
			t.Errorf("Status must be: %d, Got: %d", http.StatusOK, rw.Code)
		}

		header := request.Header.Get(emailHeaderName)
		if header != email {
			t.Errorf("Header '%s' value '%s' doesn't match expected email '%s'", emailHeaderName, header, email)

		}
	}
}

func myLazadaAPI(code string) *httptest.Server {
	answerAllowed := fmt.Sprintf(allowedTemplate, email, code)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == `/validate` {
			theCode := r.URL.Query().Get(`code`)
			if len(theCode) >= len(code) {
				fmt.Fprintln(w, answerAllowed)
			} else {
				fmt.Fprintln(w, forbidden)
			}
		}
	}))

	return ts
}
