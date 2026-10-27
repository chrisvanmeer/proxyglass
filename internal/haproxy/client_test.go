package haproxy

import (
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
	fes, bes := parseCSV(data, "edge-a")
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
