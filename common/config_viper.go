package common

import (
	"os"
	"path"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/zignd/errors"
)

// ConfigFlags configures the application flags.
func ConfigFlags() error {
	pflag.String("config-file", "$XDG_CONFIG_HOME/config.yaml", "Configuration file with information regarding the connection to the server, expressions and actions to be taken.")
	pflag.Bool("run-once", false, "Performs the verifications defined in the configuration file only once. When set to `true` the `Frequency` setting is ignored.")
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	configFile := viper.GetString("config-file")
	if configFile == "$XDG_CONFIG_HOME/config.yaml" {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		if xdgConfigHome := os.Getenv("XDG_CONFIG_HOME"); xdgConfigHome != "" {
			viper.AddConfigPath(path.Join(xdgConfigHome, "lophutch"))
		} else {
			viper.AddConfigPath("$HOME/.config/lophutch")
		}
	} else {
		if fi, err := os.Stat(configFile); (fi != nil && fi.IsDir()) || os.IsNotExist(err) {
			return errors.Wrapf(err, "could not find file %s", configFile)
		}
		viper.SetConfigFile(configFile)
	}

	viper.ReadInConfig()

	return nil
}
