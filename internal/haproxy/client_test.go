package haproxy

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"testing"
)

func TestParseCSVUsesNamedColumns(t *testing.T) {
	data := `# pxname,svname,type,scur,smax,bin,bout,status,check_status,check_code,check_duration,last_chk,lastsess
web,FRONTEND,0,3,20,1024,2048,OPEN,,,,,
web,web1,2,2,10,256,512,UP,L7OK,200,4,Layer7 check passed,16
web,web-down,2,0,10,0,0,DOWN,L4CON,,0,Connection refused,25
web,BACKEND,1,2,20,512,1024,UP,,,,,25
stats,FRONTEND,0,0,0,0,0,OPEN,,,,,
stats,BACKEND,1,0,0,0,0,UP,,,,,
`
	fes, bes, err := parseCSV(data, "edge-a")
	if err != nil {
		t.Fatalf("parseCSV returned error: %v", err)
	}
	if len(fes) != 1 || fes[0].Name != "web" || fes[0].SessionsCur != 3 {
		t.Fatalf("unexpected frontends: %#v", fes)
	}
	if len(bes) != 1 || len(bes[0].Servers) != 2 || bes[0].LastSess != "25" {
		t.Fatalf("unexpected backends: %#v", bes)
	}
	if got := bes[0].Servers[0]; got.Name != "web1" || got.CheckStatus != "L7OK" || got.CheckCode != "200" || got.LastChk != "Layer7 check passed" || got.LastSess != "16" {
		t.Fatalf("unexpected server stats: %#v", got)
	}
	if got := bes[0].Servers[1]; got.Status != "DOWN" || !strings.Contains(got.LastChk, "Connection refused") {
		t.Fatalf("expected failure detail, got %#v", got)
	}
}

func TestParseCSVReportsMalformedCSV(t *testing.T) {
	if _, _, err := parseCSV("# pxname,svname,type\nweb,\"unterminated", "edge-a"); err == nil {
		t.Fatal("malformed CSV should be reported")
	}
}

func TestParseCSVPreservesBackendRowOrder(t *testing.T) {
	data := `# pxname,svname,type,status
backend-b,BACKEND,1,UP
backend-b,b1,2,UP
backend-a,BACKEND,1,UP
backend-a,a1,2,UP
`
	_, backends, err := parseCSV(data, "edge-a")
	if err != nil {
		t.Fatal(err)
	}
	if len(backends) != 2 || backends[0].Name != "backend-b" || backends[1].Name != "backend-a" {
		t.Fatalf("backend order = %#v, want backend-b then backend-a", backends)
	}
}

func TestRequestErrorSummaryRecognizesTimeoutAndRedactsURL(t *testing.T) {
	err := &url.Error{Op: "Get", URL: "https://reader:secret@example.test/stats;csv", Err: context.DeadlineExceeded}
	if !isRequestTimeout(err) {
		t.Fatal("deadline exceeded should be recognized as a request timeout")
	}
	if got := requestErrorMessage(err); strings.Contains(got, "reader") || strings.Contains(got, "secret") || strings.Contains(got, "example.test") {
		t.Fatalf("request error summary leaked URL credentials or target URL: %q", got)
	}
	if isRequestTimeout(errors.New("connection refused")) {
		t.Fatal("connection refused should not be classified as a timeout")
	}
}
