/*
 * Copyright 2024 The RuleGo Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package processor

import (
	"github.com/rulego/rulego/api/types"
	"github.com/rulego/rulego/api/types/endpoint"
	"sync"
)

// Builtins 内置处理器，endpoint dsl通过name调用对应的处理器
var Builtins = builtins{}

func init() {
	//把http header放入消息元数据
	Builtins.Register("headersToMetadata", func(router endpoint.Router, exchange *endpoint.Exchange) bool {
		msg := exchange.In.GetMsg()
		headers := exchange.In.Headers()
		for k := range headers {
			msg.Metadata.PutValue(k, headers.Get(k))
		}
		return true
	})
	//把msg响应给http客户端
	Builtins.Register("responseToBody", func(router endpoint.Router, exchange *endpoint.Exchange) bool {
		if err := exchange.Out.GetError(); err != nil {
			//错误
			exchange.Out.SetStatusCode(400)
			exchange.Out.SetBody([]byte(exchange.Out.GetError().Error()))
		} else if exchange.Out.GetMsg() != nil {
			if exchange.Out.GetMsg().DataType == types.JSON {
				exchange.Out.Headers().Set("Content-Type", "application/json")
			}
			exchange.Out.SetBody([]byte(exchange.Out.GetMsg().Data))
		}
		return true
	})
}

type builtins struct {
	processors map[string]endpoint.Process
	lock       sync.RWMutex
}

// Register 注册内置处理器
func (b *builtins) Register(name string, processor endpoint.Process) {
	b.lock.Lock()
	defer b.lock.Unlock()
	if b.processors == nil {
		b.processors = make(map[string]endpoint.Process)
	}
	b.processors[name] = processor
}

// RegisterAll 注册内置处理器
func (b *builtins) RegisterAll(processors map[string]endpoint.Process) {
	b.lock.Lock()
	defer b.lock.Unlock()
	if b.processors == nil {
		b.processors = make(map[string]endpoint.Process)
	}
	for k, v := range processors {
		b.processors[k] = v
	}
}

// Unregister 删除内置处理器
func (b *builtins) Unregister(names ...string) {
	b.lock.Lock()
	defer b.lock.Unlock()
	for _, name := range names {
		delete(b.processors, name)
	}
}

// Get 获取内置处理器
func (b *builtins) Get(name string) (endpoint.Process, bool) {
	b.lock.RLock()
	defer b.lock.RUnlock()
	p, ok := b.processors[name]
	return p, ok
}
