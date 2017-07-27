// rcprofile.go  - Initial the profile for  kepler resouce management service
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

import "github.com/go-ini/ini"
import "fmt"

//TODO seperation the configuration into sections
type RcConfigure struct {
	Drivername         string `ini:"drivername"`
	DataSourceName     string `ini:"dataSourceName"`
	AuthMode           string `ini:"authMode"`
	AuthFilePolicy     string `ini:"authFilePolicy"`
	AuthImportPolicy   string `ini:"authImportPolicy"`
	OrmShowSQL         bool   `ini:"ormshowsql"`
	OrmMaxIdle         int    `ini:"ormMaxIdle"`
	OrmMaxOpen         int    `ini:"ormmaxOpen"`
	ServerReadTimeout  int    `ini:"serverReadTimeout"`
	ServerWriteTimeout int    `ini:"serverWriteTimeout"`
}

var rcConfigure *RcConfigure
var rcConfigurefile = "./rcmanager.ini"

// load configuration from rcmanager.ini
func initRCProfiling() {
	cfg, err := ini.Load(rcConfigurefile)
	defer ini.Empty()
	if err != nil {
		fmt.Println("Init load profile fail")
		panic(err)
	}
	// set the default value of configuration
	rcConfigure = &RcConfigure{
		Drivername:         "mysql",
		DataSourceName:     "root:Aq1sw2de3@tcp(127.0.0.1:3306)/",
		AuthMode:           "./policy/auth_model.conf",
		AuthFilePolicy:     "./policy/auth_policy.csv",
		AuthImportPolicy:   "./policy/auth_policy.csv",
		OrmShowSQL:         true,
		OrmMaxIdle:         10,
		OrmMaxOpen:         20,
		ServerReadTimeout:  10,
		ServerWriteTimeout: 10,
	}
	/*
		rcConfigure.Drivername = cfg.Section("").Key("drivername").String()
		rcConfigure.DataSourceName = cfg.Section("").Key("dataSourceName").String()
		rcConfigure.AuthMode = cfg.Section("").Key("authMode").String()
		rcConfigure.AuthFilePolicy = cfg.Section("").Key("authFilePolicy").String()
		rcConfigure.AuthImportPolicy = cfg.Section("").Key("authImportPolicy").String()
	*/
	err = cfg.MapTo(rcConfigure)
	if err == nil {
		fmt.Println("driver name : ", rcConfigure.Drivername)
		fmt.Println("dataSourceName : ", rcConfigure.DataSourceName)
		fmt.Println("authMode : ", rcConfigure.AuthMode)
		fmt.Println("authFilePolicy : ", rcConfigure.AuthFilePolicy)
		fmt.Println("authImportPolicy : ", rcConfigure.AuthImportPolicy)
		fmt.Println("OrmShowSQL :", rcConfigure.OrmShowSQL)
		fmt.Println("OrmMaxIdle :", rcConfigure.OrmMaxIdle)
		fmt.Println("OrmMaxOpen :", rcConfigure.OrmMaxOpen)
		fmt.Println("ServerReadTimeout :", rcConfigure.ServerReadTimeout)
		fmt.Println("ServerWriteTimeout :", rcConfigure.ServerWriteTimeout)
	}
}
