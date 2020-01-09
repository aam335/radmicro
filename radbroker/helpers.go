package main

// Acct-Status-Type attribute
const (
	Start         = 1
	Stop          = 2
	InterimUpdate = 3
	AccountingOn  = 7
	AccountingOff = 8
	MaxKnownValue = 8
)

// AcctStatusType to topic
func AcctStatusType(u uint32) string {
	switch u {
	case Start:
		return "Start"
	case Stop:
		return "Stop"
	case InterimUpdate:
		return "InterimUpdate"
	case AccountingOff:
		return "AccountingOff"
	case AccountingOn:
		return "AccountingOff"
	}
	return ""
}
