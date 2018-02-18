package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func uint64P(u uint64) *uint64 {
	p := u
	return &p
}

// intP returns a pointer to an int
func int64P(i int64) *int64 {
	p := i
	return &p
}

// boolP returns a pointer to a bool
func boolP(b bool) *bool {
	p := b
	return &p
}

// GetStats just uses the docker API and an already tested Unmarshal function, no
// testing needed.
func GetStats(a *AlertdContainer, c *client.Client) (*types.Stats, error) {
	cs, err := c.ContainerStats(context.Background(), a.Name, false)
	if err != nil {
		return nil, err
	}
	defer cs.Body.Close()

	d := json.NewDecoder(cs.Body)
	d.UseNumber()

	var stats types.Stats
	if err := d.Decode(&stats); err != nil {
		return nil, err
	}

	return &stats, nil
}

// ContainerInspect returns the information which can decide if the container is current;y running
// or not.
func ContainerInspect(a *AlertdContainer, c *client.Client) (*types.ContainerJSON, error) {
	containerJSON, err := c.ContainerInspect(context.Background(), a.Name)
	if err != nil {
		return nil, err
	}

	return &containerJSON, nil
}

// InitCheckers returns a slice of containers with all the info needed to run a
// check on the container. Active is for whether or not the alert is active, not the check
func InitCheckers(c *Conf) []AlertdContainer {
	// Taking the values from the conf and adding them into the AlertdContainers
	var containers []AlertdContainer
	for _, v := range c.Containers {
		containers = append(containers, AlertdContainer{
			Name: v.Name,
			AlertList: &AlertList{
				Alerts: []Alert{},
			},
			CPUCheck: &MetricCheck{
				Limit:			v.MaxCPU,
				AlertActive:	false,
				MinDelay:		v.Delay,
				Delaying:		false,
				DelaySince:		time.Now(),
			},
			MemCheck: &MetricCheck{
				Limit:       v.MaxMem,
				AlertActive: false,
				MinDelay:		v.Delay,
				Delaying:		false,
				DelaySince:		time.Now(),
			},
			PIDCheck: &MetricCheck{
				Limit:       v.MinProcs,
				AlertActive: false,
				MinDelay:		v.Delay,
				Delaying:		false,
				DelaySince:		time.Now(),
			},
			ExistenceCheck: &StaticCheck{
				Expected:    boolP(true),
				AlertActive: false,
				MinDelay:		v.Delay,
				Delaying:		false,
				DelaySince:		time.Now(),
			},
			RunningCheck: &StaticCheck{
				Expected:    v.ExpectedRunning,
				AlertActive: false,
				MinDelay:		v.Delay,
				Delaying:		false,
				DelaySince:		time.Now(),
			},
			Templates: &c.Templates,
		})
	}
	return containers
}

// CheckContainers goes through and checks all the containers in a loop
func CheckContainers(cnt []AlertdContainer, cli *client.Client, a *AlertList) {
	for _, c := range cnt {
		// make sure we have a clean alert for this loop
		c.AlertList.Clear()

		// handling whether the container exists, if these checks fail, the checking
		// process should stop
		j, err := ContainerInspect(&c, cli)
		c.CheckStatics(j, err)

		// if an alert should be sent that means it either failed existence or running
		// checks which means that nothing more can be checked
		if c.ChecksShouldStop() {
			a.Concat(c.AlertList) // add the alert in the container to the main alert
			continue
		}

		s, err := GetStats(&c, cli)
		c.CheckMetrics(s, err)

		if c.AlertList.ShouldSend() {
			a.Concat(c.AlertList)
		}
	}
}

// Monitor contains all the calls for the main loop of the monitor
func Monitor(c *Conf, a *AlertList) {
	cli, err := client.NewEnvClient()
	if err != nil {
		log.Fatal(err)
	}

	cnt := InitCheckers(c)

	switch c.Iterations {
	case 0:
		for {
			a.Clear()
			CheckContainers(cnt, cli, a)
			a.Evaluate()
			time.Sleep(time.Duration(c.Duration) * time.Millisecond)
		}
	default:
		for i := uint64(0); i < c.Iterations; i++ {
			a.Clear()
			CheckContainers(cnt, cli, a)
			a.Evaluate()
			time.Sleep(time.Duration(c.Duration) * time.Millisecond)
		}
	}
}

func AlertStarting(c *Conf, a *AlertList) {
	var message bytes.Buffer
	var title bytes.Buffer
	
	var data = struct{}{}
	
	c.Templates.Executor.ExecuteTemplate(&message, "starting-message", data)
	c.Templates.Executor.ExecuteTemplate(&title, "starting-title", data)
	
	a.Add(message.String(), title.String(), nil)
	
	a.Evaluate()
}

func AlertStopping(c *Conf, a *AlertList) {
	shutdown := make(chan os.Signal)
    
	signal.Notify(shutdown)
	
	go func() {
		<-shutdown
		
		var message bytes.Buffer
		var title bytes.Buffer
		
		var data = struct{}{}
		
		c.Templates.Executor.ExecuteTemplate(&message, "stopping-message", data)
		c.Templates.Executor.ExecuteTemplate(&title, "stopping-title", data)
		
		a.Add(message.String(), title.String(), nil)
		
		a.Evaluate()
		
		time.Sleep(time.Duration(5000) * time.Millisecond)
		
		os.Exit(1)
	}()
}

// Start the main monitor loop for a set amount of iterations
func Start(c *Conf) {
	log.Printf("starting docker-alertd\n------------------------------")
	a := &AlertList{Alerts: []Alert{}}
	
	AlertStarting(c, a)
	AlertStopping(c, a)
	
	time.Sleep(time.Duration(c.Duration) * time.Millisecond)
	
	Monitor(c, a)
}