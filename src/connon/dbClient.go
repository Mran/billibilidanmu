/*
数据库的统一管理
*/
package connon

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"time"
)

var (
	DbClient          *mongo.Client
	DbContext         mongo.SessionContext
	VideoCollection  *mongo.Collection
	DanmuCollection    *mongo.Collection
	UpperCollection    *mongo.Collection

)

const dbinfo = "mongodb://localhost:27017/"

//初始化数据库的连接
func init() {
	var clientErr error
	DbClient, clientErr = mongo.NewClient(options.Client().ApplyURI(dbinfo))
	if clientErr != nil {
		panic("客户端未建立未连接")
	}
	checkErr(clientErr)
	DbContext, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	clientErr = DbClient.Connect(DbContext)
	if clientErr != nil {
		panic("服务器未连接")
	}
	checkErr(clientErr)
	log.Println("数据库已连接")
	VideoCollection = DbClient.Database("bili").Collection("videoList")
	DanmuCollection = DbClient.Database("bili").Collection("danmu")
	UpperCollection = DbClient.Database("bili").Collection("upper")


}
func GetContext() (ctx context.Context) {
	ctx, _ = context.WithTimeout(context.Background(), 10*time.Second)
	return
}
func checkErr(err error) {
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Panicln("没有查到数据")
		} else {
			fmt.Println(err)
			os.Exit(0)
		}

	}
}
