/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012  Radu Ioan Fericean

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

package timespans

import (
	"fmt"
	"strconv"
)

type Rate struct {
	DestinationsTag                                string
	ConnectFee, Price, PricedUnits, RateIncrements float64
}

func NewRate(destinationsTag, connectFee, price, pricedUnits, rateIncrements string) (r *Rate, err error) {
	cf, err := strconv.ParseFloat(connectFee, 64)
	if err != nil {
		Logger.Err(fmt.Sprintf("Error parsing connect fee from: %v", connectFee))
		return
	}
	p, err := strconv.ParseFloat(price, 64)
	if err != nil {
		Logger.Err(fmt.Sprintf("Error parsing price from: %v", price))
		return
	}
	pu, err := strconv.ParseFloat(pricedUnits, 64)
	if err != nil {
		Logger.Err(fmt.Sprintf("Error parsing priced units from: %v", pricedUnits))
		return
	}
	ri, err := strconv.ParseFloat(rateIncrements, 64)
	if err != nil {
		Logger.Err(fmt.Sprintf("Error parsing rates increments from: %v", rateIncrements))
		return
	}
	r = &Rate{
		DestinationsTag: destinationsTag,
		ConnectFee:      cf,
		Price:           p,
		PricedUnits:     pu,
		RateIncrements:  ri,
	}
	return
}

type Timing struct {
	Years     Years
	Months    Months
	MonthDays MonthDays
	WeekDays  WeekDays
	StartTime string
}

func NewTiming(timeingInfo ...string) (rt *Timing) {
	rt = &Timing{}
	rt.Years.Parse(timeingInfo[0], ";")
	rt.Months.Parse(timeingInfo[1], ";")
	rt.MonthDays.Parse(timeingInfo[2], ";")
	rt.WeekDays.Parse(timeingInfo[3], ";")
	rt.StartTime = timeingInfo[4]
	return
}

type RateTiming struct {
	RatesTag string
	Weight   float64
	timing   *Timing
}

func NewRateTiming(ratesTag string, timing *Timing, weight string) (rt *RateTiming) {
	w, err := strconv.ParseFloat(weight, 64)
	if err != nil {
		Logger.Err(fmt.Sprintf("Error parsing weight unit from: %v", weight))
		return
	}
	rt = &RateTiming{
		RatesTag: ratesTag,
		Weight:   w,
		timing:   timing,
	}
	return
}

func (rt *RateTiming) GetInterval(r *Rate) (i *Interval) {
	i = &Interval{
		Years:          rt.timing.Years,
		Months:         rt.timing.Months,
		MonthDays:      rt.timing.MonthDays,
		WeekDays:       rt.timing.WeekDays,
		StartTime:      rt.timing.StartTime,
		Weight:         rt.Weight,
		ConnectFee:     r.ConnectFee,
		Price:          r.Price,
		PricedUnits:    r.PricedUnits,
		RateIncrements: r.RateIncrements,
	}
	return
}

type CallDescriptors []*CallDescriptor

func (cds CallDescriptors) getKey(key string) *CallDescriptor {
	for _, cd := range cds {
		if cd.GetKey() == key {
			return cd
		}
	}
	return nil
}

/*func (cds CallDescriptors) setIntervalEndTime() {
	for _, cd := range cds {
		for _, ap := range cd.ActivationPeriods {
			for x, i := range ap.Intervals {
				if x < len(ap.Intervals)-1 {
					i.EndTime = ap.Intervals[x+1].StartTime
				}
			}
		}
	}
}*/