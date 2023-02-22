package currency

import (
	"testing"
)

func TestConvertFloatToSystemCurrency(t *testing.T) {

	// init data
	var value float64 = 1.1234567891123
	var typeCurrency Currency = 1

	// call function
	systemCurrency, ok := ConvertFloatToSystemCurrency(value, typeCurrency)
	if !ok {
		t.Errorf("ConvertFloatToSystemCurrency() = %v, want %v", ok, true)
	}

	newSystemCurrency, _ := ConvertFloatToSystemCurrency(1.2342, typeCurrency)

	systemCurrency = systemCurrency.Add(newSystemCurrency)

	// check result
	if systemCurrency.GetFloat() != 2.3576567891 {
		t.Errorf("ConvertFloatToSystemCurrency() = %v, want %v", systemCurrency.GetFloat(), 2.3576567891)
	}
}
