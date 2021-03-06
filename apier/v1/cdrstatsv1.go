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

package v1

import (
	"fmt"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// Interact with Stats server
type CDRStatsV1 struct {
	CdrStats *engine.Stats
}

type AttrGetMetrics struct {
	StatsQueueId string // Id of the stats instance queried
}

func (sts *CDRStatsV1) GetMetrics(attr AttrGetMetrics, reply *map[string]float64) error {
	if len(attr.StatsQueueId) == 0 {
		return fmt.Errorf("%s:StatsQueueId", utils.ERR_MANDATORY_IE_MISSING)
	}
	return sts.CdrStats.GetValues(attr.StatsQueueId, reply)
}

func (sts *CDRStatsV1) GetQueueIds(empty string, reply *[]string) error {
	return sts.CdrStats.GetQueueIds(0, reply)
}

func (sts *CDRStatsV1) ReloadQueues(attr utils.AttrCDRStatsReloadQueues, reply *string) error {
	if err := sts.CdrStats.ReloadQueues(attr.StatsQueueIds, nil); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (sts *CDRStatsV1) ResetQueues(attr utils.AttrCDRStatsReloadQueues, reply *string) error {
	if err := sts.CdrStats.ResetQueues(attr.StatsQueueIds, nil); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}
