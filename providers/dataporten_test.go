package providers

import (
	"github.com/bmizerany/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func testDataPortenProvider(hostname string) *DataPortenProvider {
	p := NewDataPortenProvider(
		&ProviderData{
			ProviderName: "",
			LoginURL:     &url.URL{},
			RedeemURL:    &url.URL{},
			ProfileURL:   &url.URL{},
			Scope:        ""})
	if hostname != "" {
		updateURL(p.Data().LoginURL, hostname)
		updateURL(p.Data().RedeemURL, hostname)
		updateURL(p.Data().ProfileURL, hostname)
	}
	return p
}

func TestDataPortenProviderDefaults(t *testing.T) {
	p := NewDataPortenProvider(
		&ProviderData{
			LoginURL: &url.URL{
				Scheme: "https",
				Host:   "example.com",
				Path:   "/oauth/authorization"},
			RedeemURL: &url.URL{
				Scheme: "https",
				Host:   "example.com",
				Path:   "/oauth/token"},
			ProfileURL: &url.URL{
				Scheme: "https",
				Host:   "example.com",
				Path:   "/userinfo"},
		})
	assert.NotEqual(t, nil, p)
	assert.Equal(t, "DataPorten", p.Data().ProviderName)
	assert.Equal(t, "https://example.com/oauth/authorization",
		p.Data().LoginURL.String())
	assert.Equal(t, "https://example.com/oauth/token",
		p.Data().RedeemURL.String())
	assert.Equal(t, "https://example.com/userinfo",
		p.Data().ProfileURL.String())
	assert.Equal(t, "https://example.com/userinfo",
		p.Data().ValidateURL.String())
	assert.Equal(t, "groups userid", p.Data().Scope)
}

func testDataPortenBackend(userInfo string, groupInfo string) *httptest.Server {
	userinfoPath := "/userinfo"
	groupPath := "/groups/me/groups"

	return httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			url := r.URL
			if r.Header.Get("Authorization") != "Bearer imaginary_access_token" {
				w.WriteHeader(403)
			} else if url.Path == userinfoPath {
				w.WriteHeader(200)
				w.Write([]byte(userInfo))
			} else if url.Path == groupPath {
				w.WriteHeader(200)
				w.Write([]byte(groupInfo))
			} else {
				w.WriteHeader(404)
			}
		}))
}

func TestDataPortenProviderGetEmailAddress(t *testing.T) {
	b := testDataPortenBackend(
		`{"user":{"userid_sec":[], "email": "ola.norman@norge.no", "userid":"0923894sd","name":"Ola Norman"},"audience":"c110bb27-b7b8-44d3-8f52-f87203d8ff59"}`,
		`[{"id":"testgroup"}]`)
	defer b.Close()

	b_url, _ := url.Parse(b.URL)
	p := testDataPortenProvider(b_url.Host)

	session := &SessionState{AccessToken: "imaginary_access_token"}
	userid, err := p.GetEmailAddress(session)
	assert.Equal(t, nil, err)
	assert.Equal(t, "0923894sd", userid)

	// Test groups permissions
	p.GroupsURL = &url.URL{
		Scheme: "http",
		Host:   b_url.Host,
		Path:   "/groups/me/groups",
	}
	p.SetGroups("testgroup", "")
	userid, err = p.GetEmailAddress(session)
	assert.Equal(t, nil, err)
	assert.Equal(t, "0923894sd", userid)

	p.SetGroups("testgroup-fail", "")
	userid, err = p.GetEmailAddress(session)
	assert.NotEqual(t, "0923894sd", userid)
}

func TestDataPortenProviderGetEmailAddressMASGroups(t *testing.T) {
	b := testDataPortenBackend(
		`{"user":{"userid_sec":[], "email": "ola.norman@norge.no", "userid":"0923894sd","name":"Ola Norman"},"audience":"c110bb27-b7b8-44d3-8f52-f87203d8ff59"}`,
		`[{"id":"NS98237K"}]`)
	defer b.Close()

	b_url, _ := url.Parse(b.URL)
	p := testDataPortenProvider(b_url.Host)

	session := &SessionState{AccessToken: "imaginary_access_token"}
	userid, err := p.GetEmailAddress(session)
	assert.Equal(t, nil, err)
	assert.Equal(t, "0923894sd", userid)

	// Test groups permissions fetch from MASGroupsURL now
	p.SetGroups("NS98237K", "http://"+b_url.Host+"/groups/me/groups")
	userid, err = p.GetEmailAddress(session)
	assert.Equal(t, nil, err)
	assert.Equal(t, "0923894sd", userid)

	p.SetGroups("NS98237K-fail", "")
	userid, err = p.GetEmailAddress(session)
	assert.NotEqual(t, "0923894sd", userid)
}
