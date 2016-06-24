package providers

import (
	"net/url"
	"testing"

	"github.com/bmizerany/assert"
)

func testDataPortenProvider(hostname string) *DataPortenProvider {
	p := NewDataPortenProvider(
		&ProviderData{
			ProviderName: "",
			LoginURL:     &url.URL{},
			RedeemURL:    &url.URL{},
			ValidateURL:  &url.URL{},
			Scope:        ""})
	if hostname != "" {
		updateURL(p.Data().LoginURL, hostname)
		updateURL(p.Data().RedeemURL, hostname)
		updateURL(p.Data().ProfileURL, hostname)
		updateURL(p.Data().ValidateURL, hostname)
	}
	return p
}

func TestDataPortenProviderDefaults(t *testing.T) {
	p := testDataPortenProvider("")
	assert.NotEqual(t, nil, p)
}
