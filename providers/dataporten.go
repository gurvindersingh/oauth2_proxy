package providers

import (
	"errors"
	"fmt"
	"github.com/bitly/go-simplejson"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type DataPortenProvider struct {
	*ProviderData
	groups    []string
	GroupsURL *url.URL
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
		p.Scope = "email groups"
	}
	return &DataPortenProvider{ProviderData: p}
}

func (p *DataPortenProvider) GetEmailAddress(s *SessionState) (string, error) {

	// Make sure user is member of the allowed groups before continueing
	if p.groups != nil {
		if ok, err := p.isMember(s.AccessToken); err != nil || !ok {
			return "", err
		}
	}

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

// SetGroups initialized the groups information to authorize only
// group members. If no groups are provided, any authenticated users
// is allowed
func (p *DataPortenProvider) SetGroups(groups string) {
	if groups != "" {
		p.GroupsURL = &url.URL{
			Scheme: "https",
			Host:   "groups-api.dataporten.no",
			Path:   "/groups/me/groups",
		}
		p.groups = strings.Split(groups, ",")
		fmt.Println("Groups", p.groups)
	}
}

// Fetch groups from dataporten and see if the user is member of
// any allowed groups
func (p *DataPortenProvider) isMember(token string) (bool, error) {
	req, err := http.NewRequest("GET", p.GroupsURL.String(), nil)
	if err != nil {
		log.Printf("failed in building group request", err)
		return false, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("%s %s %s", req.Method, req.URL, err)
		return false, err
	}

	if resp.StatusCode != 200 {
		return false, fmt.Errorf("got %d", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	json, err := simplejson.NewJson(body)
	if err != nil {
		log.Printf("failed in fetching groups info", err)
		return false, err
	}

	groups, err := json.Array()
	if err != nil {
		log.Printf("failed in parsing groups info", err)
		return false, err
	}
	for _, group := range groups {
		for _, allowedGrp := range p.groups {
			if allowedGrp == group.(map[string]interface{})["id"] {
				return true, nil
			}
		}

	}
	return false, nil
}
