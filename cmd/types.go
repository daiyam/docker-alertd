package cmd

import (
	"log"
	"strings"
)

type Alert struct {
	Message	string
	Title	string
	Error	error
}

func (a *Alert) Log() {
	if a.Title != "" {
		log.Println(a.Title, "-", strings.Replace(a.Message, "\n", " ", -1))
	} else {
		log.Println(strings.Replace(a.Message, "\n", " ", -1))
	}
}

func (a *Alert) Dump() (s string) {
	if a.Title != "" {
		s += a.Title + " - "
	}
	
	s += strings.Replace(a.Message, "\n", " ", -1)
	
	if a.Error != nil {
		s += " - " + a.Error.Error()
	}
	
	return s
}

func (a *Alert) DumpEmail() (s string) {
	if a.Title != "" {
		s += a.Title + "\n"
	}
	
	s += a.Message
	
	if a.Error != nil {
		s += "\n" + a.Error.Error()
	}
	
	return s
}

// AlertList is the struct that stores information about alerts and its methods satisfy the
// Alerter interface
type AlertList struct {
	Alerts        []Alert
}

// ShouldSend returns true if there is an alert message to be sent
func (a *AlertList) ShouldSend() bool {
	return len(a.Alerts) > 0
}

// Evaluate will check if error should be sent and then trigger it if necessary
func (a *AlertList) Evaluate() {
	if a.ShouldSend() {
		a.Send(Config.Alerters)
	}
}

// Len returns the length of the alert message strings
func (a *AlertList) Len() int {
	return len(a.Alerts)
}

// Add should take in an error and wrap it
func (a *AlertList) Add(message string, title string, e error) {
	a.Alerts = append(a.Alerts, Alert{Message: message, Title: title, Error: e})
}

// Concat will concat different alerts from containers together into one
func (a *AlertList) Concat(b ...*AlertList) {
	for _, v := range b {
		for _, alert := range v.Alerts {
			a.Alerts = append(a.Alerts, alert)
		}
	}
}

// Log prints the alert to the log
func (a *AlertList) Log() {
	log.Println("ALERT:")
	for _, alert := range a.Alerts {
		alert.Log()
	}
}

// Clear will reset the alert to an empty string
func (a *AlertList) Clear() {
	a.Alerts = []Alert{}
}

// Dump takes the slice of alerts and dumps them to a single string
func (a *AlertList) Dump() (s string) {
	for _, alert := range a.Alerts {
		s += alert.Dump() + "\n\n"
	}
	
	return s
}

// DumpEmail behaves like dump, but formats them for email by splitting on ":" and adding
// \n\t (newline and tab) for the first two segments and joining the last segment. This
// should result in an email that is formatted as follows...
// [containerName]:
// 		[alertName]:
// 		Error: [errString]
func (a *AlertList) DumpEmail() (s string) {
	for _, alert := range a.Alerts {
		s += alert.DumpEmail() + "\n\n"
	}
	
	return s
}

func (a *AlertList) Message() (s string) {
	for _, alert := range a.Alerts {
		s += alert.Message + " "
	}
	
	return s
}

func (a *AlertList) Title() (s string) {
	for _, alert := range a.Alerts {
		s += alert.Title + " "
	}
	
	return s
}

// Send is for sending out alerts to syslog and to alerts that are active in conf
func (a *AlertList) Send(b []Alerter) {
	a.Log()
	
	for i := range b {
		go func(c Alerter) {
			err := c.Alert(a)
			if err != nil {
				log.Println(err)
			}
		}(b[i])
	}
}
