# lophutch
RabbitMQ monitoring service

Lophutch is a flexible RabbitMQ monitoring service. It constantly pulls informations from a server through the RabbitMQ Management HTTP API and executes actions based on a user defined JavaScript functions that will decide if actions should be taken or not.

## Installation

It's a Go application and can be installed through:

    go get -u github.com/tradeforce/lophutch

## Usage

You can check the available execution options providing the `--help` flag but for completion those are the available ones:

* `--run-once`
* `--config-file`

## Configuration file sample

```yaml
---
delay: 1000
servers:
- description: main server
  protocol: http
  host: localhost
  port: 15672
  user: guest
  password: guest
  rules:
  - id: rule-1
    description: /lophutch/test1 should be properly consumed
    request:
      method: GET
      path: /api/queues/lophutch/test1
    evaluator: |-
      function evaluate(body) {
        if (body.messages_ready > 10)
          return true;
        return false;
      }
    delay: 30000
    actions:
    - description: notify via Slack
      cmd: send-msg-slack
      args:
      - "--channel"
      - "#critical"
      - "--message"
      - "/lophutch/test1 is not being properly consumed, running a new container"
    - description: run a new container
      cmd: run-container
      args:
      - "--image"
      - "test1"
```
