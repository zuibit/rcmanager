// rcmanager.go  - Main file for kepler resouce management
//
// Copyright (c) 2017-2019 - Zou Wei <weizou@cogobuy.com>
//
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
// 1. Redistributions of source code must retain the above copyright notice, this
//    list of conditions and the following disclaimer.
// 2. Redistributions in binary form must reproduce the above copyright notice,
//    this list of conditions and the following disclaimer in the documentation
//    and/or other materials provided with the distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
// ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
// WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT OWNER OR CONTRIBUTORS BE LIABLE FOR
// ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
// (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
// LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
// ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
// SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/casbin/casbin"
	"github.com/casbin/xorm-adapter"
	//	"github.com/dgrijalva/jwt-go"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo"
	casbinmw "github.com/labstack/echo-contrib/casbinmw"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/middleware"
	log "github.com/sirupsen/logrus"
)

var rcEnforce *casbin.Enforcer

var ormPermissionAdapter *OrmPermissionAdapter
var ormMetadataAdapter *OrmMetadataAdapter
var rcLog = log.New()

func initLog() {
	// Log as JSON instead of the default ASCII formatter.
	rcLog.Formatter = &log.JSONFormatter{}
	//TODO seperate the log file
	if rcConfigure.LogFile != "" {
		timestamp := time.Now().Format("2006-01-02-15-04-05-")

		file, err := os.OpenFile(timestamp+rcConfigure.LogFile, os.O_CREATE|os.O_WRONLY, 0666)
		if err == nil {
			rcLog.Out = file
		} else {
			rcLog.Out = os.Stdout
		}
	}

	// Log leve is debug, info, warn, error, fatal, panic
	// TODO read the level from profile
	switch rcConfigure.LogLevel {
	case 0:
		rcLog.SetLevel(log.DebugLevel)
	case 1:
		rcLog.SetLevel(log.InfoLevel)
	case 2:
		rcLog.SetLevel(log.WarnLevel)
	case 3:
		rcLog.SetLevel(log.ErrorLevel)
	case 4:
		rcLog.SetLevel(log.FatalLevel)
	case 5:
		rcLog.SetLevel(log.PanicLevel)
	default:
		rcLog.SetLevel(log.ErrorLevel)
	}
	fmt.Println("Set Log level: ", log.GetLevel())

	//TODO add the hook to collect log in Mongo, Redis or other...
}

// just support mysql adaptor and file adaptor
// init ORM, TODO move the create table implement into rcorm
func initCasbinRCEnforce() {
	// check whether DB is created
	// if not ,first time, user could load the policy from csv file ,then init the database
	dbAdapter := xormadapter.NewAdapter(rcConfigure.Drivername, rcConfigure.DataSourceName)
	isTableEmpty, _ := dbAdapter.IsTableEmpty()
	// if table line is nil, we need to check user profile whether import from csv file
	if isTableEmpty == true && len(rcConfigure.AuthImportPolicy) != 0 {
		e := casbin.NewEnforcer(rcConfigure.AuthMode, rcConfigure.AuthImportPolicy)
		//TODO maybe it would be some performance issue if import big data, need to check later
		err := dbAdapter.SavePolicy(e.GetModel())
		if err != nil {
			rcLog.WithFields(log.Fields{
				"Import Policay": rcConfigure.AuthImportPolicy,
			}).Panic("Policy import from csv file failed")

			panic(err)
		}
		e.ClearPolicy()
	}
	rcEnforce = casbin.NewEnforcer(rcConfigure.AuthMode, dbAdapter)
	if rcEnforce == nil {
		//TODO add log management later
		rcLog.Error("Init casbin Enforcer failed")
		return
	}
	rcEnforce.LoadPolicy()
	rcLog.Info("Loading Policy successfully")
}

func initORM() {
	ormPermissionAdapter = NewPermissionAdapter()
	ormMetadataAdapter = NewMetadataAdapter()
}

func startRCServer() {
	e := echo.New()
	e.Debug = true
	e.Static("/rc", rcConfigure.StaticFolder)

	//skip the open API related function such as userpermission/add ; userpermission/delete
	config := casbinmw.Config{
		Skipper: func(c echo.Context) bool {
			return strings.Contains(c.Path(), "/userpermission") || strings.Contains(c.Path(), "/thing")
		},
		Enforcer: rcEnforce,
	}
	e.Use(casbinmw.MiddlewareWithConfig(config))

	//provide API to add,delete and check interface, would not provide the update permission

	e.GET("/userpermission/get", getPermission)
	e.POST("/userpermission/delete", deletePermission)
	e.POST("/userpermission/add", addPermission)
	e.GET("/thing/download", downloadThing)
	e.POST("/thing/upload", upLoadThing)
	e.POST("/thing/delete", deleteThing)
	e.GET("/thing/get", getThings)
	//e.POST("/userpermission/update", updatePermission)

	e.Use(session.Middleware(sessions.NewCookieStore([]byte("secret"))))
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	//TODO switch to https after enable
	//e.Pre(middleware.HTTPSWWWRedirect())

	//Reserve for auth validation
	//TODO add auth verification for API
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// TODO Suggest to use JWT token
			// Extract the credentials from HTTP request header and perform a security check

			// For invalid credentials
			// return echo.NewHTTPError(http.StatusUnauthorized, "Please provide valid credentials")
			// For valid credentials call next
			// return next(c)
			//user := c.Get("user").(*jwt.Token)
			//claims := user.Claims.(jwt.MapClaims)
			//name := claims["name"].(string)
			// check whether user is valid
			return next(c)
		}
	})
	e.Server.ReadTimeout = time.Duration(rcConfigure.ServerReadTimeout) * time.Second
	e.Server.WriteTimeout = time.Duration(rcConfigure.ServerWriteTimeout) * time.Second
	//TODO JWT security key
	//e.Use(middleware.JWT([]byte("secret")))
	e.Logger.Fatal(e.Start(":1323"))
	//TODO Enabling https shall register domain for the service
	//e.Logger.Fatal(e.StartAutoTLS(":443"))
}

func main() {

	initRCProfiling()
	initLog()
	initCasbinRCEnforce()
	initORM()
	startRCServer()
}
