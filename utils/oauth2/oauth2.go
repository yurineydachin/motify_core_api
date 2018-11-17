package oauth2

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
	"golang.org/x/oauth2/google"
)

var (
	configs = map[string]*oauth2.Config{
		"fb": &oauth2.Config{
			ClientID:     "841018762765680",
			ClientSecret: "45d763707855c3fce3b2299aa461c5fa",
			RedirectURL:  "https://mobile-api.motifyapp.com/user/social/fb_login/v1",
			Scopes:       []string{"public_profile"},
			Endpoint:     facebook.Endpoint,
		},
		"google": &oauth2.Config{
			ClientID:     "882675778848-9vdopoutu2b2sg2uu11dfomkki7h61hu.apps.googleusercontent.com",
			ClientSecret: "mYYVDdrLlQEcELT1y_kccqsQ",
			RedirectURL:  "https://mobile-api.motifyapp.com/user/social/google_login/v1",
			Scopes: []string{
				"https://www.googleapis.com/auth/plus.login",
				"https://www.googleapis.com/auth/plus.me",
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		},
	}
)

const (
	FBConf     = "fb"
	GoogleConf = "google"

	FBProfileUrl     = "https://graph.facebook.com/me"
	GoogleProfileUrl = "https://www.googleapis.com/oauth2/v1/userinfo"
)

type Profile struct {
	AccessToken string `json:"-,omitempty"`
	ID          string `json:"id"`
	Name        string `json:"name"`
	Avatar      string `json:"picture"`
	/*
	   first_name
	   last_name
	   middle_name
	   name_format
	   picture
	   short_name
	*/
}

func getConf(social string) *oauth2.Config {
	if conf, found := configs[social]; found && conf != nil {
		return conf
	}
	return &oauth2.Config{
		ClientID:     "",
		ClientSecret: "",
		RedirectURL:  "",
		Scopes:       []string{"public_profile"},
		Endpoint:     facebook.Endpoint,
	}
}

func GetCodeUrl(social string) (string, error) {
	oauthConf := getConf(social)
	res, err := url.Parse(oauthConf.Endpoint.AuthURL)
	if err != nil {
		return "", fmt.Errorf("GetCodeUrl, url.ParseUrl: %s", err)
	}
	params := url.Values{}
	params.Add("client_id", oauthConf.ClientID)
	params.Add("scope", strings.Join(oauthConf.Scopes, " "))
	params.Add("redirect_uri", oauthConf.RedirectURL)
	params.Add("response_type", "code")
	params.Add("state", "somenotverysecretstring")
	res.RawQuery = params.Encode()
	return res.String(), nil
}

func GetAccessTokenByCode(social, code string) (string, error) {
	oauthConf := getConf(social)
	if code == "" {
		return "", fmt.Errorf("No code for access_token")
	}
	token, err := oauthConf.Exchange(oauth2.NoContext, code)
	if err != nil {
		return "", fmt.Errorf("oauthConf.Exchange() failed with '%s'\n", err)
	}
	return token.AccessToken, nil
}

func getFBProfileUrlByAccessToken(token string) (string, error) {
	res, err := url.Parse(FBProfileUrl)
	if err != nil {
		return "", fmt.Errorf("getFBProfileUrl, url.ParseUrl: %s", err)
	}
	params := url.Values{}
	params.Add("access_token", token)
	res.RawQuery = params.Encode()
	return res.String(), nil
}

func GetFBUser(code, accessToken *string) (*Profile, error) {
	var (
		FBUrl string
		err   error
	)
	if accessToken != nil && *accessToken != "" {
		FBUrl, err = getFBProfileUrlByAccessToken(*accessToken)
	} else if code != nil && *code != "" {
		accessTokenFB, err := GetAccessTokenByCode(FBConf, *code)
		if err != nil {
			return nil, fmt.Errorf("getFBProfileUrl, GetAccessTokenByCode: %s", err)
		}
		accessToken = &accessTokenFB
		FBUrl, err = getFBProfileUrlByAccessToken(accessTokenFB)
	} else {
		return nil, fmt.Errorf("There are no code or access_token")
	}
	if err != nil {
		return nil, fmt.Errorf("GetFBUser: %s", err)
	}

	resp, err := http.Get(FBUrl)
	if err != nil {
		return nil, fmt.Errorf("GetFBUser, http.Get: %s", err)
	}
	defer resp.Body.Close()

	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("GetFBUser, ioutil.ReadAll: %s\n", err)
	}

	var profile Profile
	err = json.Unmarshal(response, &profile)
	if err != nil {
		return nil, fmt.Errorf("GetFBUser, json.Unmarshal: %s - text: %s\n", err, response)
	}
	if accessToken != nil {
		profile.AccessToken = *accessToken
	}
	return &profile, nil
}

func getGoogleProfileUrlByAccessToken(token string) (string, error) {
	res, err := url.Parse(GoogleProfileUrl)
	if err != nil {
		return "", fmt.Errorf("getGoogleProfileUrl, url.ParseUrl: %s", err)
	}
	params := url.Values{}
	params.Add("access_token", token)
	res.RawQuery = params.Encode()
	return res.String(), nil
}

func GetGoogleUser(code, accessToken *string) (*Profile, error) {
	var (
		GoogleUrl string
		err       error
	)
	if accessToken != nil && *accessToken != "" {
		GoogleUrl, err = getGoogleProfileUrlByAccessToken(*accessToken)
	} else if code != nil && *code != "" {
		accessTokenGoogle, err := GetAccessTokenByCode(GoogleConf, *code)
		if err != nil {
			return nil, fmt.Errorf("getGoogleProfileUrl, GetAccessTokenByCode: %s", err)
		}
		accessToken = &accessTokenGoogle
		GoogleUrl, err = getGoogleProfileUrlByAccessToken(accessTokenGoogle)
	} else {
		return nil, fmt.Errorf("There are no code or access_token")
	}
	if err != nil {
		return nil, fmt.Errorf("GetGoogleUser: %s", err)
	}

	resp, err := http.Get(GoogleUrl)
	if err != nil {
		return nil, fmt.Errorf("GetGoogleUser, http.Get: %s", err)
	}
	defer resp.Body.Close()

	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("GetGoogleUser, ioutil.ReadAll: %s\n", err)
	}

	var profile Profile
	err = json.Unmarshal(response, &profile)
	if err != nil {
		return nil, fmt.Errorf("GetGoogleUser, json.Unmarshal: %s - text: %s\n", err, response)
	}
	if accessToken != nil {
		profile.AccessToken = *accessToken
	}
	return &profile, nil
}
