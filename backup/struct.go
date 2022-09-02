package backup

type Status struct {
	Msg     string `json:"msg"`
	Version string `json:"version"`
}

type Result struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type WorkflowResult struct {
	Id       string   `json:"id"`
	Code     int      `json:"code"`
	Messages []string `json:"message"`
	State    string   `json:"state"`
}
