# Project Description

This is a POC Go program to test some ideas inspired by
- [Designing Event Driven Systems by Ben Stopford](https://www.confluent.io/designing-event-driven-systems/)
- [A pratical talk on the subject by Neil Buesing](https://www.confluent.io/resources/kafka-summit-2020/synchronous-commands-over-apache-kafka/)

# Usage

1. build the go web server and run a couple of instances of each.
2. send them requests, and watch the logs. you will get a uuid for each
3. use a tool like [kcat](https://github.com/edenhill/kcat) to "respond" to requests

Useful snippets from my shell history

Using kcat to respond:
``` sh
echo "hello there" | kcat -v -b localhost:9092 -P -t responses -H "request-id=6c50571c-1224-4e2f-baff-ee077c984158"
```

Using kcat to monitor requests topic:

``` sh
kcat \
      -b localhost:9092 \
      -t requests-0 -C \
      -f '\nKey (%K bytes): %k
  Value (%S bytes): %s
  Timestamp: %T
  Partition: %p
  Offset: %o
  Headers: %h\n'
```

# Conclusion

The POC works. I haven't benchmarked it but I think I can handle a lot a troughput.
Ideally, we would scale the app vertically as much as we can because having
multiple copies of the app runnings means each instance will be processing messages meant
for other instances.

# Ideas for improvement

- To alleviate allocation on the blocker we could use a [sync.Pool](https://pkg.go.dev/sync#Pool)
