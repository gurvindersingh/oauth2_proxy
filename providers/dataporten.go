package providers

import (
	"errors"
	"fmt"
	"github.com/bitly/go-simplejson"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

type DataPortenProvider struct {
	*ProviderData
}

func NewDataPortenProvider(p *ProviderData) *DataPortenProvider {
	const dataportenHost string = "auth.dataporten.no"
	const scheme = "https"

	p.ProviderName = "DataPorten"
	if p.LoginURL == nil || p.LoginURL.String() == "" {
		p.LoginURL = &url.URL{
			Scheme: scheme,
			Host:   dataportenHost,
			Path:   "/oauth/authorization",
		}
	}
	if p.RedeemURL == nil || p.RedeemURL.String() == "" {
		p.RedeemURL = &url.URL{
			Scheme: scheme,
			Host:   dataportenHost,
			Path:   "/oauth/token",
		}
	}
	if p.ProfileURL == nil || p.ProfileURL.String() == "" {
		p.ProfileURL = &url.URL{
			Scheme: scheme,
			Host:   dataportenHost,
			Path:   "/userinfo",
		}
	}
	// ValidationURL is the API Base URL
	if p.ValidateURL == nil || p.ValidateURL.String() == "" {
		p.ValidateURL = p.ProfileURL
	}

	if p.Scope == "" {
		p.Scope = "email"
	}
	return &DataPortenProvider{ProviderData: p}
}

func (p *DataPortenProvider) GetEmailAddress(s *SessionState) (string, error) {

	req, err := http.NewRequest("GET", p.ProfileURL.String(), nil)
	if err != nil {
		log.Printf("failed in building request", err)
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+s.AccessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("%s %s %s", req.Method, req.URL, err)
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("got %d", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	json, err := simplejson.NewJson(body)
	if err != nil {
		log.Printf("failed in fetching user info", err)
		return "", err
	}

	if data, ok := json.Get("user").CheckGet("email"); ok {
		return data.String()
	}
	return "", errors.New("Failed in getting email")
}
