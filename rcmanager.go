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
	"strings"
	"time"

	"github.com/casbin/casbin"
	"github.com/casbin/xorm-adapter"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo"
	casbinmw "github.com/labstack/echo-contrib/casbinmw"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/middleware"
)

var rcEnforce *casbin.Enforcer

var ormAdapter *OrmAdapter

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
			fmt.Println("Policy import from csv file failed")
			panic(err)
		}
		e.ClearPolicy()
	}
	rcEnforce = casbin.NewEnforcer(rcConfigure.AuthMode, dbAdapter)
	if rcEnforce == nil {
		//TODO add log management later
		fmt.Println("init rc fail")
		return
	}
	rcEnforce.LoadPolicy()
	fmt.Println("Loading Policy successfully")
}

func initORM() {
	ormAdapter = NewAdapter()
}

func startRCServer() {
	e := echo.New()
	e.Debug = true
	//TODO add static profile
	e.Static("/rc", "static")

	//skin the user related function such as user/add ; user/delete
	config := casbinmw.Config{
		Skipper: func(c echo.Context) bool {
			return strings.Contains(c.Path(), "/userpermission") || strings.Contains(c.Path(), "/file")
		},
		Enforcer: rcEnforce,
	}
	e.Use(casbinmw.MiddlewareWithConfig(config))

	//provide API to add,delete and check interface, would not provide the update permission

	e.GET("/userpermission/get", getPermission)
	e.POST("/userpermission/delete", deletePermission)
	e.POST("/userpermission/add", addPermission)
	e.GET("/file", downloadFile)
	//e.POST("/userpermission/update", updatePermission)

	e.Use(session.Middleware(sessions.NewCookieStore([]byte("secret"))))
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	//Reserve for auth validation
	//TODO add auth verification for API
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Extract the credentials from HTTP request header and perform a security check

			// For invalid credentials
			// return echo.NewHTTPError(http.StatusUnauthorized, "Please provide valid credentials")
			// For valid credentials call next
			// return next(c)
			username, _, _ := c.Request().BasicAuth()
			fmt.Println(c.Path() + " recorded" + ", user is " + username)
			return next(c)
		}
	})
	e.Server.ReadTimeout = time.Duration(rcConfigure.ServerReadTimeout) * time.Second
	e.Server.WriteTimeout = time.Duration(rcConfigure.ServerWriteTimeout) * time.Second
	e.Logger.Fatal(e.Start(":1323"))
}

func main() {
	initRCProfiling()
	initCasbinRCEnforce()
	initORM()
	startRCServer()
}
