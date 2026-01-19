package llm

import (
	"context"
	"os"
	"testing"
)

func TestClient(t *testing.T) {
	client := NewClient("8a6f5c77-20da-4228-a409-a01f467829c2")

	b, resp, err := client.Synthesize(context.Background(), "volcano_icl", "comp0ser", "S_L7R26kdR1", "黑洞的存在可以透过它与其它物质和电磁辐射（如可见光）的相互作用推断出来。落在黑洞上的物质会因为摩擦加热而在黑洞的两极产生明亮的X射线喷流。吸积物质在落入黑洞前围绕黑洞以接近光速的速度旋转，并形成包裹黑洞的扁平吸积盘，成为宇宙中最亮的一些天体。如果有其它恒星围绕着黑洞运行，它们的轨道可以用来确定黑洞的质量和位置。这种观测可以排除其它可能的天体，例如中子星。经由这种方法，天文学家在许多联星系统确认了黑洞候选者，并确定银河系核心被称为人马座A*的电波源包含一个超大质量黑洞，其质量大约是430万太阳质量。")
	if err != nil {
		t.Fatalf("tts failed: %v, resp=%+v", err, resp)
	}

	if err := os.WriteFile("test.mp3", b, 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	t.Log("saved: test.mp3")
}
