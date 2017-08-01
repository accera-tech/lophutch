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
	pflag.String("config-file", "$XDG_CONFIG_HOME/config.json", "configuration file with information regarding the connection to the server, expressions and actions to be taken")
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	configFile := viper.GetString("config-file")
	if configFile == "$XDG_CONFIG_HOME/config.json" {
		println(1)
		viper.SetConfigName("config")
		viper.SetConfigType("json")
		if xdgConfigHome := os.Getenv("XDG_CONFIG_HOME"); xdgConfigHome != "" {
			viper.AddConfigPath(path.Join(xdgConfigHome, "lophutch"))
		} else {
			viper.AddConfigPath("$HOME/.config/lophutch")
		}
	} else {
		if fi, err := os.Stat(configFile); fi.IsDir() || os.IsNotExist(err) {
			return errors.Wrap(errors.Wrapf(err, "could not find file %s", configFile), "test")
		}
		viper.SetConfigFile(configFile)
	}

	return nil
}
