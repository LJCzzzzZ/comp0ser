package api

type GenScriptReq struct {
	RawText  string `json:"rawText"`
	Subject  string `json:"subject"`
	Segments int    `json:"segments"`
	MinChars int    `json:"minChars"`
	MaxChars int    `json:"maxChars"`

	// focus for generated narrations
	Focus string `json:"focus"`

	Hook string `json:"hook"`
}

type TTSGenAllReq struct {
	FileID string `json:"fileId"`
}

type Narration struct {
	Text string `json:"text"`
}
