package client

import (
	"bytes"
	"context"
	"errors"
	"strconv"
	"sync"
	"time"
)

type Client struct {
	jsonClient      *JsonClient
	extraGenerator  ExtraGenerator
	responses       chan *Response
	listenerStore   *listenerStore
	catchersStore   *sync.Map
	successMsgStore *sync.Map
	updatesTimeout  time.Duration
	catchTimeout    time.Duration
}

type Option func(*Client)

func WithExtraGenerator(extraGenerator ExtraGenerator) Option {
	return func(client *Client) {
		client.extraGenerator = extraGenerator
	}
}

func WithCatchTimeout(timeout time.Duration) Option {
	return func(client *Client) {
		client.catchTimeout = timeout
	}
}

func WithProxy(req *AddProxyRequest) Option {
	return func(client *Client) {
		client.AddProxy(req)
	}
}

func SetLogLevel(level int32) {
	_, _ = SetLogVerbosityLevel(&SetLogVerbosityLevelRequest{
		NewVerbosityLevel: level,
	})
}

func SetFilePath(path string) {
	_, _ = SetLogStream(&SetLogStreamRequest{
		LogStream: &LogStreamFile{
			Path:           path,
			MaxFileSize:    10485760,
			RedirectStderr: true,
		},
	})
}

func NewClient(authorizationStateHandler AuthorizationStateHandler, options ...Option) (*Client, error) {
	client := &Client{
		jsonClient:      NewJsonClient(),
		responses:       make(chan *Response, 1000),
		listenerStore:   newListenerStore(),
		catchersStore:   &sync.Map{},
		successMsgStore: &sync.Map{},
	}

	client.extraGenerator = UuidV4Generator()
	client.catchTimeout = 60 * time.Second

	for _, option := range options {
		option(client)
	}

	tdlibInstance.addClient(client)

	go client.receiver()

	err := Authorize(client, authorizationStateHandler)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (client *Client) receiver() {
	for response := range client.responses {
		if response.Extra != "" {
			value, ok := client.catchersStore.Load(response.Extra)
			if ok {
				value.(chan *Response) <- response
			}
		}

		typ, err := UnmarshalType(response.Data)
		if err != nil {
			continue
		}

		if typ.GetType() == (&UpdateMessageSendSucceeded{}).GetType() {
			value, ok := client.successMsgStore.Load(typ.(*UpdateMessageSendSucceeded).OldMessageId)
			if ok {
				value.(chan *Response) <- response
			}
		}

		needGc := false
		for _, listener := range client.listenerStore.Listeners() {
			if listener.IsActive() && listener.Updates != nil && typ.GetType() == listener.Filter.GetType() { // All updates go to Updates channel if type == filter
				listener.Updates <- typ
			} else if listener.IsActive() && listener.RawUpdates != nil { // All updates go to RawUpdates channel if filter is empty
				listener.RawUpdates <- typ
			} else if !listener.IsActive() { // GC inactive listener
				needGc = true
			}
		}
		if needGc {
			client.listenerStore.gc()
		}
	}
}

func (client *Client) Send(req Request) (*Response, error) {
	req.Extra = client.extraGenerator()

	catcher := make(chan *Response, 1)

	client.catchersStore.Store(req.Extra, catcher)

	defer func() {
		client.catchersStore.Delete(req.Extra)
		close(catcher)
	}()

	client.jsonClient.Send(req)

	ctx, cancel := context.WithTimeout(context.Background(), client.catchTimeout)
	defer cancel()

	select {
	case response := <-catcher:
		if response.Type != "error" && req.Type == "sendMessage" {
			m, err := UnmarshalMessage(response.Data)
			if err != nil {
				return nil, err
			}

			if m.Content.MessageContentType() == "messageText" || m.Content.MessageContentType() == "messageDice" {
				successCatcher := make(chan *Response, 1)
				client.successMsgStore.Store(m.Id, successCatcher)

				defer (func() {
					client.successMsgStore.Delete(m.Id)
					close(successCatcher)
				})()

				select {
				case modResponse := <-successCatcher:
					m2, err2 := UnmarshalUpdateMessageSendSucceeded(modResponse.Data)
					if err2 != nil {
						return response, nil
					}
					response.Data = bytes.Replace(response.Data, []byte("{\"@type\":\"messageSendingStatePending\"}"), []byte("{\"@type\":\"updateMessageSendSucceeded\"}"), 1)
					response.Data = bytes.Replace(response.Data, []byte(strconv.FormatInt(m.Id, 10)), []byte(strconv.FormatInt(m2.Message.Id, 10)), 1)
					return response, nil
				case <-time.After(1 * time.Second):
					client.successMsgStore.Delete(m.Id)
					close(successCatcher)
				}
			}
		}
		return response, nil
	case <-ctx.Done():
		return nil, errors.New("response catching timeout")
	}
}

func (client *Client) GetListener() *Listener {
	listener := &Listener{
		isActive:   true,
		RawUpdates: make(chan Type, 1000),
	}
	client.listenerStore.Add(listener)

	return listener
}

func (client *Client) AddEventReceiver(msgType Type, channelCapacity int) *Listener {
	listener := &Listener{
		isActive: true,
		Updates:  make(chan Type, channelCapacity),
		Filter:   msgType,
	}
	client.listenerStore.Add(listener)

	return listener
}

func (client *Client) Stop() {
	client.Destroy()
}
