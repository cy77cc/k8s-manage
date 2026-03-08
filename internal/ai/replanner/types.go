package replanner

type ReplanDecision struct {
	NeedReplan  bool     `json:"need_replan"`
	Reason      string   `json:"reason,omitempty"`
	Suggestions []string `json:"suggestions,omitempty"`
}
