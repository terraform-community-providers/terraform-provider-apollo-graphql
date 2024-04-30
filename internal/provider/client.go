package provider

import "net/http"

type authedTransport struct {
	token   string
	wrapped http.RoundTripper
}

func (t *authedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("X-API-Key", t.token)

	return t.wrapped.RoundTrip(req)
}
