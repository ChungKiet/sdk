package utils
import (
	"strings"
	"strconv"
	"regexp"
	"fmt"
	"math"
)
func ToLower(s string) string{
	return strings.ToLower(s)
}
func ToUpper(s string) string{
	return strings.ToUpper(s)
}

//Convert String to Integer
func StringToInt(s string)int{
	i,err:=strconv.Atoi(s)
	if err!=nil{
		return math.MinInt32
	}
	return i
}
func StringToInt32(s string)int32{
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
    	return math.MinInt32
	}
	return int32(n)
}
func StringToInt64(s string)int64{
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
    	return math.MinInt64
	}
	return n
}
//substring from left , n character
func Left(s string,n int) string{
	runes := []rune(s)
    // ... Convert back into a string from rune slice.
	left := string(runes[0:n])
	return left
}
//right string
//n length need to slipt
func Right(s string, n int) string{
	if n<=len(s){
		last  := s[len(s)-n:]
		return last
	}
	return s
}
//substring from start to end position
func Mid(st string, s int, e int) string{
	substring:=st
	if e==-1{
		substring = st[s:]
	}else if e<len(st){
		substring = st[s:e]
	}else{
		substring = st[s:len(st)]
	}
	return substring
}
//remove newline
func RemoveNewline(str string)string{
	re := regexp.MustCompile(`\r?\n|\r/g|\t`)
	str = re.ReplaceAllString(str, " ")
	return str
}
//fill N character 
func FillChar(n int,c,str string)string{
	st:=""
	for i:=0;i<n-len(str);i++{
		st+=c
	}
	return fmt.Sprintf("%s%s",st,str)
}
/*func RightS(s string, n int) string {
	
}*/
func Explode(s string, c string) []string{
	arr := strings.Split(s, c)
	return arr
}
func ReverseStringArray(s []string) []string {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}