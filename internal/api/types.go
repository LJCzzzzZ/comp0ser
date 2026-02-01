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

type BrunReq struct {
	VideoPath string `json:"videoPath"`
	SubPath   string `json:"subPath"`
	OutPath   string `json:"outPath"`
}

type MergeReq struct {
	VideoPath string `json:"videoPath"`
	AudioPath string `json:"audioPath"`
	OutPath   string `json:"outPath"`
}

type RenderReq struct {
	Folder  string  `json:"folder"`
	Dur     float64 `json:"dur"`     // 目标总时长（秒）
	TailCut float64 `json:"tailCut"` // 每段末尾剪掉秒数（比如 10）
	Loop    bool    `json:"loop"`    // 不够 dur 是否循环补足
	Out     string  `json:"out"`     // 可选：输出文件名
}

type ConcatReq struct {
	Folder string `json:"folder"`
}

type TTSGenAllReq struct {
	Folder string `json:"folder"`
}

type TTSGenSingleReq struct {
	Folder string `json:"folder"`
	NarID  string `json:"narId"`
}

type Narration struct {
	Text string `json:"text"`
}
