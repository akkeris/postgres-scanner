package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"net"
	"os"
	"strconv"
	"time"
)


func main() {
        fmt.Println("Starting run")
        appspacelist,err:=getAppSpaceList()
        if err != nil {
           fmt.Println(err)
        }
fmt.Println(appspacelist)
	for index, element := range appspacelist {
                location, name := getPostgresLocation(element)
		sendPostgresStats(location, name, index)
	}
        fmt.Println("Done with run")
}



func getPostgresLocation(bind string) (l string, n string) {

        brokerdb := os.Getenv("BROKERDB")
        uri := brokerdb
        db, dberr := sql.Open("postgres", uri)
        if dberr != nil {
                fmt.Println(dberr)
        }
        var name string
        var username string
        var password string
        var endpoint string
        dberr = db.QueryRow("select name, username, password, endpoint from databases where id = '"+bind+"'").Scan(&name, &username, &password, &endpoint)

        if dberr != nil {
                db.Close()
                fmt.Println(dberr)
        }
        db.Close()
        return "postgres://"+username+":"+password+"@"+endpoint, name
}


func sendPostgresStats(location string, name string, appspace string) {
        db, err := sql.Open("postgres", location)
        if err != nil {
                fmt.Println(err)
        }
        var size string
        err = db.QueryRow("SELECT pg_database_size('" + name + "') AS size;").Scan(&size)
        if err != nil {
                _ = db.Close()
                fmt.Println(err)
        }

        var connections string
        err = db.QueryRow("SELECT count(*) FROM pg_stat_activity AS connections where datname='"+name+"';").Scan(&connections)
        if err != nil {
                _ = db.Close()
                fmt.Println(err)
        }
        _ = db.Close()


	tsdbconn, err := net.Dial("tcp", os.Getenv("OPENTSDB_IP"))
	if err != nil {
		fmt.Println("dial error:", err)
		return
	}

	metricname := "postgres.db.size"
	value := size
	timestamp := strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10)
	host := name
	put := "put " + metricname + " " + timestamp + " " + value + " dbinstance=" + host
        if ( appspace != ""){

           put = put + " app="+appspace
        }        
        put = put +"\n"
        fmt.Println(put)
	fmt.Fprintf(tsdbconn, put)

	metricname = "postgres.db.connections"
	value = connections
        put = "put " + metricname + " " + timestamp + " " + value + " dbinstance=" + host
        if (appspace != ""){
           put = put + " app="+appspace
        }
        put = put +"\n"
        fmt.Println(put)
	fmt.Fprintf(tsdbconn, put)

	tsdbconn.Close()

}


func getAppSpaceList() (l map[string]string, e error) {
        var list map[string]string
        list = make(map[string]string)


        brokerdb := os.Getenv("PITDB")
        uri := brokerdb
        db, dberr := sql.Open("postgres", uri)
        if dberr != nil {
                fmt.Println(dberr)
                return nil, dberr
        }
        stmt,dberr := db.Prepare("select apps.name as appname, spaces.name as space, service_attachments.service as bindname from apps, spaces, services, service_attachments where services.service=service_attachments.service and owned=true and addon_name='akkeris-postgresql' and services.deleted=false and service_attachments.deleted=false and service_attachments.app=apps.app and spaces.space=apps.space;")
        defer stmt.Close()
        rows, err := stmt.Query()
        if dberr != nil {
                db.Close()
                fmt.Println(dberr)
                return nil, dberr
        }
        defer rows.Close()
        var appname string
        var space string
        var bindname string
        for rows.Next() {
                err := rows.Scan(&appname, &space, &bindname)
                if err != nil {
                        fmt.Println(err)
                        return nil, err
                }
                list[appname+"-"+space]=bindname
        }
        err = rows.Err()
        if err != nil {
                fmt.Println(err)
                return nil, err
        }
        db.Close()
        return list, nil
}
