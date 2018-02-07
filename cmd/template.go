package cmd

import (
	"text/template"
)

type AlertTemplate struct {
	Title	string
	Message	string
}

type TemplateConfig struct {
	ExistFailure		AlertTemplate
	ExistRecovery		AlertTemplate
	RunningFailure		AlertTemplate
	RunningRecovery		AlertTemplate
	CPUFailure			AlertTemplate
	CPURecovery			AlertTemplate
	MinPIDFailure		AlertTemplate
	MinPIDRecovery		AlertTemplate
	MemoryFailure		AlertTemplate
	MemoryRecovery		AlertTemplate
	Executor			template.Template
}

func (t TemplateConfig) Build() (TemplateConfig, error) {
	var err error
	
	// {{{ Exist
	if t.ExistFailure.Message == "" {
		_, err = t.Executor.New("exist-failure-message").Parse("{{.Name}}")
	} else {
		_, err = t.Executor.New("exist-failure-message").Parse(t.ExistFailure.Message)
	}
	if err != nil {
		return t, err
	}
	
	if t.ExistFailure.Title == "" {
		_, err = t.Executor.New("exist-failure-title").Parse(ErrExistCheckFail.Error())
	} else {
		_, err = t.Executor.New("exist-failure-title").Parse(t.ExistFailure.Title)
	}
	if err != nil {
		return t, err
	}
	
	if t.ExistRecovery.Message == "" {
		_, err = t.Executor.New("exist-recovery-message").Parse("{{.Name}}")
	} else {
		_, err = t.Executor.New("exist-recovery-message").Parse(t.ExistRecovery.Message)
	}
	if err != nil {
		return t, err
	}
	
	if t.ExistRecovery.Title == "" {
		_, err = t.Executor.New("exist-recovery-title").Parse(ErrExistCheckRecovered.Error())
	} else {
		_, err = t.Executor.New("exist-recovery-title").Parse(t.ExistRecovery.Title)
	}
	if err != nil {
		return t, err
	}
	// }}}
	
	// {{{ Running
	if t.RunningFailure.Message == "" {
		_, err = t.Executor.New("running-failure-message").Parse("{{.Name}}: expected running state: {{.Expected}}, current running state: {{.Running}}")
	} else {
		_, err = t.Executor.New("running-failure-message").Parse(t.RunningFailure.Message)
	}
	if err != nil {
		return t, err
	}
	
	if t.RunningFailure.Title == "" {
		_, err = t.Executor.New("running-failure-title").Parse(ErrRunningCheckFail.Error())
	} else {
		_, err = t.Executor.New("running-failure-title").Parse(t.RunningFailure.Title)
	}
	if err != nil {
		return t, err
	}
	
	if t.RunningRecovery.Message == "" {
		_, err = t.Executor.New("running-recovery-message").Parse("{{.Name}}: expected running state: {{.Expected}}, current running state: {{.Running}}")
	} else {
		_, err = t.Executor.New("running-recovery-message").Parse(t.RunningRecovery.Message)
	}
	if err != nil {
		return t, err
	}
	
	if t.RunningRecovery.Title == "" {
		_, err = t.Executor.New("running-recovery-title").Parse(ErrRunningCheckRecovered.Error())
	} else {
		_, err = t.Executor.New("running-recovery-title").Parse(t.RunningRecovery.Title)
	}
	if err != nil {
		return t, err
	}
	// }}}
	
	// {{{ CPU
	if t.CPUFailure.Message == "" {
		_, err = t.Executor.New("cpu-failure-message").Parse("{{.Name}}: CPU limit: {{.Limit}}, current usage: {{.Usage}}")
	} else {
		_, err = t.Executor.New("cpu-failure-message").Parse(t.CPUFailure.Message)
	}
	if err != nil {
		return t, err
	}
	
	if t.CPUFailure.Title == "" {
		_, err = t.Executor.New("cpu-failure-title").Parse(ErrCPUCheckFail.Error())
	} else {
		_, err = t.Executor.New("cpu-failure-title").Parse(t.CPUFailure.Title)
	}
	if err != nil {
		return t, err
	}
	
	if t.CPURecovery.Message == "" {
		_, err = t.Executor.New("cpu-recovery-message").Parse("{{.Name}}: CPU limit: {{.Limit}}, current usage: {{.Usage}}")
	} else {
		_, err = t.Executor.New("cpu-recovery-message").Parse(t.CPURecovery.Message)
	}
	if err != nil {
		return t, err
	}
	
	if t.CPURecovery.Title == "" {
		_, err = t.Executor.New("cpu-recovery-title").Parse(ErrCPUCheckRecovered.Error())
	} else {
		_, err = t.Executor.New("cpu-recovery-title").Parse(t.CPURecovery.Title)
	}
	if err != nil {
		return t, err
	}
	// }}}
	
	// {{{ MinPID
	if t.MinPIDFailure.Message == "" {
		_, err = t.Executor.New("min-pid-failure-message").Parse("{{.Name}}: minimum PIDs: {{.Limit}}, current PIDs: {{.Usage}}")
	} else {
		_, err = t.Executor.New("min-pid-failure-message").Parse(t.MinPIDFailure.Message)
	}
	if err != nil {
		return t, err
	}
	
	if t.MinPIDFailure.Title == "" {
		_, err = t.Executor.New("min-pid-failure-title").Parse(ErrMinPIDCheckFail.Error())
	} else {
		_, err = t.Executor.New("min-pid-failure-title").Parse(t.MinPIDFailure.Title)
	}
	if err != nil {
		return t, err
	}
	
	if t.MinPIDRecovery.Message == "" {
		_, err = t.Executor.New("min-pid-recovery-message").Parse("{{.Name}}: minimum PIDs: {{.Limit}}, current PIDs: {{.Usage}}")
	} else {
		_, err = t.Executor.New("min-pid-recovery-message").Parse(t.MinPIDRecovery.Message)
	}
	if err != nil {
		return t, err
	}
	
	if t.MinPIDRecovery.Title == "" {
		_, err = t.Executor.New("min-pid-recovery-title").Parse(ErrMinPIDCheckRecovered.Error())
	} else {
		_, err = t.Executor.New("min-pid-recovery-title").Parse(t.MinPIDRecovery.Title)
	}
	if err != nil {
		return t, err
	}
	// }}}
	
	// {{{ Memory
	if t.MemoryFailure.Message == "" {
		_, err = t.Executor.New("memory-failure-message").Parse("{{.Name}}: Memory limit: {{.Limit}}, current usage: {{.Usage}}")
	} else {
		_, err = t.Executor.New("memory-failure-message").Parse(t.MemoryFailure.Message)
	}
	if err != nil {
		return t, err
	}
	
	if t.MemoryFailure.Title == "" {
		_, err = t.Executor.New("memory-failure-title").Parse(ErrMemCheckFail.Error())
	} else {
		_, err = t.Executor.New("memory-failure-title").Parse(t.MemoryFailure.Title)
	}
	if err != nil {
		return t, err
	}
	
	if t.MemoryRecovery.Message == "" {
		_, err = t.Executor.New("memory-recovery-message").Parse("{{.Name}}: Memory limit: {{.Limit}}, current usage: {{.Usage}}")
	} else {
		_, err = t.Executor.New("memory-recovery-message").Parse(t.MemoryRecovery.Message)
	}
	if err != nil {
		return t, err
	}
	
	if t.MemoryRecovery.Title == "" {
		_, err = t.Executor.New("memory-recovery-title").Parse(ErrMemCheckRecovered.Error())
	} else {
		_, err = t.Executor.New("memory-recovery-title").Parse(t.MemoryRecovery.Title)
	}
	if err != nil {
		return t, err
	}
	// }}}
	
	return t, nil
}