package trigger

type Trigger interface {
	WaitForTrigger() <-chan TriggeredBy
	Start() error
	Stop() error
}

type TriggeredBy string
