package mesh

import "testing"

func TestRuntimeOSSAffinityIncludesFederatedProviders(t *testing.T) {
  want := map[string]string{
    "superagi":"N07",
    "langgraph":"N07",
    "crewai":"N07",
    "microsoft-agent-framework":"N07",
    "openhands":"N06",
    "metagpt":"N06",
    "agentscope":"N03",
    "letta-code":"N01",
    "browser-use":"N05",
    "smolagents":"N06",
    "pydantic-ai":"N01",
    "llama-index":"N05",
    "dspy":"N06",
    "whisper":"N03",
    "kokoro":"N03",
  }
  for capability, owner := range want {
    targets := runtimeOSSAffinityTargets(capability)
    if len(targets) == 0 || targets[0] != owner {
      t.Fatalf("%s targets=%v want primary=%s", capability, targets, owner)
    }
  }
}
