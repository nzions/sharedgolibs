package ca

import (
	"net/http"
	"testing"
)

func TestSecureHTTPSServer_ListenAndServe(t *testing.T) {
	srv := &http.Server{}
	secureSrv := NewSecureHTTPSServer(srv)

	err := secureSrv.ListenAndServe()
	if err == nil {
		t.Fatal("expected error when calling ListenAndServe, got nil")
	}
	if err.Error() != "cannot call ListenAndServe() on SecureHTTPSServer - use ListenAndServeTLS()" {
		t.Errorf("unexpected error: %v", err)
	}
}
