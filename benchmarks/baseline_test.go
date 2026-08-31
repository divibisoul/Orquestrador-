package benchmarks

import("testing";"github.com/divibisoul/Orquestrador-/core/superagi")
func BenchmarkInference(b *testing.B){a:=superagi.New();for i:=0;i<b.N;i++{_ = a.Inference("benchmark payload")}}
