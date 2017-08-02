// rcmetadataorm.go  - Metadata ORM file  for  kepler resouce management service
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
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	//"github.com/go-xorm/xorm-redis-cache"
)

type Property struct {
	ClassName      string   `json:"class_name"`
	Name           string   `json:"name"`            // name of things
	Vendor         string   `json:"vendor"`          // Vendor of things
	ReleaseTime    string   `json:"release_time"`    //TODO change to time format
	ClassId        string   `json:"class_id"`        // reserved but not used
	TargetPlatform string   `json:"target_platform"` // reserved but not used
	Description    string   `json:"description"`     // reserved but not used
	DependClass    []string `json:"dependclass"`
}

// TODO add version
type Metadata struct {
	ID          int64      `xorm:autoincr`
	ClassID     string     `xorm:varchar(100) "unique"`
	PackageName string     `xorm:"varchar(100) PK index 'packageName'" json:"package_name"`
	Properties  []Property `xorm:"Text json 'properties'" json:"metadata"`
	FilePath    string     `xorm:varchar(100)`
	Version     float32    `xorm:Float`
	CreateAt    time.Time  `xorm:"created"`
	UpdateAt    time.Time  `xorm:"updated"`
	DeleteAt    time.Time  `xorm:"deleted"`
}

type OrmMetadataAdapter struct {
	driverName     string
	dataSourceName string
	engine         *xorm.Engine
}

// finalizer is the destructor for Adapter.
func metadatafinalizer(a *OrmMetadataAdapter) {
	a.engine.Close()
	fmt.Println("Metadata ORM close")
}

// NewAdapter is the constructor for Adapter.
func NewMetadataAdapter() *OrmMetadataAdapter {
	a := &OrmMetadataAdapter{}
	a.driverName = rcConfigure.Drivername
	a.dataSourceName = rcConfigure.DataSourceName

	// Open the DB, create it if not existed.
	a.open()

	// Call the destructor when the object is released.
	runtime.SetFinalizer(a, metadatafinalizer)
	return a
}

//Create mySql database
func (a *OrmMetadataAdapter) createDatabase() error {

	engine, err := xorm.NewEngine(a.driverName, a.dataSourceName)
	if err != nil {
		return err
	}
	defer engine.Close()

	_, err = engine.Exec("CREATE DATABASE IF NOT EXISTS things")

	return err
}

// Open the orm engine
func (a *OrmMetadataAdapter) open() {
	var engine *xorm.Engine
	var err error

	if err = a.createDatabase(); err != nil {
		panic(err)
	}

	if a.driverName == rcConfigure.Drivername {
		engine, err = xorm.NewEngine(a.driverName, a.dataSourceName+"things")
	}
	if err != nil {
		panic(err)
	}
	fmt.Println("ORM things engine create successful")

	a.engine = engine
	a.engine.ShowSQL(rcConfigure.OrmShowSQL)
	cacher := xorm.NewLRUCacher(xorm.NewMemoryStore(), 1000)
	a.engine.SetDefaultCacher(cacher)
	a.engine.SetMaxIdleConns(rcConfigure.OrmMaxIdle)
	a.engine.SetMaxOpenConns(rcConfigure.OrmMaxOpen)

	a.SyncTable()
}

// Sync the table structure
func (a *OrmMetadataAdapter) SyncTable() {
	err := a.engine.Sync2(new(Metadata))
	if err != nil {
		panic(err)
	}
	fmt.Println("ORM sync thing table successful")
}

func (a *OrmMetadataAdapter) addthing(metadata Metadata) (int64, error) {
	has, err := a.engine.Get(&metadata)
	if err != nil {
		return -1, err
	}
	if has == true {
		fmt.Println("Record already exist")
		return -1, ErrNotExist
	}
	id, err := a.engine.Insert(&metadata)
	if err != nil {
		fmt.Println("ORM Insert thing fail")
		return -1, err
	}
	return id, nil
}

// The query rule shall be username + resouce，find more record
func (a *OrmMetadataAdapter) findThings(metadatas *[]Metadata, metadata *Metadata) error {
	err := a.engine.Find(metadatas, metadata)
	fmt.Println("ORM find the result thing :", len(*metadatas))
	return err
}

// The query rule shall be username + resouce, get one record
func (a *OrmMetadataAdapter) getThing(metadata *Metadata) (bool, error) {
	return a.engine.Get(*metadata)
}

func (a *OrmPermissionAdapter) deleteThing(metadata Metadata) (int64, error) {
	return a.engine.Delete(metadata)
}
