package currency

// Currency
type Currency int

const (
	GM Currency = 1

	BUSD Currency = 31
	USDT Currency = 32
	USDC Currency = 33
)

var currencySymbol = map[Currency]string{
	BUSD: "BUSD",
	USDT: "USDT",
	USDC: "USDC",
	GM:   "GM",
}

func CurrencySymbol(code Currency) string {
	return currencySymbol[code]
}

type CompareSystemCurrency int

const (
	CS_SMALLER    CompareSystemCurrency = -1
	CS_EQUAL      CompareSystemCurrency = 0
	CS_BIGGER     CompareSystemCurrency = 1
	CS_DIFFERENCE CompareSystemCurrency = 99
)
