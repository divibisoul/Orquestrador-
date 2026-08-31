package neuralfabric

import "context"

type Vector []float64
type State struct { Goal string; Features Vector; Constraints map[string]float64 }
type Prediction struct { Value float64; Confidence float64 }
type Route struct { NodeID string; DeviceID string; Precision string; BatchSize int; Score float64; Confidence float64 }
type Experience struct { State State; Action Route; Reward float64; Done bool }
type Encoder interface { EncodeState(context.Context,State)(Vector,error); EncodeTask(context.Context,string)(Vector,error); Normalize(Vector) Vector }
type Predictor interface { Latency(State,Route) Prediction; Cost(State,Route) Prediction; Energy(State,Route) Prediction; Quality(State,Route) Prediction; Failure(State,Route) Prediction }
type Router interface { Route(context.Context,State,[]Route)(Route,error) }
type Learner interface { Observe(Experience); Reward(Experience) float64; Update(context.Context) error; Save(context.Context) error; Load(context.Context) error }
