/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2014 ITsysCOM

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package console

import "github.com/cgrates/cgrates/apier/v1"

func init() {
	c := &CmdAddAccount{
		name:      "account_add",
		rpcMethod: "ApierV1.SetAccount",
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdAddAccount struct {
	name      string
	rpcMethod string
	rpcParams *v1.AttrSetAccount
	*CommandExecuter
}

func (self *CmdAddAccount) Name() string {
	return self.name
}

func (self *CmdAddAccount) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdAddAccount) RpcParams() interface{} {
	if self.rpcParams == nil {
		self.rpcParams = &v1.AttrSetAccount{Direction: "*out"}
	}
	return self.rpcParams
}

func (self *CmdAddAccount) RpcResult() interface{} {
	var s string
	return &s
}
