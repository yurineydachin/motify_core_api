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
)

var (
	oauthConf = &oauth2.Config{
		ClientID:     "841018762765680",
		ClientSecret: "45d763707855c3fce3b2299aa461c5fa",
		RedirectURL:  "https://mobile-api.motifyapp.com/user/social/fb_token/v1",
		Scopes:       []string{"public_profile"},
		Endpoint:     facebook.Endpoint,
	}
	oauthStateString = "somenotverysecretstring)"
)

const (
	FBProfileUrl = "https://graph.facebook.com/me"
)

// errors:
// This authorization code has expired
// This authorization code has been used

type FBProfile struct {
	AccessToken string `json:"-,omitempty"`
	ID          string `json:"id"`
	Name        string `json:"name"`
	/*
	   first_name
	   last_name
	   middle_name
	   name_format
	   picture
	   short_name
	*/
}

func GetCodeUrl() (string, error) {
	res, err := url.Parse(oauthConf.Endpoint.AuthURL)
	if err != nil {
		return "", fmt.Errorf("GetCodeUrl, url.ParseUrl: %s", err)
	}
	params := url.Values{}
	params.Add("client_id", oauthConf.ClientID)
	params.Add("scope", strings.Join(oauthConf.Scopes, " "))
	params.Add("redirect_uri", oauthConf.RedirectURL)
	params.Add("response_type", "code")
	params.Add("state", oauthStateString)
	res.RawQuery = params.Encode()
	return res.String(), nil
}

func GetAccessTokenByCode(code string) (string, error) {
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

func GetFBUser(code, accessToken *string) (*FBProfile, error) {
	var (
		FBUrl string
		err   error
	)
	if accessToken != nil && *accessToken != "" {
		FBUrl, err = getFBProfileUrlByAccessToken(*accessToken)
	} else if code != nil && *code != "" {
		accessTokenFB, err := GetAccessTokenByCode(*code)
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

	var profile FBProfile
	err = json.Unmarshal(response, &profile)
	if err != nil {
		return nil, fmt.Errorf("GetFBUser, json.Unmarshal: %s - text: %s\n", err, response)
	}
	if accessToken != nil {
		profile.AccessToken = *accessToken
	}
	return &profile, nil
}