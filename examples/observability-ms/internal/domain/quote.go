package domain // доменная модель котировки (FinTech-пример)

// Quote represents a simple FinTech quote (domain concept).
type Quote struct {
	Symbol string
	Price  float64
}
