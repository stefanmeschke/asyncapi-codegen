// Package "v2" provides primitives to interact with the AsyncAPI specification.
//
// Code generated by github.com/lerenn/asyncapi-codegen version (devel) DO NOT EDIT.
package v2

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lerenn/asyncapi-codegen/pkg/extensions"
)

// AppSubscriber represents all handlers that are expecting messages for App
type AppSubscriber interface {
	// Hello subscribes to messages placed on the 'hello' channel
	Hello(ctx context.Context, msg HelloMessage)
}

// AppController is the structure that provides publishing capabilities to the
// developer and and connect the broker with the App
type AppController struct {
	controller
}

// NewAppController links the App to the broker
func NewAppController(bc extensions.BrokerController, options ...ControllerOption) (*AppController, error) {
	// Check if broker controller has been provided
	if bc == nil {
		return nil, extensions.ErrNilBrokerController
	}

	// Create default controller
	controller := controller{
		broker:         bc,
		cancelChannels: make(map[string]chan any),
		logger:         extensions.DummyLogger{},
		middlewares:    make([]extensions.Middleware, 0),
	}

	// Apply options
	for _, option := range options {
		option(&controller)
	}

	return &AppController{controller: controller}, nil
}

func (c AppController) wrapMiddlewares(middlewares []extensions.Middleware, last extensions.NextMiddleware) func(ctx context.Context) {
	var called bool

	// If there is no more middleware
	if len(middlewares) == 0 {
		return func(ctx context.Context) {
			if !called {
				called = true
				last(ctx)
			}
		}
	}

	// Wrap middleware into a check function that will call execute the middleware
	// and call the next wrapped middleware if the returned function has not been
	// called already
	next := c.wrapMiddlewares(middlewares[1:], last)
	return func(ctx context.Context) {
		// Call the middleware and the following if it has not been done already
		if !called {
			called = true
			ctx = middlewares[0](ctx, next)

			// If next has already been called in middleware, it should not be
			// executed again
			next(ctx)
		}
	}
}

func (c AppController) executeMiddlewares(ctx context.Context, callback func(ctx context.Context)) {
	// Wrap middleware to have 'next' function when calling them
	wrapped := c.wrapMiddlewares(c.middlewares, callback)

	// Execute wrapped middlewares
	wrapped(ctx)
}

func addAppContextValues(ctx context.Context, path string) context.Context {
	ctx = context.WithValue(ctx, extensions.ContextKeyIsVersion, "2.0.0")
	ctx = context.WithValue(ctx, extensions.ContextKeyIsProvider, "app")
	return context.WithValue(ctx, extensions.ContextKeyIsChannel, path)
}

// Close will clean up any existing resources on the controller
func (c *AppController) Close(ctx context.Context) {
	// Unsubscribing remaining channels
	c.UnsubscribeAll(ctx)

	c.logger.Info(ctx, "Closed app controller")
}

// SubscribeAll will subscribe to channels without parameters on which the app is expecting messages.
// For channels with parameters, they should be subscribed independently.
func (c *AppController) SubscribeAll(ctx context.Context, as AppSubscriber) error {
	if as == nil {
		return extensions.ErrNilAppSubscriber
	}

	if err := c.SubscribeHello(ctx, as.Hello); err != nil {
		return err
	}

	return nil
}

// UnsubscribeAll will unsubscribe all remaining subscribed channels
func (c *AppController) UnsubscribeAll(ctx context.Context) {
	c.UnsubscribeHello(ctx)
}

// SubscribeHello will subscribe to new messages from 'hello' channel.
//
// Callback function 'fn' will be called each time a new message is received.
// The 'done' argument indicates when the subscription is canceled and can be
// used to clean up resources.
func (c *AppController) SubscribeHello(ctx context.Context, fn func(ctx context.Context, msg HelloMessage)) error {
	// Get channel path
	path := "hello"

	// Set context
	ctx = addAppContextValues(ctx, path)
	ctx = context.WithValue(ctx, extensions.ContextKeyIsMessageDirection, "reception")

	// Check if there is already a subscription
	_, exists := c.cancelChannels[path]
	if exists {
		err := fmt.Errorf("%w: %q channel is already subscribed", extensions.ErrAlreadySubscribedChannel, path)
		c.logger.Error(ctx, err.Error())
		return err
	}

	// Subscribe to broker channel
	msgs, cancel, err := c.broker.Subscribe(ctx, path)
	if err != nil {
		c.logger.Error(ctx, err.Error())
		return err
	}
	c.logger.Info(ctx, "Subscribed to channel")

	// Asynchronously listen to new messages and pass them to app subscriber
	go func() {
		for {
			// Wait for next message
			bMsg, open := <-msgs

			// If subscription is closed and there is no more message
			// (i.e. uninitialized message), then exit the function
			if !open && bMsg.IsUninitialized() {
				return
			}

			// Set broker message to context
			ctx = context.WithValue(ctx, extensions.ContextKeyIsBrokerMessage, bMsg)

			// Process message
			msg, err := newHelloMessageFromBrokerMessage(bMsg)
			if err != nil {
				c.logger.Error(ctx, err.Error())
			}
			msgCtx := context.WithValue(ctx, extensions.ContextKeyIsMessage, msg)

			// Execute middlewares with the callback
			c.executeMiddlewares(msgCtx, func(ctx context.Context) {
				fn(ctx, msg)
			})
		}
	}()

	// Add the cancel channel to the inside map
	c.cancelChannels[path] = cancel

	return nil
}

// UnsubscribeHello will unsubscribe messages from 'hello' channel
func (c *AppController) UnsubscribeHello(ctx context.Context) {
	// Get channel path
	path := "hello"

	// Check if there subscribers for this channel
	cancel, exists := c.cancelChannels[path]
	if !exists {
		return
	}

	// Set context
	ctx = addAppContextValues(ctx, path)

	// Stop the subscription and wait for its closure to be complete
	cancel <- true
	<-cancel

	// Remove if from the subscribers
	delete(c.cancelChannels, path)

	c.logger.Info(ctx, "Unsubscribed from channel")
}

// UserController is the structure that provides publishing capabilities to the
// developer and and connect the broker with the User
type UserController struct {
	controller
}

// NewUserController links the User to the broker
func NewUserController(bc extensions.BrokerController, options ...ControllerOption) (*UserController, error) {
	// Check if broker controller has been provided
	if bc == nil {
		return nil, extensions.ErrNilBrokerController
	}

	// Create default controller
	controller := controller{
		broker:         bc,
		cancelChannels: make(map[string]chan any),
		logger:         extensions.DummyLogger{},
		middlewares:    make([]extensions.Middleware, 0),
	}

	// Apply options
	for _, option := range options {
		option(&controller)
	}

	return &UserController{controller: controller}, nil
}

func (c UserController) wrapMiddlewares(middlewares []extensions.Middleware, last extensions.NextMiddleware) func(ctx context.Context) {
	var called bool

	// If there is no more middleware
	if len(middlewares) == 0 {
		return func(ctx context.Context) {
			if !called {
				called = true
				last(ctx)
			}
		}
	}

	// Wrap middleware into a check function that will call execute the middleware
	// and call the next wrapped middleware if the returned function has not been
	// called already
	next := c.wrapMiddlewares(middlewares[1:], last)
	return func(ctx context.Context) {
		// Call the middleware and the following if it has not been done already
		if !called {
			called = true
			ctx = middlewares[0](ctx, next)

			// If next has already been called in middleware, it should not be
			// executed again
			next(ctx)
		}
	}
}

func (c UserController) executeMiddlewares(ctx context.Context, callback func(ctx context.Context)) {
	// Wrap middleware to have 'next' function when calling them
	wrapped := c.wrapMiddlewares(c.middlewares, callback)

	// Execute wrapped middlewares
	wrapped(ctx)
}

func addUserContextValues(ctx context.Context, path string) context.Context {
	ctx = context.WithValue(ctx, extensions.ContextKeyIsVersion, "2.0.0")
	ctx = context.WithValue(ctx, extensions.ContextKeyIsProvider, "user")
	return context.WithValue(ctx, extensions.ContextKeyIsChannel, path)
}

// Close will clean up any existing resources on the controller
func (c *UserController) Close(ctx context.Context) {
	// Unsubscribing remaining channels
}

// PublishHello will publish messages to 'hello' channel
func (c *UserController) PublishHello(ctx context.Context, msg HelloMessage) error {
	// Get channel path
	path := "hello"

	// Set context
	ctx = addUserContextValues(ctx, path)
	ctx = context.WithValue(ctx, extensions.ContextKeyIsMessage, msg)
	ctx = context.WithValue(ctx, extensions.ContextKeyIsMessageDirection, "publication")

	// Convert to BrokerMessage
	bMsg, err := msg.toBrokerMessage()
	if err != nil {
		return err
	}

	// Set broker message to context
	ctx = context.WithValue(ctx, extensions.ContextKeyIsBrokerMessage, bMsg)

	// Publish the message on event-broker through middlewares
	c.executeMiddlewares(ctx, func(ctx context.Context) {
		err = c.broker.Publish(ctx, path, bMsg)
	})

	// Return error from publication on broker
	return err
}

// controller is the controller that will be used to communicate with the broker
// It will be used internally by AppController and UserController
type controller struct {
	// broker is the broker controller that will be used to communicate
	broker extensions.BrokerController
	// cancelChannels is a map of cancel channels for each subscribed channel
	cancelChannels map[string]chan any
	// logger is the logger that will be used² to log operations on controller
	logger extensions.Logger
	// middlewares are the middlewares that will be executed when sending or
	// receiving messages
	middlewares []extensions.Middleware
}

// ControllerOption is the type of the options that can be passed
// when creating a new Controller
type ControllerOption func(controller *controller)

// WithLogger attaches a logger to the controller
func WithLogger(logger extensions.Logger) ControllerOption {
	return func(controller *controller) {
		controller.logger = logger
	}
}

// WithMiddlewares attaches middlewares that will be executed when sending or receiving messages
func WithMiddlewares(middlewares ...extensions.Middleware) ControllerOption {
	return func(controller *controller) {
		controller.middlewares = middlewares
	}
}

type MessageWithCorrelationID interface {
	CorrelationID() string
	SetCorrelationID(id string)
}

type Error struct {
	Channel string
	Err     error
}

func (e *Error) Error() string {
	return fmt.Sprintf("channel %q: err %v", e.Channel, e.Err)
}

// HelloMessage is the message expected for 'Hello' channel
type HelloMessage struct {
	// Payload will be inserted in the message payload
	Payload struct {
		Message   string    `json:"message"`
		Timestamp time.Time `json:"timestamp"`
	}
}

func NewHelloMessage() HelloMessage {
	var msg HelloMessage

	return msg
}

// newHelloMessageFromBrokerMessage will fill a new HelloMessage with data from generic broker message
func newHelloMessageFromBrokerMessage(bMsg extensions.BrokerMessage) (HelloMessage, error) {
	var msg HelloMessage

	// Unmarshal payload to expected message payload format
	err := json.Unmarshal(bMsg.Payload, &msg.Payload)
	if err != nil {
		return msg, err
	}

	// TODO: run checks on msg type

	return msg, nil
}

// toBrokerMessage will generate a generic broker message from HelloMessage data
func (msg HelloMessage) toBrokerMessage() (extensions.BrokerMessage, error) {
	// TODO: implement checks on message

	// Marshal payload to JSON
	payload, err := json.Marshal(msg.Payload)
	if err != nil {
		return extensions.BrokerMessage{}, err
	}

	// There is no headers here
	headers := make(map[string][]byte, 0)

	return extensions.BrokerMessage{
		Headers: headers,
		Payload: payload,
	}, nil
}