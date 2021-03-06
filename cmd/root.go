package cmd

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"

	"github.com/pkg/errors"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	// Config is the configuration object defined in conf-types file
	Config   Conf
	confName = ".docker-alertd"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "docker-alertd",
	Short: "docker-alertd: alert daemon for docker engine",
	Long: `docker-alerts parses a configuration file and then monitors containers through
the docker api.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {

		err := Config.Validate()
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}

		Start(&Config)
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports Persistent Flags, which, if defined here,
	// will be global for your application.

	RootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "",
		fmt.Sprintf("config file (default is ./%[1]s.yaml or $HOME/%[1]s.yaml)", confName))
	RootCmd.PersistentFlags().Uint64P("iterations", "i", 0,
		"the number of iterations that the monitor will run. (default 0 is infinite)")
	RootCmd.PersistentFlags().Uint64P("duration", "t", 1000,
		"the duration between monitor calls to the docker API in milliseconds (default 1000)")

	// Cobra also supports local flags, which will only run
	// Bind all the flags to viper for handling
	viper.BindPFlag("iterations", RootCmd.PersistentFlags().Lookup("iterations"))
	viper.BindPFlag("duration", RootCmd.PersistentFlags().Lookup("duration"))

	// local flags for when this action is called directly.
	//RootCmd.Flags().BoolVarP(&version, "version", "v", false, "Print `docker-alertd` version")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName(confName) // name of config file (without extension)
	// add working dir path above the home dir for the config file
	path, _ := os.Getwd()
	viper.AddConfigPath(path)
	viper.AddConfigPath(os.Getenv("HOME")) // adding home directory as first search path
	viper.AutomaticEnv()                   // read in environment variables that match

	// If a config file is found, read it in.
	err := viper.ReadInConfig()
	switch {
	case err != nil:
		log.Println(err)
	default:
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	err = viper.Unmarshal(&Config)
	if err != nil {
		log.Println(errors.Wrap(err, "unmarshal config file"))
	}
}

// Container gets data from the Unmarshaling of the configuration file JSON and stores
// the data throughout the course of the monitor.
type Container struct {
	Name            string
	MaxCPU          *uint64
	MaxMem          *uint64
	MinProcs        *uint64
	ExpectedRunning *bool
	Delay			*uint64
}

// Conf struct that combines containers and email settings structs
type Conf struct {
	Containers []Container
	Email      Email
	Slack      Slack
	Pushover   Pushover
	Pushbullet Pushbullet
	Iterations uint64
	Duration   uint64
	Alerters   []Alerter
	Templates  TemplateConfig
}

// ValidateEmailSettings calls valid on the Email settings and adds them to the alerters
// if everything is ok
func (c *Conf) ValidateEmailSettings() error {
	err := c.Email.Valid()
	switch {
	case reflect.DeepEqual(Email{}, c.Email):
		return nil
	case err != nil:
		return err
	default:
		c.Alerters = append(c.Alerters, c.Email)
		log.Println("email alerts active")
		return nil
	}
}

// ValidateSlackSettings validates slack settings and adds it to the alerters
func (c *Conf) ValidateSlackSettings() error {
	err := c.Slack.Valid()
	switch {
	case reflect.DeepEqual(Slack{}, c.Slack):
		return nil // assume that slack was omitted and not wanted
	case err != nil:
		return err
	default:
		c.Alerters = append(c.Alerters, c.Slack)
		log.Println("slack alerts active")
		return nil
	}
}

// ValidatePushoverSettings validates pushover settings and adds it to the alerters
func (c *Conf) ValidatePushoverSettings() error {
	err := c.Pushover.Valid()
	switch {
	case reflect.DeepEqual(Pushover{}, c.Pushover):
		return nil // assume that pushover was omitted and not wanted
	case err != nil:
		return err
	default:
		c.Alerters = append(c.Alerters, c.Pushover)
		log.Println("pushover alerts active")
		return nil
	}
}

// ValidatePushbulletSettings validates pushover settings and adds it to the alerters
func (c *Conf) ValidatePushbulletSettings() error {
	err := c.Pushbullet.Valid()
	switch {
	case reflect.DeepEqual(Pushbullet{}, c.Pushbullet):
		return nil // assume that pushover was omitted and not wanted
	case err != nil:
		return err
	default:
		c.Alerters = append(c.Alerters, c.Pushbullet)
		log.Println("pushbullet alerts active")
		return nil
	}
}

func (c *Conf) ValidateTemplatesSettings() error {
	var err error
	
	c.Templates, err = c.Templates.Build()
	
	switch {
	case err != nil:
		return err
	default:
		return nil
	}
}

// Validate validates the configuration that was passed in
func (c *Conf) Validate() error {
	// the error to wrap and return at the end
	errString := []string{}

	if reflect.DeepEqual(&Conf{}, c) {
		errString = append(errString, ErrEmptyConfig.Error())
	}

	if len(c.Containers) < 1 {
		errString = append(errString, ErrNoContainers.Error())
	}

	if err := c.ValidateEmailSettings(); err != nil {
		errString = append(errString, err.Error())
	}

	if err := c.ValidateSlackSettings(); err != nil {
		errString = append(errString, err.Error())
	}

	if err := c.ValidatePushoverSettings(); err != nil {
		errString = append(errString, err.Error())
	}
	
	if err := c.ValidatePushbulletSettings(); err != nil {
		errString = append(errString, err.Error())
	}
	
	if err := c.ValidateTemplatesSettings(); err != nil {
		errString = append(errString, err.Error())
	}
	
	// if the length of the string of errors is 0 then everything has completed
	// successfully and everything is valid.
	if len(errString) == 0 {
		return nil
	}

	delimErr := strings.Join(errString, ", ")
	err := errors.New(delimErr)

	return errors.Wrap(err, "config validation fail")
}
