package prober

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPClientRoutesIdentityTrafficThroughConfiguredProxy(t *testing.T) {
	requests := make(chan string, 1)
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests <- r.URL.Host
		_, _ = fmt.Fprint(w, "proxy-egress")
	}))
	defer proxy.Close()

	client := NewProber(4, time.Second).WithProxy(proxy.URL).getHTTPClient()
	response, err := client.Get("http://identity.invalid/trace")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	body, _ := io.ReadAll(response.Body)
	if string(body) != "proxy-egress" {
		t.Fatalf("response body = %q", body)
	}
	select {
	case host := <-requests:
		if host != "identity.invalid" {
			t.Fatalf("proxy received host %q", host)
		}
	default:
		t.Fatal("configured proxy did not receive the identity request")
	}
}
