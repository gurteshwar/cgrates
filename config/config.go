/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

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

package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cgrates/cgrates/utils"
)

const (
	DISABLED = "disabled"
	JSON     = "json"
	GOB      = "gob"
	POSTGRES = "postgres"
	MONGO    = "mongo"
	REDIS    = "redis"
	SAME     = "same"
	FS       = "freeswitch"
)

var (
	cgrCfg              *CGRConfig       // will be shared
	dfltFsConnConfig    *FsConnConfig    // Default FreeSWITCH Connection configuration, built out of json default configuration
	dfltKamConnConfig   *KamConnConfig   // Default Kamailio Connection configuration
	dfltOsipsConnConfig *OsipsConnConfig // Default OpenSIPS Connection configuration
)

// Used to retrieve system configuration from other packages
func CgrConfig() *CGRConfig {
	return cgrCfg
}

// Used to set system configuration from other places
func SetCgrConfig(cfg *CGRConfig) {
	cgrCfg = cfg
}

func NewDefaultCGRConfig() (*CGRConfig, error) {
	cfg := new(CGRConfig)
	cfg.DataFolderPath = "/usr/share/cgrates/"
	cfg.SmFsConfig = new(SmFsConfig)
	cfg.SmKamConfig = new(SmKamConfig)
	cfg.SmOsipsConfig = new(SmOsipsConfig)
	cfg.ConfigReloads = make(map[string]chan struct{})
	cfg.ConfigReloads[utils.CDRC] = make(chan struct{})
	cgrJsonCfg, err := NewCgrJsonCfgFromReader(strings.NewReader(CGRATES_CFG_JSON))
	if err != nil {
		return nil, err
	}
	if err := cfg.loadFromJsonCfg(cgrJsonCfg); err != nil {
		return nil, err
	}
	cfg.dfltCdreProfile = cfg.CdreProfiles[utils.META_DEFAULT].Clone() // So default will stay unique, will have nil pointer in case of no defaults loaded which is an extra check
	cfg.dfltCdrcProfile = cfg.CdrcProfiles["/var/log/cgrates/cdrc/in"][utils.META_DEFAULT].Clone()
	dfltFsConnConfig = cfg.SmFsConfig.Connections[0] // We leave it crashing here on purpose if no Connection defaults defined
	dfltKamConnConfig = cfg.SmKamConfig.Connections[0]
	dfltOsipsConnConfig = cfg.SmOsipsConfig.Connections[0]
	if err := cfg.checkConfigSanity(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func NewCGRConfigFromJsonString(cfgJsonStr string) (*CGRConfig, error) {
	cfg := new(CGRConfig)
	if jsnCfg, err := NewCgrJsonCfgFromReader(strings.NewReader(cfgJsonStr)); err != nil {
		return nil, err
	} else if err := cfg.loadFromJsonCfg(jsnCfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Reads all .json files out of a folder/subfolders and loads them up in lexical order
func NewCGRConfigFromFolder(cfgDir string) (*CGRConfig, error) {
	if fi, err := os.Stat(cfgDir); err != nil {
		return nil, err
	} else if !fi.IsDir() {
		return nil, fmt.Errorf("Path: %s not a directory.", cfgDir)
	}
	cfg, err := NewDefaultCGRConfig()
	if err != nil {
		return nil, err
	}
	jsonFilesFound := false
	err = filepath.Walk(cfgDir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			return nil
		}
		cfgFiles, err := filepath.Glob(filepath.Join(path, "*.json"))
		if err != nil {
			return err
		}
		if cfgFiles == nil { // No need of processing further since there are no config files in the folder
			return nil
		}
		if !jsonFilesFound {
			jsonFilesFound = true
		}
		for _, jsonFilePath := range cfgFiles {
			if cgrJsonCfg, err := NewCgrJsonCfgFromFile(jsonFilePath); err != nil {
				return err
			} else if err := cfg.loadFromJsonCfg(cgrJsonCfg); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	} else if !jsonFilesFound {
		return nil, fmt.Errorf("No config file found on path %s", cfgDir)
	}
	if err := cfg.checkConfigSanity(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Holds system configuration, defaults are overwritten with values from config file if found
type CGRConfig struct {
	RatingDBType          string
	RatingDBHost          string // The host to connect to. Values that start with / are for UNIX domain sockets.
	RatingDBPort          string // The port to bind to.
	RatingDBName          string // The name of the database to connect to.
	RatingDBUser          string // The user to sign in as.
	RatingDBPass          string // The user's password.
	AccountDBType         string
	AccountDBHost         string // The host to connect to. Values that start with / are for UNIX domain sockets.
	AccountDBPort         string // The port to bind to.
	AccountDBName         string // The name of the database to connect to.
	AccountDBUser         string // The user to sign in as.
	AccountDBPass         string // The user's password.
	StorDBType            string // Should reflect the database type used to store logs
	StorDBHost            string // The host to connect to. Values that start with / are for UNIX domain sockets.
	StorDBPort            string // Th e port to bind to.
	StorDBName            string // The name of the database to connect to.
	StorDBUser            string // The user to sign in as.
	StorDBPass            string // The user's password.
	StorDBMaxOpenConns    int    // Maximum database connections opened
	StorDBMaxIdleConns    int    // Maximum idle connections to keep opened
	DBDataEncoding        string // The encoding used to store object data in strings: <msgpack|json>
	RPCJSONListen         string // RPC JSON listening address
	RPCGOBListen          string // RPC GOB listening address
	HTTPListen            string // HTTP listening address
	DefaultReqType        string // Use this request type if not defined on top
	DefaultCategory       string // set default type of record
	DefaultTenant         string // set default tenant
	DefaultSubject        string // set default rating subject, useful in case of fallback
	RoundingDecimals      int    // Number of decimals to round end prices at
	HttpSkipTlsVerify     bool   // If enabled Http Client will accept any TLS certificate
	TpExportPath          string // Path towards export folder for offline Tariff Plans
	RaterEnabled          bool   // start standalone server (no balancer)
	RaterBalancer         string // balancer address host:port
	BalancerEnabled       bool
	SchedulerEnabled      bool
	CDRSEnabled           bool              // Enable CDR Server service
	CDRSExtraFields       []*utils.RSRField // Extra fields to store in CDRs
	CDRSMediator          string            // Address where to reach the Mediator. Empty for disabling mediation. <""|internal>
	CDRSStats             string            // Address where to reach the Mediator. <""|intenal>
	CDRSStoreDisable      bool              // When true, CDRs will not longer be saved in stordb, useful for cdrstats only scenario
	CDRStatsEnabled       bool              // Enable CDR Stats service
	CDRStatConfig         *CdrStatsConfig   // Active cdr stats configuration instances, platform level
	CdreProfiles          map[string]*CdreConfig
	CdrcProfiles          map[string]map[string]*CdrcConfig // Number of CDRC instances running imports, format map[dirPath]map[instanceName]{Configs}
	SmFsConfig            *SmFsConfig                       // SM-FreeSWITCH configuration
	SmKamConfig           *SmKamConfig                      // SM-Kamailio Configuration
	SmOsipsConfig         *SmOsipsConfig                    // SM-OpenSIPS Configuration
	SMEnabled             bool
	SMSwitchType          string
	SMRater               string        // address where to access rater. Can be internal, direct rater address or the address of a balancer
	SMCdrS                string        // Connection towards CDR server
	SMReconnects          int           // Number of reconnect attempts to rater
	SMDebitInterval       int           // the period to be debited in advanced during a call (in seconds)
	SMMaxCallDuration     time.Duration // The maximum duration of a call
	SMMinCallDuration     time.Duration // Only authorize calls with allowed duration bigger than this
	MediatorEnabled       bool          // Starts Mediator service: <true|false>.
	MediatorReconnects    int           // Number of reconnects to rater before giving up.
	MediatorRater         string
	MediatorStats         string                   // Address where to reach the Rater: <internal|x.y.z.y:1234>
	MediatorStoreDisable  bool                     // When true, CDRs will not longer be saved in stordb, useful for cdrstats only scenario
	FreeswitchServer      string                   // freeswitch address host:port
	FreeswitchPass        string                   // FS socket password
	FreeswitchReconnects  int                      // number of times to attempt reconnect after connect fails
	FSMinDurLowBalance    time.Duration            // Threshold which will trigger low balance warnings
	FSLowBalanceAnnFile   string                   // File to be played when low balance is reached
	FSEmptyBalanceContext string                   // If defined, call will be transfered to this context on empty balance
	FSEmptyBalanceAnnFile string                   // File to be played before disconnecting prepaid calls (applies only if no context defined)
	FSCdrExtraFields      []*utils.RSRField        // Extra fields to store in CDRs in case of processing them
	OsipsListenUdp        string                   // Address where to listen for event datagrams coming from OpenSIPS
	OsipsMiAddr           string                   // Adress where to reach OpenSIPS mi_datagram module
	OsipsEvSubscInterval  time.Duration            // Refresh event subscription at this interval
	OsipsReconnects       int                      // Number of attempts on connect failure.
	KamailioEvApiAddr     string                   // Address of the kamailio evapi server
	KamailioReconnects    int                      // Number of reconnect attempts on connection lost
	HistoryAgentEnabled   bool                     // Starts History as an agent: <true|false>.
	HistoryServer         string                   // Address where to reach the master history server: <internal|x.y.z.y:1234>
	HistoryServerEnabled  bool                     // Starts History as server: <true|false>.
	HistoryDir            string                   // Location on disk where to store history files.
	HistorySaveInterval   time.Duration            // The timout duration between history writes
	MailerServer          string                   // The server to use when sending emails out
	MailerAuthUser        string                   // Authenticate to email server using this user
	MailerAuthPass        string                   // Authenticate to email server with this password
	MailerFromAddr        string                   // From address used when sending emails out
	DataFolderPath        string                   // Path towards data folder, for tests internal usage, not loading out of .json options
	ConfigReloads         map[string]chan struct{} // Signals to specific entities that a config reload should occur
	// Cache defaults loaded from json and needing clones
	dfltCdreProfile *CdreConfig // Default cdreConfig profile
	dfltCdrcProfile *CdrcConfig // Default cdrcConfig profile

}

func (self *CGRConfig) checkConfigSanity() error {
	// CDRC sanity checks
	for _, cdrcInst := range self.CdrcProfiles["/var/log/cgrates/cdrc/in"] {
		if cdrcInst.Enabled == true {
			if len(cdrcInst.CdrFields) == 0 {
				return errors.New("CdrC enabled but no fields to be processed defined!")
			}
			if cdrcInst.CdrFormat == utils.CSV {
				for _, cdrFld := range cdrcInst.CdrFields {
					for _, rsrFld := range cdrFld.Value {
						if _, errConv := strconv.Atoi(rsrFld.Id); errConv != nil && !rsrFld.IsStatic() {
							return fmt.Errorf("CDR fields must be indices in case of .csv files, have instead: %s", rsrFld.Id)
						}
					}
				}
			}
		}
	}
	if self.CDRSStats == utils.INTERNAL && !self.CDRStatsEnabled {
		return errors.New("CDRStats not enabled but requested by CDRS component.")
	}
	if self.MediatorStats == utils.INTERNAL && !self.CDRStatsEnabled {
		return errors.New("CDRStats not enabled but requested by Mediator.")
	}
	if self.SMCdrS == utils.INTERNAL && !self.CDRSEnabled {
		return errors.New("CDRS not enabled but requested by SessionManager")
	}
	return nil
}

// Loads from json configuration object, will be used for defaults, config from file and reload, might need lock
func (self *CGRConfig) loadFromJsonCfg(jsnCfg *CgrJsonCfg) error {
	// Load sections out of JSON config, stop on error
	jsnGeneralCfg, err := jsnCfg.GeneralJsonCfg()
	if err != nil {
		return err
	}
	jsnListenCfg, err := jsnCfg.ListenJsonCfg()
	if err != nil {
		return err
	}
	jsnRatingDbCfg, err := jsnCfg.DbJsonCfg(RATINGDB_JSN)
	if err != nil {
		return err
	}
	jsnAccountingDbCfg, err := jsnCfg.DbJsonCfg(ACCOUNTINGDB_JSN)
	if err != nil {
		return err
	}
	jsnStorDbCfg, err := jsnCfg.DbJsonCfg(STORDB_JSN)
	if err != nil {
		return err
	}
	jsnBalancerCfg, err := jsnCfg.BalancerJsonCfg()
	if err != nil {
		return err
	}
	jsnRaterCfg, err := jsnCfg.RaterJsonCfg()
	if err != nil {
		return err
	}
	jsnSchedCfg, err := jsnCfg.SchedulerJsonCfg()
	if err != nil {
		return err
	}
	jsnCdrsCfg, err := jsnCfg.CdrsJsonCfg()
	if err != nil {
		return err
	}
	jsnMediatorCfg, err := jsnCfg.MediatorJsonCfg()
	if err != nil {
		return err
	}
	jsnCdrstatsCfg, err := jsnCfg.CdrStatsJsonCfg()
	if err != nil {
		return err
	}
	jsnCdreCfg, err := jsnCfg.CdreJsonCfgs()
	if err != nil {
		return err
	}
	jsnCdrcCfg, err := jsnCfg.CdrcJsonCfg()
	if err != nil {
		return err
	}
	jsnSMCfg, err := jsnCfg.SessionManagerJsonCfg()
	if err != nil {
		return err
	}
	jsnFSCfg, err := jsnCfg.FSJsonCfg()
	if err != nil {
		return err
	}
	jsnKamCfg, err := jsnCfg.KamailioJsonCfg()
	if err != nil {
		return err
	}
	jsnOsipsCfg, err := jsnCfg.OsipsJsonCfg()
	if err != nil {
		return err
	}
	jsnSmFsCfg, err := jsnCfg.SmFsJsonCfg()
	if err != nil {
		return err
	}
	jsnSmKamCfg, err := jsnCfg.SmKamJsonCfg()
	if err != nil {
		return err
	}
	jsnSmOsipsCfg, err := jsnCfg.SmOsipsJsonCfg()
	if err != nil {
		return err
	}
	jsnHistServCfg, err := jsnCfg.HistServJsonCfg()
	if err != nil {
		return err
	}
	jsnHistAgentCfg, err := jsnCfg.HistAgentJsonCfg()
	if err != nil {
		return err
	}
	jsnMailerCfg, err := jsnCfg.MailerJsonCfg()
	if err != nil {
		return err
	}
	// All good, start populating config variables
	if jsnRatingDbCfg != nil {
		if jsnRatingDbCfg.Db_type != nil {
			self.RatingDBType = *jsnRatingDbCfg.Db_type
		}
		if jsnRatingDbCfg.Db_host != nil {
			self.RatingDBHost = *jsnRatingDbCfg.Db_host
		}
		if jsnRatingDbCfg.Db_port != nil {
			self.RatingDBPort = strconv.Itoa(*jsnRatingDbCfg.Db_port)
		}
		if jsnRatingDbCfg.Db_name != nil {
			self.RatingDBName = *jsnRatingDbCfg.Db_name
		}
		if jsnRatingDbCfg.Db_user != nil {
			self.RatingDBUser = *jsnRatingDbCfg.Db_user
		}
		if jsnRatingDbCfg.Db_passwd != nil {
			self.RatingDBPass = *jsnRatingDbCfg.Db_passwd
		}
	}
	if jsnAccountingDbCfg != nil {
		if jsnAccountingDbCfg.Db_type != nil {
			self.AccountDBType = *jsnAccountingDbCfg.Db_type
		}
		if jsnAccountingDbCfg.Db_host != nil {
			self.AccountDBHost = *jsnAccountingDbCfg.Db_host
		}
		if jsnAccountingDbCfg.Db_port != nil {
			self.AccountDBPort = strconv.Itoa(*jsnAccountingDbCfg.Db_port)
		}
		if jsnAccountingDbCfg.Db_name != nil {
			self.AccountDBName = *jsnAccountingDbCfg.Db_name
		}
		if jsnAccountingDbCfg.Db_user != nil {
			self.AccountDBUser = *jsnAccountingDbCfg.Db_user
		}
		if jsnAccountingDbCfg.Db_passwd != nil {
			self.AccountDBPass = *jsnAccountingDbCfg.Db_passwd
		}
	}
	if jsnStorDbCfg != nil {
		if jsnStorDbCfg.Db_type != nil {
			self.StorDBType = *jsnStorDbCfg.Db_type
		}
		if jsnStorDbCfg.Db_host != nil {
			self.StorDBHost = *jsnStorDbCfg.Db_host
		}
		if jsnStorDbCfg.Db_port != nil {
			self.StorDBPort = strconv.Itoa(*jsnStorDbCfg.Db_port)
		}
		if jsnStorDbCfg.Db_name != nil {
			self.StorDBName = *jsnStorDbCfg.Db_name
		}
		if jsnStorDbCfg.Db_user != nil {
			self.StorDBUser = *jsnStorDbCfg.Db_user
		}
		if jsnStorDbCfg.Db_passwd != nil {
			self.StorDBPass = *jsnStorDbCfg.Db_passwd
		}
		if jsnStorDbCfg.Max_open_conns != nil {
			self.StorDBMaxOpenConns = *jsnStorDbCfg.Max_open_conns
		}
		if jsnStorDbCfg.Max_idle_conns != nil {
			self.StorDBMaxIdleConns = *jsnStorDbCfg.Max_idle_conns
		}
	}
	if jsnGeneralCfg != nil {
		if jsnGeneralCfg.Dbdata_encoding != nil {
			self.DBDataEncoding = *jsnGeneralCfg.Dbdata_encoding
		}
		if jsnGeneralCfg.Default_reqtype != nil {
			self.DefaultReqType = *jsnGeneralCfg.Default_reqtype
		}
		if jsnGeneralCfg.Default_category != nil {
			self.DefaultCategory = *jsnGeneralCfg.Default_category
		}
		if jsnGeneralCfg.Default_tenant != nil {
			self.DefaultTenant = *jsnGeneralCfg.Default_tenant
		}
		if jsnGeneralCfg.Default_subject != nil {
			self.DefaultSubject = *jsnGeneralCfg.Default_subject
		}
		if jsnGeneralCfg.Rounding_decimals != nil {
			self.RoundingDecimals = *jsnGeneralCfg.Rounding_decimals
		}
		if jsnGeneralCfg.Http_skip_tls_veify != nil {
			self.HttpSkipTlsVerify = *jsnGeneralCfg.Http_skip_tls_veify
		}
		if jsnGeneralCfg.Tpexport_dir != nil {
			self.TpExportPath = *jsnGeneralCfg.Tpexport_dir
		}
	}
	if jsnListenCfg != nil {
		if jsnListenCfg.Rpc_json != nil {
			self.RPCJSONListen = *jsnListenCfg.Rpc_json
		}
		if jsnListenCfg.Rpc_gob != nil {
			self.RPCGOBListen = *jsnListenCfg.Rpc_gob
		}
		if jsnListenCfg.Http != nil {
			self.HTTPListen = *jsnListenCfg.Http
		}
	}
	if jsnRaterCfg != nil {
		if jsnRaterCfg.Enabled != nil {
			self.RaterEnabled = *jsnRaterCfg.Enabled
		}
		if jsnRaterCfg.Balancer != nil {
			self.RaterBalancer = *jsnRaterCfg.Balancer
		}
	}
	if jsnBalancerCfg != nil && jsnBalancerCfg.Enabled != nil {
		self.BalancerEnabled = *jsnBalancerCfg.Enabled
	}
	if jsnSchedCfg != nil && jsnSchedCfg.Enabled != nil {
		self.SchedulerEnabled = *jsnSchedCfg.Enabled
	}
	if jsnCdrsCfg != nil {
		if jsnCdrsCfg.Enabled != nil {
			self.CDRSEnabled = *jsnCdrsCfg.Enabled
		}
		if jsnCdrsCfg.Extra_fields != nil {
			if self.CDRSExtraFields, err = utils.ParseRSRFieldsFromSlice(*jsnCdrsCfg.Extra_fields); err != nil {
				return err
			}
		}
		if jsnCdrsCfg.Mediator != nil {
			self.CDRSMediator = *jsnCdrsCfg.Mediator
		}
		if jsnCdrsCfg.Cdrstats != nil {
			self.CDRSStats = *jsnCdrsCfg.Cdrstats
		}
		if jsnCdrsCfg.Store_disable != nil {
			self.CDRSStoreDisable = *jsnCdrsCfg.Store_disable
		}
	}
	if jsnCdrstatsCfg != nil {
		if jsnCdrstatsCfg.Enabled != nil {
			self.CDRStatsEnabled = *jsnCdrstatsCfg.Enabled
		}
		if jsnCdrstatsCfg != nil { // Have CDRStats config, load it in default object
			if self.CDRStatConfig == nil {
				self.CDRStatConfig = &CdrStatsConfig{Id: utils.META_DEFAULT}
			}
			if err = self.CDRStatConfig.loadFromJsonCfg(jsnCdrstatsCfg); err != nil {
				return err
			}
		}
	}
	if jsnCdreCfg != nil {
		if self.CdreProfiles == nil {
			self.CdreProfiles = make(map[string]*CdreConfig)
		}
		for profileName, jsnCdre1Cfg := range jsnCdreCfg {
			if _, hasProfile := self.CdreProfiles[profileName]; !hasProfile { // New profile, create before loading from json
				self.CdreProfiles[profileName] = new(CdreConfig)
				if profileName != utils.META_DEFAULT {
					self.CdreProfiles[profileName] = self.dfltCdreProfile.Clone() // Clone default so we do not inherit pointers
				}
			}
			if err = self.CdreProfiles[profileName].loadFromJsonCfg(jsnCdre1Cfg); err != nil { // Update the existing profile with content from json config
				return err
			}
		}
	}
	if jsnCdrcCfg != nil {
		if self.CdrcProfiles == nil {
			self.CdrcProfiles = make(map[string]map[string]*CdrcConfig)
		}
		for profileName, jsnCrc1Cfg := range jsnCdrcCfg {
			if _, hasDir := self.CdrcProfiles[*jsnCrc1Cfg.Cdr_in_dir]; !hasDir {
				self.CdrcProfiles[*jsnCrc1Cfg.Cdr_in_dir] = make(map[string]*CdrcConfig)
			}
			if _, hasProfile := self.CdrcProfiles[profileName]; !hasProfile {
				if profileName == utils.META_DEFAULT {
					self.CdrcProfiles[*jsnCrc1Cfg.Cdr_in_dir][profileName] = new(CdrcConfig)
				} else {
					self.CdrcProfiles[*jsnCrc1Cfg.Cdr_in_dir][profileName] = self.dfltCdrcProfile.Clone() // Clone default so we do not inherit pointers
				}
			}
			if err = self.CdrcProfiles[*jsnCrc1Cfg.Cdr_in_dir][profileName].loadFromJsonCfg(jsnCrc1Cfg); err != nil {
				return err
			}
		}
	}
	if jsnSMCfg != nil {
		if jsnSMCfg.Enabled != nil {
			self.SMEnabled = *jsnSMCfg.Enabled
		}
		if jsnSMCfg.Switch_type != nil {
			self.SMSwitchType = *jsnSMCfg.Switch_type
		}
		if jsnSMCfg.Rater != nil {
			self.SMRater = *jsnSMCfg.Rater
		}
		if jsnSMCfg.Cdrs != nil {
			self.SMCdrS = *jsnSMCfg.Cdrs
		}
		if jsnSMCfg.Reconnects != nil {
			self.SMReconnects = *jsnSMCfg.Reconnects
		}
		if jsnSMCfg.Debit_interval != nil {
			self.SMDebitInterval = *jsnSMCfg.Debit_interval
		}
		if jsnSMCfg.Max_call_duration != nil {
			if self.SMMaxCallDuration, err = utils.ParseDurationWithSecs(*jsnSMCfg.Max_call_duration); err != nil {
				return err
			}
		}
		if jsnSMCfg.Min_call_duration != nil {
			if self.SMMinCallDuration, err = utils.ParseDurationWithSecs(*jsnSMCfg.Min_call_duration); err != nil {
				return err
			}
		}
	}
	if jsnFSCfg != nil {
		if jsnFSCfg.Server != nil {
			self.FreeswitchServer = *jsnFSCfg.Server
		}
		if jsnFSCfg.Password != nil {
			self.FreeswitchPass = *jsnFSCfg.Password
		}
		if jsnFSCfg.Reconnects != nil {
			self.FreeswitchReconnects = *jsnFSCfg.Reconnects
		}
		if jsnFSCfg.Min_dur_low_balance != nil {
			if self.FSMinDurLowBalance, err = utils.ParseDurationWithSecs(*jsnFSCfg.Min_dur_low_balance); err != nil {
				return err
			}
		}
		if jsnFSCfg.Low_balance_ann_file != nil {
			self.FSLowBalanceAnnFile = *jsnFSCfg.Low_balance_ann_file
		}
		if jsnFSCfg.Empty_balance_context != nil {
			self.FSEmptyBalanceContext = *jsnFSCfg.Empty_balance_context
		}
		if jsnFSCfg.Empty_balance_ann_file != nil {
			self.FSEmptyBalanceAnnFile = *jsnFSCfg.Empty_balance_ann_file
		}
		if jsnFSCfg.Cdr_extra_fields != nil {
			if self.FSCdrExtraFields, err = utils.ParseRSRFieldsFromSlice(*jsnFSCfg.Cdr_extra_fields); err != nil {
				return err
			}
		}
	}
	if jsnOsipsCfg != nil {
		if jsnOsipsCfg.Listen_udp != nil {
			self.OsipsListenUdp = *jsnOsipsCfg.Listen_udp
		}
		if jsnOsipsCfg.Mi_addr != nil {
			self.OsipsMiAddr = *jsnOsipsCfg.Mi_addr
		}
		if jsnOsipsCfg.Events_subscribe_interval != nil {
			if self.OsipsEvSubscInterval, err = utils.ParseDurationWithSecs(*jsnOsipsCfg.Events_subscribe_interval); err != nil {
				return err
			}
		}
		if jsnOsipsCfg.Reconnects != nil {
			self.OsipsReconnects = *jsnOsipsCfg.Reconnects
		}
	}
	if jsnKamCfg != nil {
		if jsnKamCfg.Evapi_addr != nil {
			self.KamailioEvApiAddr = *jsnKamCfg.Evapi_addr
		}
		if jsnKamCfg.Reconnects != nil {
			self.KamailioReconnects = *jsnKamCfg.Reconnects
		}
	}
	if jsnSmFsCfg != nil {
		if err := self.SmFsConfig.loadFromJsonCfg(jsnSmFsCfg); err != nil {
			return err
		}
	}
	if jsnSmKamCfg != nil {
		if err := self.SmKamConfig.loadFromJsonCfg(jsnSmKamCfg); err != nil {
			return err
		}
	}
	if jsnSmOsipsCfg != nil {
		if err := self.SmOsipsConfig.loadFromJsonCfg(jsnSmOsipsCfg); err != nil {
			return err
		}
	}
	if jsnMediatorCfg != nil {
		if jsnMediatorCfg.Enabled != nil {
			self.MediatorEnabled = *jsnMediatorCfg.Enabled
		}
		if jsnMediatorCfg.Reconnects != nil {
			self.MediatorReconnects = *jsnMediatorCfg.Reconnects
		}
		if jsnMediatorCfg.Rater != nil {
			self.MediatorRater = *jsnMediatorCfg.Rater
		}
		if jsnMediatorCfg.Cdrstats != nil {
			self.MediatorStats = *jsnMediatorCfg.Cdrstats
		}
		if jsnMediatorCfg.Store_disable != nil {
			self.MediatorStoreDisable = *jsnMediatorCfg.Store_disable
		}
	}

	if jsnHistAgentCfg != nil {
		if jsnHistAgentCfg.Enabled != nil {
			self.HistoryAgentEnabled = *jsnHistAgentCfg.Enabled
		}
		if jsnHistAgentCfg.Server != nil {
			self.HistoryServer = *jsnHistAgentCfg.Server
		}
	}
	if jsnHistServCfg != nil {
		if jsnHistServCfg.Enabled != nil {
			self.HistoryServerEnabled = *jsnHistServCfg.Enabled
		}
		if jsnHistServCfg.History_dir != nil {
			self.HistoryDir = *jsnHistServCfg.History_dir
		}
		if jsnHistServCfg.Save_interval != nil {
			if self.HistorySaveInterval, err = utils.ParseDurationWithSecs(*jsnHistServCfg.Save_interval); err != nil {
				return err
			}
		}
	}
	if jsnMailerCfg != nil {
		if jsnMailerCfg.Server != nil {
			self.MailerServer = *jsnMailerCfg.Server
		}
		if jsnMailerCfg.Auth_user != nil {
			self.MailerAuthUser = *jsnMailerCfg.Auth_user
		}
		if jsnMailerCfg.Auth_passwd != nil {
			self.MailerAuthPass = *jsnMailerCfg.Auth_passwd
		}
		if jsnMailerCfg.From_address != nil {
			self.MailerFromAddr = *jsnMailerCfg.From_address
		}
	}
	return nil
}
