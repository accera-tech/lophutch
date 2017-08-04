package hutch

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
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

	if res.StatusCode != 200 {
		return errors.Errorcf(map[string]interface{}{
			"Request":  req,
			"Response": res,
		}, "HTTP request returned an unexpected response, %d %s", res.StatusCode, res.Status)
	}

	content := Queue{}
	if err := json.NewDecoder(res.Body).Decode(&content); err != nil {
		return errors.Wrap(err, "failed to parse the response body")
	}

	if content.Messages > rule.Limit {
		log.Printf("Definition: \"%s\"; Rule: \"%s\"; Broken. Executing defined actions...", def.Name, rule.Name)

		for _, action := range rule.Actions {
			cmd := exec.Command(action.Cmd, action.Args...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				log.Printf("Definition: \"%s\"; Rule: \"%s\"; Broken. Executing defined actions... FAIL", def.Name, rule.Name)
				return errors.Wrapc(err, map[string]interface{}{
					"Action": action.Name,
				}, "failed to execute action")
			}
		}

		log.Printf("Definition: \"%s\"; Rule: \"%s\"; Broken. Executing defined actions... OK", def.Name, rule.Name)
	} else {
		log.Printf("Definition: \"%s\"; Rule: \"%s\"; Good", def.Name, rule.Name)
	}

	return nil
}

type Queue struct {
	Messages int
}
