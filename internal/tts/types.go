package tts

type SynthesizeResp struct {
	ReqID     string `json:"reqid"`
	Code      int    `json:"code"`
	Message   string
	Operation string `json:"operation"`
	Sequence  int    `json:"sequence"`
	Data      string `json:"data"`
}

type SynthesizeReq struct {
	App     app     `json:"app"`
	User    user    `json:"user"`
	Audio   audio   `json:"audio"`
	Request request `json:"request"`
}

type app struct {
	Cluster string `json:"cluster"`
}

type user struct {
	UID string `json:"uid"`
}

type audio struct {
	VoiceType  string  `json:"voice_type"`
	Encoding   string  `json:"encoding"`
	SpeedRatio float64 `json:"speed_ratio"`
}

type request struct {
	ReqID     string `json:"reqid"`
	Text      string `json:"text"`
	Operation string `json:"operation"`
}
