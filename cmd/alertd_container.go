package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
)

// MetricCheck stores the name of the alert, a function, and a active boolean
type MetricCheck struct {
	AlertActive bool
	Limit       *uint64
	MinDelay	*uint64
	Delaying	bool
	DelaySince	time.Time
}

// ToggleAlertActive changes the state of the alert
func (c *MetricCheck) ToggleAlertActive() {
	c.AlertActive = !c.AlertActive
}

// StaticCheck checks the container for some static thing that is not based on usage
// statistics, like its existence, whether it is running or not, etc.
type StaticCheck struct {
	AlertActive bool
	Expected    *bool
	MinDelay	*uint64
	Delaying	bool
	DelaySince	time.Time
}

// ToggleAlertActive changes the state of the alert
func (c *StaticCheck) ToggleAlertActive() {
	c.AlertActive = !c.AlertActive
}

// Checker interface has all of the methods necessary to check a container
type Checker interface {
	CPUCheck(s *types.Stats)
	MemCheck(s *types.Stats)
	PIDCheck(s *types.Stats)
	ExistenceCheck(a *types.ContainerJSON, e error)
	RunningCheck(a *types.ContainerJSON, e error)
}

// AlertdContainer has the name of the container and the StaticChecks, and MetricChecks
// which are to be run on the container.
type AlertdContainer struct {
	Name     string `json:"name"`
	AlertList    *AlertList
	CPUCheck *MetricCheck
	MemCheck *MetricCheck
	PIDCheck *MetricCheck

	// static checks only below...
	ExistenceCheck *StaticCheck
	RunningCheck   *StaticCheck
	
	Templates	*TemplateConfig
}

// CheckMetrics checks everything where the Limit is not 0, there is no return because the
// checks modify the error in AlertdContainer
func (c *AlertdContainer) CheckMetrics(s *types.Stats, e error) {
	switch {
	case e != nil:
		c.AlertList.Add("Received an unknown error", "", e)
	default:
		if c.CPUCheck.Limit != nil {
			c.CheckCPUUsage(s)
		}
		if c.PIDCheck.Limit != nil {
			c.CheckMinPids(s)
		}
		if c.MemCheck.Limit != nil {
			c.CheckMemory(s)
		}
	}
}

// CheckStatics will run all of the static checks that are listed for a container.
func (c *AlertdContainer) CheckStatics(j *types.ContainerJSON, e error) {
	c.CheckExist(e)
	if j != nil && c.RunningCheck.Expected != nil {
		c.CheckRunning(j)
	}
}

// ChecksShouldStop returns whether the checks should stop after the static checks or
// continue onto the metric checks.
func (c *AlertdContainer) ChecksShouldStop() bool {
	switch {
	case c.RunningCheck.AlertActive:
		return true
	case c.ExistenceCheck.AlertActive:
		return true
	case c.RunningCheck.Expected != nil && !*c.RunningCheck.Expected:
		return true
	case c.AlertList.ShouldSend():
		return true
	default:
		return false
	}
}

// IsUnknown is a check that takes the error when the docker API is polled, it
// mathes part of the error string that is returned.
func (c *AlertdContainer) IsUnknown(err error) bool {
	if err != nil {
		return strings.Contains(err.Error(), "No such container:")
	}
	return false
}

// HasErrored returns true when the error is something other than `isUnknownContainer`
// which means that docker-alertd probably crashed.
func (c *AlertdContainer) HasErrored(e error) bool {
	return e != nil && !c.IsUnknown(e)
}

// HasBecomeKnown returns true if there is an active alert and error is nil, which means
// that the container check was successful and the 0 index check (existence check) cannot
// be active
func (c *AlertdContainer) HasBecomeKnown(e error) bool {
	return c.ExistenceCheck.AlertActive && e == nil
}

// CheckExist checks that the container exists, running or not
func (c *AlertdContainer) CheckExist(e error) {
	var message bytes.Buffer
	var title bytes.Buffer
	
	data := struct {
		Name		string
	}{
		c.Name,
	}
	
	switch {
	case c.IsUnknown(e) && !c.ExistenceCheck.AlertActive:
		if c.ShouldDelayStatic(true, c.ExistenceCheck) {
			return
		}
		
		// if the alert is not active I need to alert and make it active
		c.Templates.Executor.ExecuteTemplate(&message, "exist-failure-message", data)
		c.Templates.Executor.ExecuteTemplate(&title, "exist-failure-title", data)
		
		c.AlertList.Add(message.String(), title.String(), nil)
		
		c.ExistenceCheck.ToggleAlertActive()

	case c.IsUnknown(e) && c.ExistenceCheck.AlertActive:
		// do nothing
	case c.HasErrored(e):
		// if there is some other error besides an existence check error
		c.AlertList.Add(fmt.Sprintf("%s", c.Name), ErrUnknown.Error(), e)

	case c.HasBecomeKnown(e):
		c.ShouldDelayStatic(false, c.ExistenceCheck)
		
		c.Templates.Executor.ExecuteTemplate(&message, "exist-recovery-message", data)
		c.Templates.Executor.ExecuteTemplate(&title, "exist-recovery-title", data)
		
		c.AlertList.Add(message.String(), title.String(), nil)
		
		c.ExistenceCheck.ToggleAlertActive()
	default:
		return // nothing is wrong, just keep going
	}
}

// ShouldAlertRunning returns whether the running state is as expected
func (c *AlertdContainer) ShouldAlertRunning(j *types.ContainerJSON) bool {
	// if they are not equal, return true (send alert)
	return *c.RunningCheck.Expected != j.State.Running
}

// CheckRunning will check to see if the container is currently running or not
func (c *AlertdContainer) CheckRunning(j *types.ContainerJSON) {
	a := c.ShouldAlertRunning(j)
	
	if c.ShouldDelayStatic(a, c.RunningCheck) {
		return
	}
	
	var message bytes.Buffer
	var title bytes.Buffer
	
	data := struct {
		Name		string
		Expected	bool
		Running		bool
	}{
		c.Name,
		*c.RunningCheck.Expected,
		j.State.Running,
	}
	
	switch {
	case a && !c.RunningCheck.AlertActive:
		c.Templates.Executor.ExecuteTemplate(&message, "running-failure-message", data)
		c.Templates.Executor.ExecuteTemplate(&title, "running-failure-title", data)
		
		c.AlertList.Add(message.String(), title.String(), nil)

		c.RunningCheck.ToggleAlertActive()

	case !a && c.RunningCheck.AlertActive:
		c.Templates.Executor.ExecuteTemplate(&message, "running-recovery-message", data)
		c.Templates.Executor.ExecuteTemplate(&title, "running-recovery-title", data)
		
		c.AlertList.Add(message.String(), title.String(), nil)

		c.RunningCheck.ToggleAlertActive()
	}
}

// RealCPUUsage calculates the CPU usage based on the ContainerJSON info
func (c *AlertdContainer) RealCPUUsage(s *types.Stats) uint64 {
	totalUsage := float64(s.CPUStats.CPUUsage.TotalUsage)
	preTotalUsage := float64(s.PreCPUStats.CPUUsage.TotalUsage)
	systemCPUUsage := float64(s.CPUStats.SystemUsage)
	preSystemCPUUsage := float64(s.PreCPUStats.SystemUsage)

	u := (totalUsage - preTotalUsage) / (systemCPUUsage - preSystemCPUUsage) * 100
	return uint64(u)
}

// ShouldAlertCPU returns true if the limit is breached
func (c *AlertdContainer) ShouldAlertCPU(u uint64) bool {
	return u > *c.CPUCheck.Limit
}

// CheckCPUUsage takes care of sending the alerts if they are needed
func (c *AlertdContainer) CheckCPUUsage(s *types.Stats) {
	if c.CPUCheck.Limit == nil {
		return
	}
	
	u := c.RealCPUUsage(s)
	a := c.ShouldAlertCPU(u)
	
	if c.ShouldDelayMetric(a, c.CPUCheck) {
		return
	}
	
	var message bytes.Buffer
	var title bytes.Buffer
	
	data := struct {
		Name	string
		Limit	uint64
		Usage	uint64
	}{
		c.Name,
		*c.CPUCheck.Limit,
		u,
	}

	switch {
	case a && !c.CPUCheck.AlertActive:
		c.Templates.Executor.ExecuteTemplate(&message, "cpu-failure-message", data)
		c.Templates.Executor.ExecuteTemplate(&title, "cpu-failure-title", data)
		
		c.AlertList.Add(message.String(), title.String(), nil)

		c.CPUCheck.ToggleAlertActive()

	case !a && c.CPUCheck.AlertActive:
		c.Templates.Executor.ExecuteTemplate(&message, "cpu-recovery-message", data)
		c.Templates.Executor.ExecuteTemplate(&title, "cpu-recovery-title", data)
		
		c.AlertList.Add(message.String(), title.String(), nil)

		c.CPUCheck.ToggleAlertActive()
	}
}

// ShouldAlertMinPIDS returns true if the minPID check fails
func (c *AlertdContainer) ShouldAlertMinPIDS(s *types.Stats) bool {
	return s.PidsStats.Current < *c.PIDCheck.Limit
}

// CheckMinPids uses the min pids setting and check the number of PIDS in the container
// returns true if alerts should be sent, and also returns the amount of running pids.
func (c *AlertdContainer) CheckMinPids(s *types.Stats) {
	if c.PIDCheck.Limit == nil {
		return
	}
	
	a := c.ShouldAlertMinPIDS(s)
	
	if c.ShouldDelayMetric(a, c.PIDCheck) {
		return
	}
	
	var message bytes.Buffer
	var title bytes.Buffer
	
	data := struct {
		Name	string
		Limit	uint64
		Usage	uint64
	}{
		c.Name,
		*c.PIDCheck.Limit,
		s.PidsStats.Current,
	}
	
	switch {
	case a && !c.PIDCheck.AlertActive:
		c.Templates.Executor.ExecuteTemplate(&message, "min-pid-failure-message", data)
		c.Templates.Executor.ExecuteTemplate(&title, "min-pid-failure-title", data)
		
		c.AlertList.Add(message.String(), title.String(), nil)

		c.PIDCheck.ToggleAlertActive()

	case !a && c.PIDCheck.AlertActive:
		c.Templates.Executor.ExecuteTemplate(&message, "min-pid-recovery-message", data)
		c.Templates.Executor.ExecuteTemplate(&title, "min-pid-recovery-title", data)
		
		c.AlertList.Add(message.String(), title.String(), nil)

		c.PIDCheck.ToggleAlertActive()
	}
}

// MemUsageMB returns the memory usage in MB
func (c *AlertdContainer) MemUsageMB(s *types.Stats) uint64 {
	return s.MemoryStats.Usage / 1000000
}

// ShouldAlertMemory returns whether the memory limit has been exceeded
func (c *AlertdContainer) ShouldAlertMemory(s *types.Stats) bool {
	// Memory level in MB
	u := c.MemUsageMB(s)
	return u > *c.MemCheck.Limit
}

// CheckMemory checks the memory used by the container in MB, returns true if an
// error should be sent as well as the actual memory usage
func (c *AlertdContainer) CheckMemory(s *types.Stats) {
	if c.MemCheck.Limit == nil {
		return
	}
	
	a := c.ShouldAlertMemory(s)
	
	if c.ShouldDelayMetric(a, c.MemCheck) {
		return
	}
	
	var message bytes.Buffer
	var title bytes.Buffer
	
	data := struct {
		Name	string
		Limit	uint64
		Usage	uint64
	}{
		c.Name,
		*c.MemCheck.Limit,
		c.MemUsageMB(s),
	}

	if a && !c.MemCheck.AlertActive {
		c.Templates.Executor.ExecuteTemplate(&message, "memory-failure-message", data)
		c.Templates.Executor.ExecuteTemplate(&title, "memory-failure-title", data)
		
		c.AlertList.Add(message.String(), title.String(), nil)
		
		c.MemCheck.ToggleAlertActive()
		
	} else if !a && c.MemCheck.AlertActive {
		c.Templates.Executor.ExecuteTemplate(&message, "memory-recovery-message", data)
		c.Templates.Executor.ExecuteTemplate(&title, "memory-recovery-title", data)
		
		c.AlertList.Add(message.String(), title.String(), nil)

		c.MemCheck.ToggleAlertActive()
	}
}

func (c *AlertdContainer) ShouldDelayMetric(alert bool, metric *MetricCheck) bool {
	if metric.MinDelay != nil {
		if alert {
			if !metric.Delaying {
				metric.Delaying = true
				metric.DelaySince = time.Now()
				return true
			
			} else if uint64(time.Now().Sub(metric.DelaySince).Seconds()) < *metric.MinDelay {
				return true
			}
		
		} else {
			if metric.Delaying {
				metric.Delaying = false
			}
		}
	}
	
	return false
}

func (c *AlertdContainer) ShouldDelayStatic(alert bool, static *StaticCheck) bool {
	if static.MinDelay != nil {
		if alert {
			if !static.Delaying {
				static.Delaying = true
				static.DelaySince = time.Now()
				return true
			
			} else if uint64(time.Now().Sub(static.DelaySince).Seconds()) < *static.MinDelay {
				return true
			}
		
		} else {
			if static.Delaying {
				static.Delaying = false
			}
		}
	}
	
	return false
}