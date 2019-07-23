package main

const (
	//StopSig stop signal
	StopSig = 1
	//ErrorSig error signal
	ErrorSig = 2
	//DoneSig done sysnal
	DoneSig = 3
)

//Signal status of worker
type Signal struct {
	Sig int
	Err error
}

// Worker worker for clawers
type Worker interface {
	RunJob(job <-chan string, quit <-chan int, reportSignal chan<- Signal)
}