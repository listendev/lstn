package listentype

type Encoder interface {
	Encode() string
}

type Checker interface {
	Ok() bool
}

type AnalysisRequester interface {
	Encoder
	Checker
}
