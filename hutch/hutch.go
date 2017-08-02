package hutch

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/viper"
	"github.com/tradeforce/lophutch/common"
	"github.com/zignd/errors"
)

func Schedule(done <-chan struct{}) error {
	ticker := time.NewTicker(viper.GetDuration("Frequency") * time.Second)
	for {
		select {
		case <-ticker.C:
			if err := Scout(); err != nil {
				return errors.Wrap(err, "a scout failed")
			}
		case <-done:
			ticker.Stop()
			break
		}
	}
}

func Scout() error {
	var defs []common.Definition
	if err := viper.UnmarshalKey("Definitions", &defs); err != nil {
		return errors.Wrap(err, "failed to unmarshal the `Definitions` setting")
	}

	if len(defs) == 0 {
		return errors.New("no `Definitions` to scout")
	}

	for _, def := range defs {
		for _, rule := range def.Rules {
			if err := ProcessRule(def, rule); err != nil {
				return errors.Wrapc(err, map[string]interface{}{
					"Definition": def.Name,
					"Rule":       rule.Name,
				}, "failed to process rule")
			}
		}
	}

	return nil
}

func ProcessRule(def common.Definition, rule common.Rule) error {
	urlStr := fmt.Sprintf("%s://%s:%d%s", def.Protocol, def.Host, def.Port, rule.Path)
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return errors.Wrapf(err, "failed to create an HTTP request to %s", urlStr)
	}
	req.SetBasicAuth(def.User, def.Password)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrapf(err, "failed to perform an HTTP request to %s", rule.Path)
	}
	defer func() {
		res.Body.Close()
	}()

	v := map[string]interface{}{}
	if err := json.NewDecoder(res.Body).Decode(&v); err != nil {
		return errors.Wrap(err, "failed to parse the response body")
	}
	fmt.Println(v)
	return nil
}
