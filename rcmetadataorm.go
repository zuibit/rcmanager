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
	"runtime"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	log "github.com/sirupsen/logrus"
	//"github.com/go-xorm/xorm-redis-cache"
)

type Property struct {
	ClassName      string   `json:"class_name"`
	Name           string   `json:"name"`   // name of things
	Vendor         string   `json:"vendor"` // Vendor of things
	Author         string   `json:"author"`
	Version        string   `json:"version"`
	ReleaseTime    string   `json:"release_time"`    // change to time format
	ClassId        string   `json:"class_id"`        // reserved but not used
	TargetPlatform string   `json:"target_platform"` // reserved but not used
	Description    string   `json:"description"`     // reserved ut not used
	DependClass    []string `json:"dependclass"`
}

type Metadata struct {
	//Id          int64      `xorm:"autoincr" json:"-"`
	ClassId     string     `xorm:"varchar(100) PK index unique"`
	PackageName string     `xorm:"varchar(100)  index 'packageName'" json:"package_name"`
	Properties  []Property `xorm:"Text json 'properties'" json:"metadata"`
	FilePath    string     `xorm:"varchar(100)" json:"uri"`
	Version     string     `xorm:"varchar(100)"`
	CreateAt    time.Time  `xorm:"created" json:"-"`
	//UpdateAt    time.Time  `xorm:"updated" json:"-"`
}

type OrmMetadataAdapter struct {
	driverName     string
	dataSourceName string
	engine         *xorm.Engine
}

// finalizer is the destructor for Adapter.
func metadatafinalizer(a *OrmMetadataAdapter) {
	a.engine.Close()
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
		rcLog.Panic("Could not create thing database")
		panic(err)
	}

	if a.driverName == rcConfigure.Drivername {
		engine, err = xorm.NewEngine(a.driverName, a.dataSourceName+"things")
	}
	if err != nil {
		rcLog.WithFields(log.Fields{
			"Driver ":      a.driverName,
			"Data Source ": a.dataSourceName + "things",
		}).Panic("Could not setup the ORM engine")
		panic(err)
	}

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
		rcLog.WithFields(log.Fields{
			"Driver ":      a.driverName,
			"Data Source ": a.dataSourceName,
		}).Panic("Could not sync the table things")
		panic(err)
	}
}

func (a *OrmMetadataAdapter) addthing(metadata Metadata) (int64, error) {

	rcLog.WithFields(log.Fields{
		"PackageName ": metadata.PackageName,
		"ClassID ":     metadata.ClassId,
	}).Info("Add thing in database")

	has, err := a.engine.Get(&metadata)
	if err != nil {
		return -1, err
	}
	if has == true {
		return -1, ErrAlreadyExist
	}
	id, err := a.engine.Insert(&metadata)
	if err != nil {
		rcLog.WithFields(log.Fields{
			"PackageName ": metadata.PackageName,
			"ClassID ":     metadata.ClassId,
		}).Error("Could not add thing in database")

		return -1, err
	}
	return id, nil
}

// The query rule shall be username + resouce，find more record
func (a *OrmMetadataAdapter) findThings(metadatas *[]Metadata, metadata *Metadata) error {

	rcLog.WithFields(log.Fields{
		"PackageName ": metadata.PackageName,
		"ClassID ":     metadata.ClassId,
	}).Info("Find thing in database")

	return a.engine.Find(metadatas, metadata)
}

func (a *OrmMetadataAdapter) getThing(metadata *Metadata) (bool, error) {
	rcLog.WithFields(log.Fields{
		"PackageName ": metadata.PackageName,
		"ClassID ":     metadata.ClassId,
	}).Info("Get thing in database")

	return a.engine.Get(metadata)
}

func (a *OrmMetadataAdapter) deleteThing(metadata *Metadata) (int64, error) {

	rcLog.WithFields(log.Fields{
		"PackageName ": metadata.PackageName,
		"ClassID ":     metadata.ClassId,
	}).Info("Delete thing in database")

	return a.engine.Where("class_Id = ?", metadata.ClassId).Delete(metadata)
}
