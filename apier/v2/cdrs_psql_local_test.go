/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can Storagetribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITH*out ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package v2

import (
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os/exec"
	"path"
	"testing"
	"time"
)

var cdrsPsqlCfgPath string
var cdrsPsqlCfg *config.CGRConfig
var cdrsPsqlRpc *rpc.Client
var cmdEngineCdrPsql *exec.Cmd

func TestV2CdrsPsqlInitConfig(t *testing.T) {
	if !*testLocal {
		return
	}
	var err error
	cdrsPsqlCfgPath = path.Join(*dataDir, "conf", "samples", "cdrsv2psql")
	if cdrsPsqlCfg, err = config.NewCGRConfigFromFolder(cdrsPsqlCfgPath); err != nil {
		t.Fatal(err)
	}
}

func TestV2CdrsPsqlInitDataDb(t *testing.T) {
	if !*testLocal {
		return
	}
	if err := engine.InitDataDb(cdrsPsqlCfg); err != nil {
		t.Fatal(err)
	}
}

// InitDb so we can rely on count
func TestV2CdrsPsqlInitCdrDb(t *testing.T) {
	if !*testLocal {
		return
	}
	if err := engine.InitCdrDb(cdrsPsqlCfg); err != nil {
		t.Fatal(err)
	}
}

func TestV2CdrsPsqlInjectUnratedCdr(t *testing.T) {
	if !*testLocal {
		return
	}
	var psqlDb *engine.PostgresStorage
	if d, err := engine.NewPostgresStorage(cdrsPsqlCfg.StorDBHost, cdrsPsqlCfg.StorDBPort, cdrsPsqlCfg.StorDBName, cdrsPsqlCfg.StorDBUser, cdrsPsqlCfg.StorDBPass,
		cdrsPsqlCfg.StorDBMaxOpenConns, cdrsPsqlCfg.StorDBMaxIdleConns); err != nil {
		t.Error("Error on opening database connection: ", err)
		return
	} else {
		psqlDb = d.(*engine.PostgresStorage)
	}
	strCdr1 := &utils.StoredCdr{CgrId: utils.Sha1("bbb1", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String()),
		TOR: utils.VOICE, AccId: "bbb1", CdrHost: "192.168.1.1", CdrSource: "UNKNOWN", ReqType: "rated",
		Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC), AnswerTime: time.Date(2013, 12, 7, 8, 42, 26, 0, time.UTC),
		Usage: time.Duration(10) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		MediationRunId: utils.DEFAULT_RUNID, Cost: 1.201}
	if err := psqlDb.SetCdr(strCdr1); err != nil {
		t.Error(err.Error())
	}
}

func TestV2CdrsPsqlStartEngine(t *testing.T) {
	if !*testLocal {
		return
	}
	var err error
	if cmdEngineCdrPsql, err = engine.StartEngine(cdrsPsqlCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestV2CdrsPsqlPsqlRpcConn(t *testing.T) {
	if !*testLocal {
		return
	}
	var err error
	cdrsPsqlRpc, err = jsonrpc.Dial("tcp", cdrsPsqlCfg.RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

// Insert some CDRs
func TestV2CdrsPsqlProcessCdr(t *testing.T) {
	if !*testLocal {
		return
	}
	var reply string
	cdrs := []*utils.StoredCdr{
		&utils.StoredCdr{CgrId: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), OrderId: 123, TOR: utils.VOICE, AccId: "dsafdsaf",
			CdrHost: "192.168.1.1", CdrSource: "test", ReqType: "rated", Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
			SetupTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), MediationRunId: utils.DEFAULT_RUNID,
			Usage: time.Duration(10) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
			RatedAccount: "dan", RatedSubject: "dans", Rated: true,
		},
		&utils.StoredCdr{CgrId: utils.Sha1("abcdeftg", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), OrderId: 123, TOR: utils.VOICE, AccId: "dsafdsaf",
			CdrHost: "192.168.1.1", CdrSource: "test", ReqType: "rated", Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "1002", Subject: "1002", Destination: "1002",
			SetupTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), MediationRunId: utils.DEFAULT_RUNID,
			Usage: time.Duration(10) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
			RatedAccount: "dan", RatedSubject: "dans",
		},
		&utils.StoredCdr{CgrId: utils.Sha1("aererfddf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), OrderId: 123, TOR: utils.VOICE, AccId: "dsafdsaf",
			CdrHost: "192.168.1.1", CdrSource: "test", ReqType: "rated", Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "1003", Subject: "1003", Destination: "1002",
			SetupTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), MediationRunId: utils.DEFAULT_RUNID,
			Usage: time.Duration(10) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
			RatedAccount: "dan", RatedSubject: "dans",
		},
	}
	for _, cdr := range cdrs {
		if err := cdrsPsqlRpc.Call("CdrsV2.ProcessCdr", cdr, &reply); err != nil {
			t.Error("Unexpected error: ", err.Error())
		} else if reply != utils.OK {
			t.Error("Unexpected reply received: ", reply)
		}
	}
}

func TestV2CdrsPsqlGetCdrs(t *testing.T) {
	if !*testLocal {
		return
	}
	var reply []*utils.StoredCdr
	req := utils.RpcCdrsFilter{}
	if err := cdrsPsqlRpc.Call("ApierV2.GetCdrs", req, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 4 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	}
	// CDRs with errors
	req = utils.RpcCdrsFilter{CostStart: utils.Float64Pointer(-1.0), CostEnd: utils.Float64Pointer(0.0)}
	if err := cdrsPsqlRpc.Call("ApierV2.GetCdrs", req, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 2 {
		t.Error("Unexpected number of CDRs returned: ", reply)
	}
	// CDRs Rated
	req = utils.RpcCdrsFilter{CostStart: utils.Float64Pointer(-1.0)}
	if err := cdrsPsqlRpc.Call("ApierV2.GetCdrs", req, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 3 {
		t.Error("Unexpected number of CDRs returned: ", reply)
	}
	// CDRs non rated OR SkipRated
	req = utils.RpcCdrsFilter{CostEnd: utils.Float64Pointer(-1.0)}
	if err := cdrsPsqlRpc.Call("ApierV2.GetCdrs", req, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 1 {
		t.Error("Unexpected number of CDRs returned: ", reply)
	}
	// Skip Errors
	req = utils.RpcCdrsFilter{CostStart: utils.Float64Pointer(0.0), CostEnd: utils.Float64Pointer(-1.0)}
	if err := cdrsPsqlRpc.Call("ApierV2.GetCdrs", req, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 2 {
		t.Error("Unexpected number of CDRs returned: ", reply)
	}
}

func TestV2CdrsPsqlCountCdrs(t *testing.T) {
	if !*testLocal {
		return
	}
	var reply int64
	req := utils.AttrGetCdrs{}
	if err := cdrsPsqlRpc.Call("ApierV2.CountCdrs", req, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != 4 {
		t.Error("Unexpected number of CDRs returned: ", reply)
	}
}

func TestV2CdrsPsqlKillEngine(t *testing.T) {
	if !*testLocal {
		return
	}
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
