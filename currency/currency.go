package currency

import (
	"log"
	"math"
	"math/big"

	"github.com/goonma/sdk/currency/decimal"
)

const SYSTEM_DECIMALS int = 12

const BLOCK_CHAIN_DECIMALS int = 18

type SystemCurrency struct {
	Type  Currency
	Name  string
	Value decimal.Decimal
}

func (s SystemCurrency) GetRealValue() *big.Int {
	// convert string to big int
	amountInWei := s.Value.Div(decimal.NewFromInt(int64(math.Pow10(SYSTEM_DECIMALS))))
	return amountInWei.BigInt()

}

func (s SystemCurrency) GetFloat() float64 {
	// convert string to big int
	amountInWei := s.Value.Div(decimal.NewFromInt(int64(math.Pow10(SYSTEM_DECIMALS))))
	// result, ok := amountInWei.Float64()
	// if !ok {
	// 	log.Printf("convert float err = %v", amountInWei)
	// 	return -1
	// }

	return amountInWei.InexactFloat64()

}

func (s SystemCurrency) Validate() bool {
	return true //TODO: Handle validate later //s.Value >= MAX_CURRENCY && s.Value <= MAX_CURRENCY_BLOCK_CHAIN && s.Decimals == SYSTEM_DECIMALS
}

func (s SystemCurrency) Compare(other SystemCurrency) CompareSystemCurrency {
	if s.Type != other.Type {
		return CS_DIFFERENCE
	}

	if s.Value.Cmp(other.Value) < 0 {
		return CS_SMALLER
	}
	if s.Value.Cmp(other.Value) > 0 {
		return CS_BIGGER
	}
	return CS_EQUAL
}

func (s SystemCurrency) Mul(other SystemCurrency) SystemCurrency {
	return SystemCurrency{
		Type:  s.Type,
		Name:  s.Name,
		Value: s.Value.Mul(other.Value),
	}
}

// multiple with float
func (s SystemCurrency) MulWithFloat(value float64) SystemCurrency {
	return SystemCurrency{
		Type:  s.Type,
		Name:  s.Name,
		Value: s.Value.Mul(decimal.NewFromFloat(value)),
	}
}

func (s SystemCurrency) Add(other SystemCurrency) SystemCurrency {
	return SystemCurrency{
		Type:  s.Type,
		Name:  s.Name,
		Value: s.Value.Add(other.Value),
	}
}

func (s SystemCurrency) Sub(other SystemCurrency) SystemCurrency {
	return SystemCurrency{
		Type:  s.Type,
		Name:  s.Name,
		Value: s.Value.Sub(other.Value),
	}
}

func (s SystemCurrency) Div(other SystemCurrency) SystemCurrency {
	return SystemCurrency{
		Type:  s.Type,
		Name:  s.Name,
		Value: s.Value.Div(other.Value),
	}
}

func (s SystemCurrency) DivWithFloat(value float64) SystemCurrency {
	return SystemCurrency{
		Type:  s.Type,
		Name:  s.Name,
		Value: s.Value.Div(decimal.NewFromFloat(value)),
	}
}

func ConvertInt64ToSystemCurrency(realValue int64, typeCurrency Currency) (SystemCurrency, bool) {
	if realValue < 0 {
		log.Printf("Amount in wie error, realValue = %v", realValue)
		return SystemCurrency{}, false
	}

	value := decimal.NewFromInt(realValue)
	systemValue := value
	if typeCurrency > 0 {
		systemValue = value.Mul(decimal.NewFromInt(int64(math.Pow10(SYSTEM_DECIMALS))))
	}

	systemCurrency := SystemCurrency{
		Type:  typeCurrency,
		Name:  CurrencySymbol(typeCurrency),
		Value: systemValue,
	}

	return systemCurrency, true
}

func ConvertFloatToSystemCurrency(realValue float64, typeCurrency Currency) (SystemCurrency, bool) {
	// validate realValue
	if realValue < 0 {
		log.Printf("Amount in wie error, realValue = %v", realValue)
		return SystemCurrency{}, false
	}
	value := decimal.NewFromFloat(realValue)
	systemValue := value
	if typeCurrency > 0 {
		systemValue = value.Mul(decimal.NewFromInt(int64(math.Pow10(SYSTEM_DECIMALS))))
	}

	systemCurrency := SystemCurrency{
		Type:  typeCurrency,
		Name:  CurrencySymbol(typeCurrency),
		Value: systemValue,
	}

	return systemCurrency, true
}

func ConvertFromAmountInWieBlockChain(valueInBlockChain string, typeCurrency Currency) (SystemCurrency, bool) {

	amountInWei, err := decimal.NewFromString(valueInBlockChain)
	if err != nil {
		log.Printf("Amount in wie error, amountInWei = %v", valueInBlockChain)
		return SystemCurrency{}, false
	}

	systemValue := amountInWei.Div(decimal.NewFromInt(int64(math.Pow10(BLOCK_CHAIN_DECIMALS - SYSTEM_DECIMALS))))

	systemCurrency := SystemCurrency{
		Type:  typeCurrency,
		Name:  CurrencySymbol(typeCurrency),
		Value: systemValue,
	}
	return systemCurrency, true
}

func (s SystemCurrency) ConvertValueToBigInt() *big.Int {

	// convert string to big int
	amountInWei := s.Value.Mul(decimal.NewFromInt(int64(math.Pow10(BLOCK_CHAIN_DECIMALS - SYSTEM_DECIMALS))))
	return amountInWei.BigInt()
}

func NewSystemCurrency(value string, typeCurrency Currency) (SystemCurrency, bool) {

	amount, err := decimal.NewFromString(value)
	if err != nil {
		log.Printf("Amount in wie error, amountInWei = %v", value)
		return SystemCurrency{}, false
	}

	systemCurrency := SystemCurrency{
		Type:  typeCurrency,
		Name:  CurrencySymbol(typeCurrency),
		Value: amount,
	}
	return systemCurrency, true

}
