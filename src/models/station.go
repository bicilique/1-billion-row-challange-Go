package models

type Part struct {
	Offset int64
	Size   int64
}

type LineSplit struct {
	Station     []byte
	Temperature []byte
}

type TempStat struct {
	Sum   float32
	Min   float32
	Max   float32
	Count int32
}

type Anomaly struct {
	Station string
	Temp    float32
	Reason  string
}
