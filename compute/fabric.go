package compute

import "context"

type DeviceKind string
const ( CPU DeviceKind = "cpu"; GPU DeviceKind = "gpu"; NPU DeviceKind = "npu" )
type Precision string
const ( FP32 Precision = "fp32"; FP16 Precision = "fp16"; BF16 Precision = "bf16"; INT8 Precision = "int8"; INT4 Precision = "int4" )
type Device struct { ID string; Kind DeviceKind; FLOPs float64; MemoryMB int64; VRAMMB int64; TemperatureC float64; PowerWatts float64; Utilization float64; Precisions []Precision; Ready bool }
type Job struct { ID string; Model string; Tokens int; Precision Precision; DeadlineMs float64; MaxEnergyJ float64; QualityTarget float64 }
type Scheduler interface { Schedule(context.Context, Job, []Device) (Device,error); Batch([]Job) [][]Job; MicroBatch([]Job,int) [][]Job; Quantize(Job,Precision) Job; Migrate(context.Context,Job,Device) error }
type Fabric interface { Devices(context.Context) ([]Device,error); Execute(context.Context,Job,Device) (Result,error); Health(context.Context) error }
type Result struct { JobID string; DeviceID string; LatencyMs float64; EnergyJ float64; Quality float64; FLOPs float64 }
