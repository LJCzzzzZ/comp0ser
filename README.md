# comp0ser 

An early-stage **AI sleep-aid** media composer: collect raw relaxing assets → let an LLM segment a script into scenes → generate narration with TTS per segment → merge everything into a final sleep video via FFmpeg.

Current focus: **space / universe-themed** sleep content.



## Quick Start

### Requirements

- Go (recent stable)
- FFmpeg available in your `PATH`

### Run

```
go run ./cmd/main.go
```



## TODO

- [ ] Move From single-image video to multi-asset video
- [ ] Add subtitle generation (SRT) and optional burn-in
- [ ] Improve audio quality: reduce artifacts/noise and keep a consistent voice timbre across segments
- [ ] Better audio-visual alignment: match visuals to the corresponding narration segments




