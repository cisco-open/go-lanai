// Copyright 2023 Cisco Systems, Inc. and its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package kafka

import (
	"context"
	"errors"
	"fmt"
	"github.com/IBM/sarama"
	"sync"
)

type saramaGroupConsumer struct {
	sync.Mutex
	topic       string
	group       string
	brokers     []string
	config      *bindingConfig
	dispatcher  *saramaDispatcher
	provisioner *saramaTopicProvisioner
	started     bool
	consumer    sarama.ConsumerGroup
	cancelFunc  context.CancelFunc
	closed      bool
}

func newSaramaGroupConsumer(topic string, group string, addrs []string, config *bindingConfig, provisioner *saramaTopicProvisioner) (*saramaGroupConsumer, error) {
	if group == "" {
		return nil, ErrorSubTypeBindingInternal.WithMessage("group is required and cannot be empty")
	}

	//config.Consumer.Return.Errors = true
	return &saramaGroupConsumer{
		topic:       topic,
		group:       group,
		brokers:     addrs,
		config:      config,
		dispatcher:  newSaramaDispatcher(config),
		provisioner: provisioner,
	}, nil
}

func (g *saramaGroupConsumer) Topic() string {
	return g.topic
}

func (g *saramaGroupConsumer) Group() string {
	return g.group
}

func (g *saramaGroupConsumer) Start(ctx context.Context) (err error) {
	g.Lock()
	defer g.Unlock()
	defer func() {
		if err == nil {
			g.started = true
		}
	}()
	switch {
	case g.closed:
		return ErrorStartClosedBinding.WithMessage("cannot re-start a closed consumer [%s]", g.topic)
	case g.started:
		return nil
	}

	if ok, e := g.provisioner.topicExists(g.topic); e != nil || !ok {
		return NewKafkaError(ErrorCodeIllegalState, fmt.Sprintf(`topic "%s" does not exists`, g.topic))
	}

	var e error
	g.consumer, e = sarama.NewConsumerGroup(g.brokers, g.group, &g.config.sarama)
	if e != nil {
		err = translateSaramaBindingError(e, e.Error())
		return
	}

	cancelCtx, cancelFunc := context.WithCancel(ctx)
	if g.config.sarama.Consumer.Return.Errors {
		go g.monitorGroupErrors(cancelCtx, g.consumer)
	}
	go g.handleGroup(cancelCtx, g.consumer)
	g.cancelFunc = cancelFunc
	return
}

func (g *saramaGroupConsumer) Close() error {
	g.Lock()
	defer g.Unlock()
	defer func() {
		g.started = false
		g.closed = true
	}()

	if g.cancelFunc != nil {
		g.cancelFunc()
		g.cancelFunc = nil
	}

	if g.consumer == nil {
		return nil
	}

	if e := g.consumer.Close(); e != nil {
		return NewKafkaError(ErrorCodeIllegalState, "error when closing group consumer: %v", e)
	}

	return nil
}

func (g *saramaGroupConsumer) Closed() bool {
	g.Lock()
	defer g.Unlock()
	return g.closed
}

func (g *saramaGroupConsumer) AddHandler(handlerFunc MessageHandlerFunc, opts ...DispatchOptions) error {
	return g.dispatcher.addHandler(handlerFunc, &g.config.consumer, opts)
}

// monitorGroupErrors should be run in separate goroutine
func (g *saramaGroupConsumer) monitorGroupErrors(ctx context.Context, group sarama.ConsumerGroup) {
	for {
		select {
		case e, ok := <-group.Errors():
			if !ok {
				return
			}
			if errors.Is(e, sarama.ErrClosedConsumerGroup) {
				return
			}
			logger.WithContext(ctx).Warnf("Consumer Group Error: %v", e)
		case <-ctx.Done():
			return
		}
	}
}

// handleGroup should be run in separate goroutine
func (g *saramaGroupConsumer) handleGroup(ctx context.Context, group sarama.ConsumerGroup) {
	gh := saramaGroupHandler{
		owner:      g,
		dispatcher: g.dispatcher,
	}

	for {
		// `Consume` should be called inside an infinite loop, when a server-side re-balance happens, the consumer session will need to be recreated to get the new claims
		if e := group.Consume(ctx, []string{g.topic}, gh); e != nil {
			if errors.Is(e, sarama.ErrClosedConsumerGroup) {
				return
			}
			logger.WithContext(ctx).Warnf("Consumer Error: %v", e)
		}
	}
}

// saramaGroupHandler implements sarama.ConsumerGroupHandler
type saramaGroupHandler struct {
	owner      *saramaGroupConsumer
	dispatcher *saramaDispatcher
}

func (h saramaGroupHandler) Setup(session sarama.ConsumerGroupSession) error {
	for topic, parts := range session.Claims() {
		logger.WithContext(session.Context()).
			Debugf("Consumer joined group [%s] Topic=[%s] Partitions=%v MemberID=[%s]", h.owner.group, topic, parts, session.MemberID())
	}
	return nil
}

func (h saramaGroupHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	for topic, parts := range session.Claims() {
		logger.WithContext(session.Context()).
			Debugf("Consumer left group [%s] Topic=[%s] Partitions=%v MemberID=[%s]", h.owner.group, topic, parts, session.MemberID())
	}
	return nil
}

// ConsumeClaim is run in separate goroutine
func (h saramaGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case msg, ok := <-claim.Messages():
			if !ok {
				return nil
			}
			go h.handleMessage(session.Context(), session, msg)
		case <-session.Context().Done():
			return nil
		}
	}
}

// handleMessage intended to run in separate goroutine
func (h saramaGroupHandler) handleMessage(ctx context.Context, session sarama.ConsumerGroupSession, raw *sarama.ConsumerMessage) {
	if e := h.dispatcher.dispatch(ctx, raw, h.owner); e != nil {
		logger.WithContext(ctx).Warnf("failed to handle message: %v", e)
		// TODO we should consider limit retry count, or let Handler decide whether to retry by specifying a special error type
		session.ResetOffset(raw.Topic, raw.Partition, raw.Offset, e.Error())
		return
	}
	session.MarkMessage(raw, "")
}
