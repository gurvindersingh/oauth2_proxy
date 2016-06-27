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
	allowedGroups []string
	GroupsURL     *url.URL
	MASGroupsURL  *url.URL
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
		p.Scope = "groups userid"
	}
	return &DataPortenProvider{ProviderData: p}
}

// GetEmailAddress returns authenticated user's ID. We can't get email address from
// every vendor and thus chose to use dataporten User ID
func (p *DataPortenProvider) GetEmailAddress(s *SessionState) (string, error) {

	// Make sure user is member of the allowed groups before continueing
	if p.allowedGroups != nil {
		if !p.isMember(s.AccessToken) {
			return "", nil
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

	if data, ok := json.Get("user").CheckGet("userid"); ok {
		userid, err := data.String()
		if err != nil {
			log.Printf("Failed in parsing userid", err)
			return "", err
		}
		return userid, nil
	}
	return "", errors.New("Failed in getting userid")
}

// SetGroups initialized the groups information to authorize only
// group members. If no groups are provided, any authenticated users
// is allowed
func (p *DataPortenProvider) SetGroups(groups string, masGroupsURL string) {
	if groups != "" {
		p.allowedGroups = strings.Split(groups, ",")
		// IF groups
		if p.GroupsURL == nil {
			p.GroupsURL = &url.URL{
				Scheme: "https",
				Host:   "groups-api.dataporten.no",
				Path:   "/groups/me/groups",
			}
		}
		if masGroupsURL != "" {
			url, err := url.Parse(masGroupsURL)
			if err != nil {
				log.Printf("Failed in parsing MASGroupsURL %s", masGroupsURL)
				return
			}
			p.MASGroupsURL = url
		}
	}
}

// Fetch groups from dataporten and MAS for given token
func getGroups(
	token string,
	url string) ([]string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("failed in building group request", err)
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("%s %s %s", req.Method, req.URL, err)
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("got %d", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	json, err := simplejson.NewJson(body)
	if err != nil {
		log.Printf("failed in fetching groups info", err)
		return nil, err
	}
	jsonGroups, err := json.Array()
	if err != nil {
		log.Printf("failed in parsing groups info", err)
		return nil, err
	}

	var groups []string
	for _, grp := range jsonGroups {
		groups = append(groups, grp.(map[string]interface{})["id"].(string))
	}

	return groups, nil
}

// see if the user is member of any allowed groups
func (p *DataPortenProvider) isMember(token string) bool {
	groups, err := getGroups(token, p.GroupsURL.String())
	if err != nil {
		log.Printf("failed in getting groups info from Dataporten", err)
	}
	if p.MASGroupsURL != nil {
		masGroups, err := getGroups(token, p.MASGroupsURL.String())
		if err != nil {
			log.Printf("failed in getting groups info from MAS", err)
		} else {
			groups = append(groups, masGroups...)
		}
	}
	for _, group := range groups {
		for _, allowedGrp := range p.allowedGroups {
			if allowedGrp == group {
				return true
			}
		}
	}
	return false
}
