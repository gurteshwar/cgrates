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

package sessionmanager

//"github.com/cgrates/cgrates/config"
//"testing"

var (
	newEventBody = `
"Event-Name":	"HEARTBEAT",
"Core-UUID":	"d5abc5b0-95c6-11e1-be05-43c90197c914",
"FreeSWITCH-Hostname":	"grace",
"FreeSWITCH-Switchname":	"grace",
"FreeSWITCH-IPv4":	"172.17.77.126",
"variable_sip_full_from":	"rif",
"variable_cgr_account":	"rif",
"variable_sip_full_to":	"0723045326",
"Caller-Dialplan":	"vdf",
"FreeSWITCH-IPv6":	"::1",
"Event-Date-Local":	"2012-05-04 14:38:23",
"Event-Date-GMT":	"Fri, 03 May 2012 11:38:23 GMT",
"Event-Date-Timestamp":	"1336131503218867",
"Event-Calling-File":	"switch_core.c",
"Event-Calling-Function":	"send_heartbeat",
"Event-Calling-Line-Number":	"68",
"Event-Sequence":	"4171",
"Event-Info":	"System Ready",
"Up-Time":	"0 years, 0 days, 2 hours, 43 minutes, 21 seconds, 349 milliseconds, 683 microseconds",
"Session-Count":	"0",
"Max-Sessions":	"1000",
"Session-Per-Sec":	"30",
"Session-Since-Startup":	"122",
"Idle-CPU":	"100.000000"
`
	conf_data = []byte(`
### Test data, not for production usage

[global]
default_reqtype=
`)
)

/* Missing parameter is not longer tested in NewSession. ToDo: expand this test for more util information
func TestSessionNilSession(t *testing.T) {
	var errCfg error
	cfg, errCfg = config.NewCGRConfigBytes(conf_data) // Needed here to avoid nil on cfg variable
	if errCfg != nil {
		t.Errorf("Cannot get configuration %v", errCfg)
	}
	newEvent := new(FSEvent).New("")
	sm := &FSSessionManager{}
	s := NewSession(newEvent, sm)
	if s != nil {
		t.Error("no account and it still created session.")
	}
}
*/
