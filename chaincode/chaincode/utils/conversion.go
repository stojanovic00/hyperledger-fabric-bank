package utils

func EurToDin(eurAmount float64) float64 {
	exchangeRate := 117.0
	dinAmount := eurAmount * exchangeRate
	return dinAmount
}

func DinToEur(dinAmount float64) float64 {
	exchangeRate := 117.0
	eurAmount := dinAmount / exchangeRate
	return eurAmount
}
