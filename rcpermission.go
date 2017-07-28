// rcpermission.go  - OpenAPI for kepler resouce management
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
	"net/http"

	"github.com/labstack/echo"
)

// TODO use data structure to include the return data
type PermissionResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Lines   []Line `json:"result"`
	Count   int    `json:"count"`
	Error   string `json:"error"`
}

func (res *PermissionResponse) setResBody(status int, message string, count int, error string, lines []Line) {
	res.Count = count
	res.Error = error
	res.Lines = lines
	res.Message = message
	res.Status = status
}

// The query rule shall be username + resouce
// TODO refactor the response code ,add one function
func getPermission(c echo.Context) error {

	res := &PermissionResponse{}
	line := new(Line)
	lines := make([]Line, 0)

	line.Username = c.QueryParam("username")
	line.PType = c.QueryParam("ptype")
	line.Resource = c.QueryParam("resource")
	line.Action = c.QueryParam("action")

	fmt.Println("username: ", line.Username, "Resource: ", line.Resource, "Action: ", line.Action)

	isinvalid := len(line.Username) == 0 && len(line.Resource) == 0 && len(line.Action) == 0 && len(line.PType) == 0
	if isinvalid == false {
		fmt.Println("query user :", line.Username)
		err := ormAdapter.findUserPermission(&lines, line)
		if err == nil {
			res.setResBody(200, "Find :"+line.Username, len(lines), "", lines)
		} else {
			res.setResBody(500, "Server exception for find action", 0, err.Error(), nil)
		}
	} else {
		res.setResBody(400, "Find Param is invalid, pls provide Ptype ,Username, Resource or Action", 0, ErrParamsType.Error(), nil)

	}
	return c.JSON(http.StatusOK, res)
}

// Delete one policy from database, it is better to have detail info to delete precisely
func deletePermission(c echo.Context) error {
	// Get team and member from the query string
	res := &PermissionResponse{}

	line := new(Line)
	if err := c.Bind(line); err != nil {
		res.setResBody(400, "Delete request Param error,  pls provide Username, Resource & Action", 0, err.Error(), nil)
		return c.JSON(http.StatusOK, res)
	}
	fmt.Println("username: ", line.Username, "Resource: ", line.Resource, "Action: ", line.Action)
	isinvalid := len(line.Username) == 0 || len(line.Resource) == 0 || len(line.Action) == 0
	if isinvalid == false {
		//TODO check return value
		rcEnforce.RemovePolicy(line.Username, line.Resource, line.Action)
		err := rcEnforce.SavePolicy()

		//_, err := ormAdapter.deleteUserPermission(*line)
		if err == nil {
			res.setResBody(200, "Delete successful", 0, "", nil)

		} else {
			res.setResBody(500, "Server exception for delete action", 0, err.Error(), nil)
		}

	} else {
		res.setResBody(400, "delete Param is invalid, pls provide Username, Resource & Action", 0, ErrNeedDeletedCond.Error(), nil)
	}
	return c.JSON(http.StatusOK, res)
}

// update one policy from database, must have all the field : username, ptype, resource and action
func updatePermission(c echo.Context) error {

	res := &PermissionResponse{}
	line := new(Line)
	if err := c.Bind(line); err != nil {
		res.setResBody(400, "Update request Param error,  pls provide Ptype Username, Resource & Action", 0, err.Error(), nil)
		return c.JSON(http.StatusOK, res)
	}

	fmt.Println("username: ", line.Username, "Resource: ", line.Resource, "Action: ", line.Action)

	isinvalid := len(line.Username) == 0 || len(line.Resource) == 0 || len(line.Action) == 0 || len(line.PType) == 0
	if isinvalid != true {
		_, err := ormAdapter.updateUserPermission(line)
		if err == nil {
			res.setResBody(200, "Update successful :"+line.Username, 0, err.Error(), nil)
		} else {
			res.setResBody(500, "Server exception for update action", 0, err.Error(), nil)
		}
	} else {
		res.setResBody(400, "Update Param is invalid, pls provide the full field as ptype,username, resource & action", 0, ErrParamsType.Error(), nil)
	}
	return c.JSON(http.StatusOK, res)
}

func addPermission(c echo.Context) error {

	res := &PermissionResponse{}
	line := new(Line)
	if err := c.Bind(line); err != nil {
		res.setResBody(400, "Add request Param error,  pls provide Ptype Username, Resource & Action", 0, err.Error(), nil)

		return c.JSON(http.StatusOK, res)
	}

	fmt.Println("username: ", line.Username, "Resource: ", line.Resource, "Action: ", line.Action)

	isinvalid := len(line.Username) == 0 || len(line.Resource) == 0 || len(line.Action) == 0 || len(line.PType) == 0
	if isinvalid != true {
		fmt.Println("Add Policy ", "Ptype :", line.PType, "User :", line.Username, "Resource: ", line.Resource, "Action :", line.Action)
		//_, err := ormAdapter.addUserPermission(*line)
		rcEnforce.AddPolicy(line.Username, line.Resource, line.Action)
		err := rcEnforce.SavePolicy()
		if err == nil {
			res.setResBody(200, "Add new policy successful !", 0, "", nil)
		} else {
			res.setResBody(500, "Server exception for Add policy action", 0, err.Error(), nil)
		}
	} else {
		res.setResBody(400, "Add Param is invalid, pls provide the full field as ptype,username, resource & action", 0, ErrParamsType.Error(), nil)
	}

	return c.JSON(http.StatusOK, res)
}

func downloadFile(c echo.Context) error {
	//just for test
	return c.Attachment("1.txt", "2.txt")
	//return c.File("1.txt")
}
