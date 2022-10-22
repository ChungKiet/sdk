package datetime 
import (
	"time"
	"github.com/goonma/sdk/utils"
)
//agrs[0] format
func DateFromTimeStamp(timestamp int64, args...string) string{
	tm := time.Unix(timestamp, 0)
	year_str := utils.IntToS(tm.Year())
	month_str := utils.FillChar(2,"0",utils.IntToS(int(tm.Month())))
	day_str := utils.FillChar(2,"0",utils.IntToS(tm.Day()))
	h_str := utils.FillChar(2,"0",utils.IntToS(tm.Hour()))
	min_str := utils.FillChar(2,"0",utils.IntToS(tm.Minute()))
    s_str := utils.FillChar(2,"0",utils.IntToS(tm.Second()))
    //n_str := utils.IntToS(tm.Nanosecond())
	datetime_str:=""
	if len(args)==0{//mm/dd/yyyy
		datetime_str = month_str + "/" + day_str + "/" + year_str + " " + h_str+":"+min_str+":"+s_str
	}else if args[0]=="dd/mm/yyyy"{
		datetime_str =  day_str + "/" + month_str +"/"+ year_str +" " + h_str+":"+min_str+":"+s_str
	}else if args[0]=="yyyy-mm-dd"{
		datetime_str = year_str +"-"+ month_str + "-" + day_str +  " " + h_str+":"+min_str+":"+s_str
	}else if args[0]=="yyyy/mm/dd"{
		datetime_str = year_str +"/"+ month_str + "/" + day_str +  " " + h_str+":"+min_str+":"+s_str
	}
	return datetime_str
}
//timestamp milisecond
func TimeStampZeroHMS(timestamp int64) (int64,error){
	tm := time.Unix(timestamp, 0)
	//
	layout := "01/02/2006 15:04:05"
	//
	year_str := utils.IntToS(tm.Year())
	month_str := utils.FillChar(2,"0",utils.IntToS(int(tm.Month())))
	day_str := utils.FillChar(2,"0",utils.IntToS(tm.Day()))
	datetime_str := month_str + "/" + day_str + "/" + year_str + " " + "00:00:00"
    t, err := time.Parse(layout, datetime_str)
    if err != nil {
        return -1,err 
    }
	return t.Unix(),nil
}

//date to timestamp
//default datetime has format: 2006-01-02T15:04:05.000Z (mongodb default)
func DateTimeToTimestamp(datetime string, args...string)(int64,error){
	layout := "2006-01-02T15:04:05.000Z"
	if len(args)>0{
		layout=args[0]
	}
    t, err := time.Parse(layout, datetime)
    if err != nil {
        return -1,err 
    }
    return t.Unix(),nil
}
