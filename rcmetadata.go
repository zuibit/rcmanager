// rcmetadata.go  - Metadata in upload things package  for  kepler resouce management service
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

//TODO move this file to thing api

package main

import (
	io "io/ioutil"
	"strings"

	json "encoding/json"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/mholt/archiver"
)

var tempExtractFolder = "thingstmp"

type ThingsMetadata struct {
}

func NewThingsMetadata() *ThingsMetadata {
	return &ThingsMetadata{}
}

func (self *ThingsMetadata) load(filename string, v interface{}) error {
	data, err := io.ReadFile(filename)

	if err != nil {
		rcLog.WithFields(log.Fields{
			"Thing Metadata File": filename,
			"Error":               err,
		}).Error("Read Json file error")
		return err
	}

	datajson := []byte(data)
	err = json.Unmarshal(datajson, v)

	if err != nil {
		return err
	}
	return nil
}

// return the file path if no error
func (self *ThingsMetadata) getJsonFilePath(filename string) (string, error) {
	jsonFile := ""
	err := filepath.Walk(filename, func(path string, f os.FileInfo, err error) error {
		//TODO assume that the ".json" file store in the root path of the thing file
		//depth check to avoid the performance issue
		depth := strings.Count(path, "/") - strings.Count(filename, "/")
		if depth > 3 {
			return filepath.SkipDir
		}

		if f == nil {
			return err
		}
		if f.IsDir() {
			return nil
		}
		//TODO define the file format and use const to replace the hard code
		//shall be only one .json file
		if strings.Contains(path, ".json") {
			rcLog.WithFields(log.Fields{
				"Thing Metadata File": path,
				"Archive Package":     filename,
			}).Info("Found Metadata File in compress package")
			jsonFile = path
		}
		return nil
	})
	if err != nil {
		rcLog.WithFields(log.Fields{
			"Archive Package": filename,
			"Error is ":       err,
		}).Info("Not Found Metadata File in compress package")
	}
	return jsonFile, err
}

//support .ZIP, .TAR or .tar.gz
func handleUploadThing(filename string, ext string, outputfolder string) (*Metadata, error) {

	rcLog.WithFields(log.Fields{
		"Archive Package": filename,
		"Ext ":            ext,
	}).Info("Handle archive package")

	thingsMetadata := NewThingsMetadata()
	metaData := Metadata{}

	var err error
	//extract folder such as static/speaker/test
	desFilePath, err := io.TempDir(rcConfigure.StaticFolder, outputfolder)
	if err != nil {
		return nil, err
	}
	switch ext {
	case ".zip":
		err = archiver.Zip.Open(filename, desFilePath)

	case ".tar":
		err = archiver.Tar.Open(filename, desFilePath)

	case ".gz":
		err = archiver.TarGz.Open(filename, desFilePath)

	default:
		err = ErrFileTypeNotSupported
	}

	if nil != err {
		rcLog.WithFields(log.Fields{
			"Archive Package": filename,
			"Error ":          err,
		}).Error("Could not extract archive package")

		return nil, err
	}
	jsonFile, err := thingsMetadata.getJsonFilePath(desFilePath)

	if err != nil {
		return nil, err
	}
	if jsonFile == "" {
		return nil, ErrJSONFileNotFound
	}
	if err = thingsMetadata.load(jsonFile, &metaData); err != nil {
		return nil, err
	}
	//Each version thing only have one ClassId
	metaData.ClassId = metaData.Properties[0].ClassId
	metaData.FilePath = filename
	metaData.Version = metaData.Properties[0].Version
	_, err = ormMetadataAdapter.addthing(metaData)
	os.RemoveAll(desFilePath)
	/*
		fmt.Println("Load data :", metaData)
		fmt.Println("Package Name is ", metaData.PackageName)
		fmt.Println("Class Name is ", metaData.Properties[0].ClassName)
		fmt.Println("Class ID is ", metaData.Properties[0].ClassId)
		fmt.Println("Description is ", metaData.Properties[0].Description)
		fmt.Println("Name is ", metaData.Properties[0].Name)
		fmt.Println("Release Time is ", metaData.Properties[0].ReleaseTime)
		fmt.Println("Release Time is ", metaData.Properties[0].TargetPlatform)
		fmt.Println("Vender is ", metaData.Properties[0].Vendor)
		fmt.Println("DependClass is ", metaData.Properties[0].DependClass)
	*/
	return &metaData, err
}
