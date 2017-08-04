// rcthingapi.go  - thing OpenAPI for kepler resouce management
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
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/labstack/echo"
	log "github.com/sirupsen/logrus"
)

type ThingResponse struct {
	Status   int        `json:"status"`
	Message  string     `json:"message"`
	Property []Metadata `json:"result"`
	Error    string     `json:"error"`
}

func (res *ThingResponse) setResBody(status int, message string, property []Metadata, err string) {
	res.Error = err
	res.Message = message
	res.Property = property
	res.Status = status
}

func upLoadThing(c echo.Context) error {
	//TODO Limit the upload file size
	//TODO store the dst file on cloud, not local
	res := &ThingResponse{}
	metadata := new(Metadata)

	thing := c.FormValue("thing")

	if thing == "" {
		res.setResBody(400, "please provide thing name", nil, ErrParamsType.Error())
		return c.JSON(http.StatusOK, res)
	}

	// Source
	file, err := c.FormFile("file")
	rcLog.WithFields(log.Fields{
		"Request IP":  c.Request().RemoteAddr,
		"Upload File": file.Filename,
		"File Header": file.Header,
	}).Info("Delete thing in database")

	if err != nil {
		res.setResBody(400, "Source file error", nil, err.Error())
		return c.JSON(http.StatusOK, res)
	}

	src, err := file.Open()
	if err != nil {
		res.setResBody(400, "Upload file open exception", nil, err.Error())
		return c.JSON(http.StatusOK, res)
	}
	defer src.Close()
	//TODO file path add the timestamp to avoid upload duplicate problem
	//	timestamp := strconv.FormatInt(time.Now().Unix(), 16)
	time := time.Now().Format("20060102-150405")

	desFolder := rcConfigure.StaticFolder + "/" + thing + time
	os.Mkdir(desFolder, os.ModePerm)
	dst, err := os.Create(desFolder + "/" + file.Filename)

	if err != nil {
		res.setResBody(500, "Server file create exception", nil, err.Error())
		return c.JSON(http.StatusOK, res)
	}
	defer dst.Close()

	// Copy
	if _, err = io.Copy(dst, src); err != nil {
		res.setResBody(500, "Server file copy exception", nil, err.Error())
		return c.JSON(http.StatusOK, res)
	}
	rcLog.WithFields(log.Fields{
		"Request IP":  c.Request().RemoteAddr,
		"Upload File": file.Filename,
		"Dst File":    dst.Name(),
	}).Info("File Copy complete")

	if fileSuffix := path.Ext(file.Filename); strings.Contains(fileSuffix, "zip") || strings.Contains(fileSuffix, "tar") || strings.Contains(fileSuffix, ".gz") {
		//TODO maybe some performance issue

		metadata, err = handleUploadThing(dst.Name(), fileSuffix, thing)
	}
	dstName := dst.Name()
	// if don't close the file ,could not remove the file properly
	dst.Close()
	if err != nil {
		//os.Remove(dstName)
		os.RemoveAll(path.Dir(dstName))
		res.setResBody(500, "Server exception for upload file", nil, err.Error())
	} else {
		metadatas := make([]Metadata, 1)
		metadatas[0] = *metadata
		res.setResBody(200, "Server thing create successfuly", metadatas, "")
	}
	return c.JSON(http.StatusOK, res)
}

// delete thing according to the class_ID
func deleteThing(c echo.Context) error {
	res := &ThingResponse{}
	metadata := Metadata{}
	metadatas := make([]Metadata, 1)

	metadata.ClassId = c.FormValue("classid")
	if metadata.ClassId == "" {
		res.setResBody(400, "Pls provide classid for delete action ", nil, ErrParamsType.Error())
		return c.JSON(http.StatusOK, res)
	}
	has, err := ormMetadataAdapter.getThing(&metadata)
	if has != true {
		res.setResBody(400, "Delete thing does not exist", nil, ErrNotExist.Error())
		return c.JSON(http.StatusOK, res)
	}

	rcLog.WithFields(log.Fields{
		"Request IP":            c.Request().RemoteAddr,
		"Delete File on server": metadata.FilePath,
	}).Info("Delete Server File")

	//we don't need to check the io delete result
	//just make sure database record is removed
	os.RemoveAll(path.Dir(metadata.FilePath))
	_, err = ormMetadataAdapter.deleteThing(&metadata)
	metadatas[0] = metadata

	if err != nil {
		rcLog.WithFields(log.Fields{
			"Request IP":     c.Request().RemoteAddr,
			"Delete ClassID": metadata.ClassId,
		}).Error("Delete Server Thing from DB fail")

		res.setResBody(500, "Delete thing exception", metadatas, err.Error())
		return c.JSON(http.StatusOK, res)
	}
	res.setResBody(200, "Delete thing successfully", metadatas, "")
	return c.JSON(http.StatusOK, res)
}

// find the things with packagename or classid
func getThings(c echo.Context) error {
	res := &ThingResponse{}
	metadata := new(Metadata)
	metadatas := make([]Metadata, 0)

	metadata.PackageName = c.QueryParam("packagename")
	metadata.ClassId = c.QueryParam("classid")

	isinvalid := len(metadata.PackageName) == 0 && len(metadata.ClassId) == 0
	if isinvalid == false {
		err := ormMetadataAdapter.findThings(&metadatas, metadata)
		if err == nil {
			res.setResBody(200, "Find :"+metadata.PackageName, metadatas, "")
		} else {
			rcLog.WithFields(log.Fields{
				"Request IP":  c.Request().RemoteAddr,
				"ClassID":     metadata.ClassId,
				"PackageName": metadata.PackageName,
				"Error":       err,
			}).Error("Find Server Thing from DB exception")

			res.setResBody(500, "Server exception for find action", nil, err.Error())
		}
	} else {
		res.setResBody(400, "Find Param is invalid, pls provide classid or packagename", nil, ErrParamsType.Error())
	}
	return c.JSON(http.StatusOK, res)
}

// provide the thing classid, return the download file
func downloadThing(c echo.Context) error {
	res := &ThingResponse{}
	metadata := Metadata{}
	metadatas := make([]Metadata, 1)

	metadata.ClassId = c.QueryParam("classid")
	metadatas[0] = metadata
	isinvalid := len(metadata.ClassId) == 0
	if isinvalid == false {
		has, err := ormMetadataAdapter.getThing(&metadata)
		if err != nil {
			rcLog.WithFields(log.Fields{
				"Request IP":  c.Request().RemoteAddr,
				"ClassID":     metadata.ClassId,
				"PackageName": metadata.PackageName,
				"Error":       err,
			}).Error("Find Server Thing from DB , could not download")

			res.setResBody(500, "Find ClassID exception:"+metadata.ClassId, metadatas, err.Error())
			return c.JSON(http.StatusOK, res)
		}
		if has == false {
			//Did not find the record
			res.setResBody(500, "Things does not exist", metadatas, ErrNotExist.Error())
			return c.JSON(http.StatusOK, res)
		}
		//get the file path and return the file

		_, err = os.Stat(metadata.FilePath)
		if err != nil {
			if os.IsNotExist(err) {
				res.setResBody(500, "Things file does not exist", metadatas, ErrNotExist.Error())
				return c.JSON(http.StatusOK, res)
			}
		}
		return c.Attachment(metadata.FilePath, path.Base(metadata.FilePath))
	} else {
		res.setResBody(400, "Find Param is invalid, pls provide ClassID", metadatas, ErrParamsType.Error())
		return c.JSON(http.StatusOK, res)
	}
}
