// rcorm.go  - ORM management of kepler resouce management database
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
	_ "errors"
	"fmt"
	"runtime"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	//"github.com/go-xorm/xorm-redis-cache"
)

type Line struct {
	PType    string `xorm:"varchar(100)  PK 'p_type'" json:"ptype"` // PType could be user or group
	Username string `xorm:"varchar(100)  PK 'V0'" json:"username"`  // username
	Resource string `xorm:"varchar(100)  PK 'V1'" json:"resource"`  // resource could be access
	Action   string `xorm:"varchar(100)  PK 'V2'" json:"action"`    // action such as get/delete/put/ *
	V3       string `xorm:"varchar(100) 'V3'" json:"-"`             // reserved but not used
	V4       string `xorm:"varchar(100) 'V4'" json:"-"`             // reserved but not used
	V5       string `xorm:"varchar(100) 'V5'" json:"-"`             // reserved but not used
}

type Lines struct {
	Lines []Line `json:"result"`
}

type OrmAdapter struct {
	driverName     string
	dataSourceName string
	engine         *xorm.Engine
}

// finalizer is the destructor for Adapter.
func finalizer(a *OrmAdapter) {
	a.engine.Close()
	fmt.Println("Casbin ORM close")
}

// NewAdapter is the constructor for Adapter.
func NewAdapter() *OrmAdapter {
	//TODO Inital the driver and source from profile
	a := &OrmAdapter{}
	a.driverName = rcConfigure.Drivername
	a.dataSourceName = rcConfigure.DataSourceName

	// Open the DB, create it if not existed.
	a.open()

	// Call the destructor when the object is released.
	runtime.SetFinalizer(a, finalizer)
	return a
}

// Open the orm engine
func (a *OrmAdapter) open() {
	//Assume that DB is already exist , not need to create
	var engine *xorm.Engine
	var err error
	if a.driverName == rcConfigure.Drivername {
		engine, err = xorm.NewEngine(a.driverName, a.dataSourceName+"casbin")
	}
	if err != nil {
		panic(err)
	}
	fmt.Println("ORM casbin engine create successful")

	a.engine = engine
	a.engine.ShowSQL(rcConfigure.OrmShowSQL)
	//set the redis
	//a.engine.SetDefaultCacher(xormrediscache.NewRedisCacher("192.168.152.161:6379", "", xormrediscache.DEFAULT_EXPIRATION, nil))
	cacher := xorm.NewLRUCacher(xorm.NewMemoryStore(), 1000)
	a.engine.SetDefaultCacher(cacher)
	a.engine.SetMaxIdleConns(rcConfigure.OrmMaxIdle)
	a.engine.SetMaxOpenConns(rcConfigure.OrmMaxOpen)

	a.SyncTable()
}

// Sync the table structure
func (a *OrmAdapter) SyncTable() {
	err := a.engine.Sync2(new(Line))
	if err != nil {
		panic(err)
	}
	fmt.Println("ORM sync casbin table successful")
}

// The query rule shall be username + resouce，find more record
//TODO add the validation from route level
func (a *OrmAdapter) findUserPermission(lines *[]Line, line *Line) error {
	//result := make([]Line, 0)

	//return a.engine.Find(&result, line)
	err := a.engine.Find(lines, line)
	fmt.Println("ORM find the result account :", len(*lines))
	return err
}

// The query rule shall be username + resouce, get one record
//TODO add the validation from route level
func (a *OrmAdapter) getUserPermission(line *Line) (bool, error) {
	return a.engine.Get(*line)
}

func (a *OrmAdapter) deleteUserPermission(line Line) (int64, error) {
	return a.engine.Delete(line)
}

//TODO No need to update the permission, for one kind of resource , only provide add, delete and get
//would remove it
func (a *OrmAdapter) updateUserPermission(line *Line) (int64, error) {
	mLine := &Line{}
	has, err := a.engine.Where("username=?", mLine.Username).Get(mLine)

	if err != nil {
		fmt.Println("Exception, ORM could not find the record properly")
		return -1, err
	} else if has != true {
		fmt.Println("ORM could not find the record properly")
		return -1, err
	} else {
		return a.engine.Update(line)
	}
}

func (a *OrmAdapter) addUserPermission(line Line) (int64, error) {
	//TODO, consider whether to use transaction
	session := a.engine.NewSession()
	defer session.Close()
	err := session.Begin()
	has, err := session.Get(&line)
	if err != nil {
		session.Rollback()
		return -1, err
	}
	if has == true {
		// TODO return new error
		fmt.Println("Record already exist")
		return -1, ErrNotExist
	}
	id, err := session.Insert(&line)
	if err != nil {
		fmt.Println("ORM Insert Data fail")
		session.Rollback()
		return -1, err
	}
	err = session.Commit()
	if err != nil {
		return -1, err
	}
	return id, nil
}
