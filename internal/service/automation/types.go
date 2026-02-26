package automation

type previewRunReq struct {
	Action string         `json:"action"`
	Params map[string]any `json:"params"`
}

type executeRunReq struct {
	ApprovalToken string         `json:"approval_token"`
	Action        string         `json:"action"`
	Params        map[string]any `json:"params"`
}

type createInventoryReq struct {
	Name      string `json:"name"`
	HostsJSON string `json:"hosts_json"`
}

type createPlaybookReq struct {
	Name       string `json:"name"`
	ContentYML string `json:"content_yml"`
	RiskLevel  string `json:"risk_level"`
}
