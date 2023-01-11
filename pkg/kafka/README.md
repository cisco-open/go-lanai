# Kafka

The Kafka module provides an abstraction for interfacing with Kafaka so that application code can focus on writing message processing code.

## Binder

The Binder is the main interface for working with Kafka. The Kafka module provides a ```Kafka.Binder``` interface which your application can
inject. Once your application have a reference to the Binder interface, you can create message producer from the Binder, or add your message 
consumer/subscriber to the binder. 

## Example

1. To activate the Kafka module

```go 
Kafka.Use()
```

2. Add Kafka properties to application.yml

```yaml
# Following configuration serve as an example
# values specified in `kafka.bindings.default.*` are same as hardcoded defaults
#
# To overwrite defaults, add section with prefix `kafka.bindings.<your binding name>`,
# and specify the binding name when using Binder with `BindingName(...)` option
kafka:
  bindings:
    default:
      producer:
        log-level: "debug"
        ack-mode: "local" # all, local or none
        ack-timeout: 10s
        max-retry: 3
        backoff-interval: 100ms
        provisioning:
          auto-create-topic: true
          auto-add-partitions: true
          allow-lower-partitions: true
          partition-count: 1
          replication-factor: 1
      consumer:
        log-level: "debug"
        join-timeout: 60s
        max-retry: 4
        backoff-interval: 2s
    binding-name:
      producer:
        ...
      consumer:
        ...
```

4. Inject the ```Kafka.Binder``` into your application

```go		
fx.Provide(NewComponent)
```

To create a producer from a ```Binder```.

```go
func NewComponent(b kafka.Binder) (*MyComponent, error) {
	p, err := b.Produce("MY_TOPIC", kafka.BindingName("my-binding-name"))
	if err != nil {
		return nil, err
	}
	return &MyComponent{Producer: p}, nil
}
```

Here you will have a component that have a reference to a message producer. The ```BindingName``` option allows binding specific configuration
to be applied to your producer. See the documentation on ```BindingName``` for more details.

To add a consumer to the ```Binder```, use ```fx.Invoke``` to registers the functions so that it's executed eagerly on application start.
See fx documentation for the difference between ```fx.Invoke``` and ```fx.Provide```.  

```go
fx.Invoke(AddConsumer)
```

```go
func AddConsumer(Binder kafka.Binder) error {
	mc := &MyConsumer{
	}
	consumer, e := di.Binder.Consume("MY_TOPIC", kafkaGroup, kafka.BindingName("my-binding-name"))
	if e != nil {
		return e
	}
	if e := consumer.AddHandler(mc.MyMessageHandler); e != nil {
		return e
	}
	return nil
}
```

```*MyConsumer``` has a method that implements ```Kafka.MessageHandlerFunc```  

See ```Kafka.MessageHandlerFunc``` for details on what methods are acceptable as message handler functions you can use in the ```consumer.AddHandler``` call.

See ```Kafka.Binder``` for details on additional details with regard to creating Producer, Consumer and Subscriber.