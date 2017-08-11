package hutch

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/robertkrimen/otto"
	"github.com/spf13/viper"
	"github.com/tradeforce/lophutch/common"
	"github.com/zignd/errors"
)

func Schedule(done <-chan struct{}) error {
	ticker := time.NewTicker(viper.GetDuration("delay") * time.Millisecond)
	delays := make(map[string]time.Time)
	for {
		select {
		case <-ticker.C:
			if err := Scout(delays); err != nil {
				return errors.Wrap(err, "a scout failed")
			}
		case <-done:
			ticker.Stop()
			break
		}
	}
}

func Scout(delays map[string]time.Time) error {
	servers, err := getServers()
	if err != nil {
		return errors.Wrap(err, "failed to retrieve the configured servers")
	}
	for _, server := range servers {
		for _, rule := range server.Rules {
			log.Printf("Server: %s | Rule: %s | Processing...", server.Description, rule.Description)
			if err := processRule(server, rule, delays); err != nil {
				err = errors.Wrapcf(err, map[string]interface{}{
					"server": server.Description,
					"rule":   rule.Description,
				}, "failed to process rule %s", rule.Description)
				log.Printf("Server: %s | Rule: %s | Processing... Fail - %s", server.Description, rule.Description, err.Error())
				continue
			}
			log.Printf("Server: %s | Rule: %s | Processing... OK", server.Description, rule.Description)
		}
	}

	return nil
}

func getServers() ([]common.Server, error) {
	var servers []common.Server
	if err := viper.UnmarshalKey("Servers", &servers); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal the `Servers` setting")
	}

	if len(servers) == 0 {
		return nil, errors.New("`Servers` setting has no servers")
	}

	// TODO: Add more validations

	return servers, nil
}

func processRule(server common.Server, rule common.Rule, delays map[string]time.Time) error {
	bodyStr, err := performRequest(server, rule.Request)
	if err != nil {
		return errors.Wrap(err, "failed to perform the configured HTTP request")
	}

	log.Printf("Server: %s | Rule: %s | Evaluating rule...", server.Description, rule.Description)
	result, err := evaluateRule(rule.Evaluator, bodyStr)
	if err != nil {
		return errors.Wrapc(err, map[string]interface{}{
			"evaluator": rule.Evaluator,
			"body":      bodyStr,
		}, "failed to evaluate rule")
	}

	if !result {
		log.Printf("Server: %s | Rule: %s | Evaluated to false", server.Description, rule.Description)
		return nil
	}

	log.Printf("Server: %s | Rule: %s | Evaluated to true", server.Description, rule.Description)

	delay, ok := delays[rule.ID]
	if ok {
		if delay.Before(time.Now()) {
			delete(delays, rule.ID)
		} else {
			log.Printf("Server: %s | Rule: %s | Delayed", server.Description, rule.Description)
			return nil
		}
	}

	log.Printf("Server: %s | Rule: %s | Executing actions...", server.Description, rule.Description)

	for _, action := range rule.Actions {
		log.Printf("Server: %s | Rule: %s | Executing action %s...", server.Description, rule.Description, action.Description)
		if err := act(action); err != nil {
			err = errors.Wrapcf(err, map[string]interface{}{
				"action": action,
			}, "failed to execute action %s", action.Description)
			log.Printf("Server: %s | Rule: %s | Executing action %s... Fail - %s", server.Description, rule.Description, action.Description, err.Error())
			break
		}
		log.Printf("Server: %s | Rule: %s | Executing action %s... OK", server.Description, rule.Description, action.Description)
	}

	log.Printf("Server: %s | Rule: %s | Executing actions... OK", server.Description, rule.Description)

	delays[rule.ID] = time.Now().Add(rule.Delay * time.Millisecond)

	return nil
}

func performRequest(server common.Server, request common.Request) (string, error) {
	urlStr := fmt.Sprintf("%s://%s:%d%s", server.Protocol, server.Host, server.Port, request.Path)
	req, err := http.NewRequest(request.Method, urlStr, nil)
	if err != nil {
		return "", errors.Wrapf(err, "failed to create an HTTP request to %s", urlStr)
	}
	req.SetBasicAuth(server.User, server.Password)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", errors.Wrapf(err, "failed to perform an HTTP %s request to %s", request.Method, request.Path)
	}
	defer func() {
		res.Body.Close()
	}()

	if res.StatusCode != 200 {
		return "", errors.Errorcf(map[string]interface{}{
			"request":  req,
			"response": res,
		}, "HTTP request returned an unexpected response, %s", res.Status)
	}

	buf := bytes.Buffer{}
	if _, err := buf.ReadFrom(res.Body); err != nil {
		return "", errors.Wrap(err, "failed to read the response body and append it to a buffer")
	}

	return buf.String(), nil
}

func evaluateRule(evaluator string, bodyStr string) (bool, error) {
	vm := otto.New()
	if err := vm.Set("_body", bodyStr); err != nil {
		return false, errors.Wrap(err, "failed to set the body variable")
	}

	script := fmt.Sprintf(`
		%s
		_result = evaluate(JSON.parse(_body));
	`, evaluator)

	_, err := vm.Run(script)
	if err != nil {
		return false, errors.Wrapc(err, map[string]interface{}{
			"script": script,
		}, "failed to run the script")
	}

	v, err := vm.Get("_result")
	if err != nil {
		return false, errors.Wrap(err, "failed retrieve the _result variable")
	}

	if !v.IsBoolean() {
		return false, errors.Wrap(err, "_result is not a boolean")
	}

	result, err := v.ToBoolean()
	if err != nil {
		return false, errors.Wrap(err, "failed to convert _result to bool")
	}

	return result, nil
}

func act(action common.Action) error {
	cmd := exec.Command(action.Cmd, action.Args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return errors.Wrapc(err, map[string]interface{}{
			"command": cmd.Args,
		}, "failed to execute command")
	}
	return nil
}
