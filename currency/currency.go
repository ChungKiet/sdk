package currency

import (
	"log"
	"math"
	"math/big"
)

const SYSTEM_DECIMALS int = 8
const BLOCK_CHAIN_DECIMALS int = 18

type SystemCurrency struct {
	Type     Currency `json:"type" bson:"type"`
	Name     string   `json:"name,omitempty"  bson:"name,omitempty"`
	Value    int64    `json:"value" bson:"value"`
	Decimals int      `json:"decimals" bson:"decimals"`
}

type ExchangeRes struct {
	Success bool    `json:"success"`
	Data    float64 `json:"data"`
}

func (s SystemCurrency) GetRealValue() float64 {
	return float64(s.Value) / math.Pow10(SYSTEM_DECIMALS)
}

func (s SystemCurrency) Validate() bool {
	return true //TODO: Handle validate later //s.Value >= MAX_CURRENCY && s.Value <= MAX_CURRENCY_BLOCK_CHAIN && s.Decimals == SYSTEM_DECIMALS
}

func (s SystemCurrency) Compare(other SystemCurrency) CompareSystemCurrency {
	if s.Type != other.Type || s.Decimals != other.Decimals {
		return CS_DIFFERENCE
	}
	if s.Value < other.Value {
		return CS_SMALLER
	}
	if s.Value > other.Value {
		return CS_BIGGER
	}
	return CS_EQUAL
}

func (s SystemCurrency) Multiple(mul int64) SystemCurrency {
	return SystemCurrency{
		Type:     s.Type,
		Name:     s.Name,
		Value:    s.Value * mul,
		Decimals: s.Decimals,
	}
}

func (s SystemCurrency) MultipleFloat(mul float64) SystemCurrency {
	return SystemCurrency{
		Type:     s.Type,
		Name:     s.Name,
		Value:    int64(float64(s.Value) * mul), //TODO: Anh Lam check lai giup em
		Decimals: s.Decimals,
	}
}

func ConvertIntToSystemCurrency(realValue int, typeCurrency Currency) SystemCurrency {
	//TODO: validate realValue
	systemValue := int64(math.Pow10(SYSTEM_DECIMALS)) * int64(realValue)
	return convertSystemCurrency(systemValue, typeCurrency)
}

// khi lay price da duoc convert roi
func ConvertInt64ToSystemCurrencyWithSystemValue(realValue int64, typeCurrency Currency) SystemCurrency {
	//TODO: validate realValue
	return convertSystemCurrency(realValue, typeCurrency)
}

func ConvertInt64ToSystemCurrency(realValue int64, typeCurrency Currency) SystemCurrency {
	//TODO: validate realValue
	systemValue := int64(math.Pow10(SYSTEM_DECIMALS)) * realValue
	return convertSystemCurrency(systemValue, typeCurrency)
}

func ConvertFloatToSystemCurrency(realValue float64, typeCurrency Currency) SystemCurrency {
	//TODO: validate realValue
	systemValue := int64(math.Round(math.Pow10(SYSTEM_DECIMALS) * realValue))
	return convertSystemCurrency(systemValue, typeCurrency)
}

func ConvertFromAmountInWieBlockChain(valueInBlockChain string, typeCurrency Currency) (*SystemCurrency, bool) {

	amountInWei := new(big.Int)
	amountInWei, ok := amountInWei.SetString(valueInBlockChain, 10)
	if !ok {
		log.Printf("Amount in wie error, amountInWei = %v", valueInBlockChain)
		return nil, false
	}	
	systemValue := new(big.Int).Div(amountInWei, big.NewInt(int64(math.Pow10(BLOCK_CHAIN_DECIMALS-SYSTEM_DECIMALS))))
	systemCurrency := &SystemCurrency{
		Type:     typeCurrency,
		Name:     CurrencySymbol(typeCurrency),
		Value:    systemValue.Int64(),
		Decimals: SYSTEM_DECIMALS,
	}
	return systemCurrency, true
}

func (s SystemCurrency) ConvertValueToWieBlockChain() *big.Int {
	systemValue := big.NewInt(s.Value)
	return new(big.Int).Mul(systemValue, big.NewInt(int64(math.Pow10(BLOCK_CHAIN_DECIMALS-SYSTEM_DECIMALS))))
}

func convertSystemCurrency(systemValue int64, typeCurrency Currency) SystemCurrency {
	systemCurrency := SystemCurrency{
		Type:     typeCurrency,
		Name:     CurrencySymbol(typeCurrency),
		Value:    systemValue,
		Decimals: SYSTEM_DECIMALS,
	}
	return systemCurrency
}

func NewSystemCurrency(systemValue int64, typeCurrency Currency) SystemCurrency {
	return convertSystemCurrency(systemValue, typeCurrency)
}
