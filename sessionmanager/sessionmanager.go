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

package sessionmanager

import (
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type SessionManager interface {
	Connect() error
	DisconnectSession(utils.Event, string)
	RemoveSession(string)
	MaxDebit(*engine.CallDescriptor, *engine.CallCost) error
	GetDebitPeriod() time.Duration
	GetDbLogger() engine.LogStorage
	Rater() engine.Connector
	Shutdown() error
}
