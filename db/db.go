package db
import (
	dbcf "github.com/goonma/sdk/db/mongo/config"
	"github.com/goonma/sdk/utils"
	e "github.com/goonma/sdk/base/error"
	"github.com/goonma/sdk/encrypt"
	"github.com/goonma/sdk/log"
	"github.com/goonma/sdk/config/vault"
	"github.com/goonma/sdk/db/mongo"
	mgoDriver "go.mongodb.org/mongo-driver/mongo"
	"fmt"
	//"errors"
	"strings"
)
type MongoDB struct{
	Cols map[string]*mongo.Collection
	Conn map[string]*mongo.Client
	Config *vault.Vault
	map_conn_cols map[string][]string
}
/*
collections: list struct of collections need to be create collections[collection_name]=collection struct schema
*/
func (mgo *MongoDB)Initial(config *vault.Vault,collections map[string]interface{}) *e.Error{
	log.Info("Initialing Mongo DB...","MONGO_DB")
	if mgo.Config==nil{
		mgo.Config=config
	}
	if len(mgo.Cols)==0{
		mgo.Cols=make(map[string]*mongo.Collection)
		mgo.map_conn_cols=make(map[string][]string)
	}
	if len(mgo.Conn)==0{
		mgo.Conn=make(map[string]*mongo.Client)
	}
	if collections==nil{
		return e.New("Collection schema is empty","MONGO_DB","INIT_CONNECTION")
	}
	service_name:=mgo.Config.GetServiceName()
	service_config_path:=strings.ReplaceAll(service_name,".","/")
	//golbal config
	global_db_config_path:=fmt.Sprintf("%s/%s",service_config_path,"db/general")
	global_config_map:=dbcf.GetConfig(mgo.Config,global_db_config_path)
	//db config
	db_config_path:=fmt.Sprintf("%s/%s",service_config_path,"db")
	//get all collection
	cols:=mgo.Config.ListItemByPath(db_config_path)
	//
	//map_conn_cols:=make(map[string][]string)
	//init DB connection 
	var config_map map[string]string
	if len(cols)>0{
		//create all db connections
		for _,col:=range cols{
			if col!="general"{
				db_path:=fmt.Sprintf("%s/%s",db_config_path,col)
				local_config_map:=dbcf.GetConfig(mgo.Config,db_path)
				config_map=dbcf.MergeConfig(global_config_map,local_config_map)
					if utils.MapI_contains(collections,col){
					err:=mgo.initialDBConn(config_map,collections[col],col)
					if err!=nil{
						return err
					}
				}
			}
		}
		//bind collection to db connection
		for hash,_:=range mgo.Conn{
			list_cols:="Mongo DB collections: "
			mgo.Conn[hash].OnConnected = func(database *mgoDriver.Database) error {
				for _,col:=range mgo.map_conn_cols[hash]{
					mgo.Cols[col].ApplyDatabase(database)
					list_cols=fmt.Sprintf("%s, %s",list_cols,col)
				}
				return nil
			}
			err := mgo.Conn[hash].Connect()
			if err != nil {
				return e.New(err.Error(),"MONGO_DB","INIT_CONNECTION")
			}
			list_cols=fmt.Sprintf("HOST:%s %s: %s",config_map["HOST"],list_cols," connected")
			log.Info(list_cols,"MONGO_DB","INIT_CONNECTION")
		}
	}
	return nil
}

func (mgo *MongoDB)InitialByConfigPath(config *vault.Vault,path,collection_name string, collection_struct interface{}) *e.Error{
	if mgo.Config==nil{
		mgo.Config=config
	}
	if len(mgo.Cols)==0{
		mgo.Cols=make(map[string]*mongo.Collection)
		mgo.map_conn_cols=make(map[string][]string)
	}
	if len(mgo.Conn)==0{
		mgo.Conn=make(map[string]*mongo.Client)
	}
	service_name:=mgo.Config.GetServiceName()
	service_config_path:=strings.ReplaceAll(service_name,".","/")
	db_config_path:=fmt.Sprintf("%s/%s",service_config_path,path)
	config_map:=dbcf.GetConfig(mgo.Config,db_config_path)
	//init db connection
	err:=mgo.initialDBConn(config_map,collection_struct,collection_name)
	if err!=nil{
		return err
	}
	//applly collection
	db_info_str:=fmt.Sprintf("%s%s%s",config_map["HOST"],config_map["DB"],config_map["COLLECTION"])
	hash:=encrypt.HashMD5(db_info_str)
	mgo.Conn[hash].OnConnected = func(database *mgoDriver.Database) error {
		mgo.Cols[collection_name].ApplyDatabase(database)
		return nil
	}
	return nil
}
/*

*/
func (mgo *MongoDB)initialDBConn(config_map map[string]string, collection interface{}, col string) *e.Error{
	//check config exists
	if !utils.Map_contains(config_map,"HOST") && config_map["HOST"]!=""{
		return e.New("Host not found","MONGO_DB","INIT_CONNECTION")
	}
	if !utils.Map_contains(config_map,"USERNAME") && config_map["USERNAME"]!=""{
		return e.New("Username not found","MONGO_DB","INIT_CONNECTION")
	}
	if !utils.Map_contains(config_map,"PASSWORD") && config_map["PASSWORD"]!=""{
		return e.New("Password not found","MONGO_DB","INIT_CONNECTION")
	}
	if !utils.Map_contains(config_map,"DB") && config_map["DB"]!=""{
		return e.New("DB not found","MONGO_DB","INIT_CONNECTION")
	}
	if !utils.Map_contains(config_map,"AUTHDB") && config_map["AUTHDB"]!=""{
		return e.New("AuthDB not found","MONGO_DB","INIT_CONNECTION")
	}
	db_info_str:=fmt.Sprintf("%s%s%s",config_map["HOST"],config_map["DB"],config_map["COLLECTION"])
	hash_db_info:=encrypt.HashMD5(db_info_str)
	//create collection instance
	mgo.Cols[col]=&mongo.Collection{
		ColName 	  :col,
		TemplateObject: collection,
	}
	//create connections
	if !mongo.MapCusorI_contains(mgo.Conn,hash_db_info){
		db_cfg:=mongo.MapToDBConfig(config_map)
		mgo.Conn[hash_db_info]=mongo.NewClient(db_cfg)
	}
	mgo.map_conn_cols[hash_db_info]=append(mgo.map_conn_cols[hash_db_info],col)
	return nil
}
func (mgo *MongoDB)Clean(){
	if mgo.Conn!=nil{
		for _,c:=range mgo.Conn{
			c.Disconnect()
		}
	}
}
