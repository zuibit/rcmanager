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

package main

import (
	"fmt"
	io "io/ioutil"
	"strings"

	json "encoding/json"
	"os"
	"path/filepath"

	"github.com/mholt/archiver"
)

//file path to store things
//TODO set it from profile
var thingsOuputFolder = "things/"

type ThingsMetadata struct {
}

func NewThingsMetadata() *ThingsMetadata {
	return &ThingsMetadata{}
}

func (self *ThingsMetadata) load(filename string, v interface{}) error {

	data, err := io.ReadFile(filename)

	if err != nil {
		fmt.Println("Read Json file error :", err)
		return err
	}

	datajson := []byte(data)
	err = json.Unmarshal(datajson, v)

	if err != nil {
		fmt.Println("Unmarshal Json file error :", err)
		return err
	}
	return nil
}

// return the file path if no error
func (self *ThingsMetadata) getJsonFilePath(filename string) string {
	jsonFile := ""
	err := filepath.Walk(filename, func(path string, f os.FileInfo, err error) error {
		//TODO assume that the ".json" file store in the root path of the thing file, depth check to avoid the performance issue
		depth := strings.Count(path, "/") - strings.Count(filename, "/")
		if depth > 4 {
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
			fmt.Println("Json file path is ", path)
			jsonFile = path
		}

		return nil
	})
	if err != nil {
		fmt.Printf("filepath.Walk() returned %v\n", err)
	}
	return jsonFile
}

//support .ZIP, .TAR or .tar.gz
func handleUploadThing(filename string, ext string, outputfolder string, version float32) error {
	fmt.Println("Handle archive file: ", filename, " ext is :", ext)
	thingsMetadata := NewThingsMetadata()
	metaData := Metadata{}

	var err error
	var desFilePath = rcConfigure.StaticFolder + outputfolder
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
		fmt.Println("Extract file error : ", err)
		return err
	}
	jsonFile := thingsMetadata.getJsonFilePath(desFilePath)
	fmt.Println("Get Json file :", jsonFile)
	if jsonFile == "" {
		return ErrJSONFileNotFound
	}
	if err = thingsMetadata.load(jsonFile, &metaData); err != nil {
		fmt.Println("Load Data error :", err)
		return err
	}
	//TODO need to confirm that each thing only have one property
	metaData.ClassID = metaData.Properties[0].ClassId
	metaData.FilePath = filename
	metaData.Version = version

	ormMetadataAdapter.addthing(metaData)
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
	return err
}

//just for test
func testLoadJsonSetting() {
	thingsMetadata := NewThingsMetadata()
	metaData := Metadata{}
	if err := thingsMetadata.load("static/metadata.json", &metaData); err != nil {
		fmt.Println("Load Data error :", err)
		return
	}
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

}
