package model

import "time"

type Target struct {
	Name      string
	URL       string
	Username  string
	Password  string
	VerifyTLS bool
}

type Server struct {
	Name        string
	Backend     string
	Target      string
	Status      string
	CheckStatus string
	CheckCode   string
	LastChk     string
	LastChkDur  string
	LastSess    string
	SessionsCur int64
	SessionsMax int64
	SessionsLim string
	BytesIn     int64
	BytesOut    int64
	ErrorsReq   string
	ErrorsConn  string
	ErrorsResp  string
	Wredis      string
	Wretr       string
}

type Backend struct {
	Name         string
	Target       string
	Mode         string
	Status       string
	CheckStatus  string
	CheckCode    string
	LastChk      string
	LastChkDur   string
	SessionsCur  int64
	SessionsMax  int64
	SessionsTot  int64
	LastSess     string
	BytesIn      int64
	BytesOut     int64
	LastActivity time.Time
	Servers      []Server
}

type Frontend struct {
	Name         string
	Target       string
	Mode         string
	Status       string
	SessionsCur  int64
	SessionsMax  int64
	SessionsTot  int64
	Rate         int64
	RateMax      int64
	BytesIn      int64
	BytesOut     int64
	DeniedReq    int64
	DeniedResp   int64
	ErrorsReq    int64
	LastActivity time.Time
}

type Snapshot struct {
	Targets   []string
	Frontends []Frontend
	Backends  []Backend
	FetchedAt time.Time
	Errors    map[string]error
}
