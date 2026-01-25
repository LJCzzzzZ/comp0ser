package worker

type GenScriptPlayLoad struct {
	RawText  string `json:"rawText"`
	Subject  string `json:"subject"`
	Segments int    `json:"segments"`
	MinChars int    `json:"minChars"`
	MaxChars int    `json:"maxChars"`

	// focus for generated narrations
	Focus string `json:"focus"`

	Hook string `json:"hook"`
}

type GenTTSPlayLoad struct {
	FileID string `json:"fileId"`
}

type Narration struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
}
