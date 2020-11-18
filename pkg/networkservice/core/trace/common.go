// Copyright (c) 2020 Cisco Systems, Inc.
//
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package trace

import (
	"reflect"
	"strconv"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/networkservicemesh/sdk/pkg/tools/spanhelper"
)

func logRequest(span spanhelper.SpanHelper, request proto.Message) {
	connInfo, ok := trace(span.Context())
	if ok && !proto.Equal(connInfo.Request, request) {
		if connInfo.Request != nil {
			requestDiff, hadChanges := Diff(connInfo.Request.ProtoReflect(), request.ProtoReflect())
			if hadChanges {
				span.LogObject("request-diff", requestDiff)
			}
		} else {
			span.LogObject("request", request)
		}
		connInfo.Request = proto.Clone(request)
	}
}

func logResponse(span spanhelper.SpanHelper, response proto.Message) {
	connInfo, ok := trace(span.Context())
	if ok && !proto.Equal(connInfo.Response, response) {
		if connInfo.Response != nil {
			responseDiff, changed := Diff(connInfo.Response.ProtoReflect(), response.ProtoReflect())
			if changed {
				span.LogObject("response-diff", responseDiff)
			}
		} else {
			span.LogObject("response", response)
		}
		connInfo.Response = proto.Clone(response)
		return
	}
}

// Diff - calculate a protobuf messge diff
func Diff(oldMessage, newMessage protoreflect.Message) (map[string]interface{}, bool) {
	diffMessage := map[string]interface{}{}
	oldReflectMessage := oldMessage

	// Marker we had any changes
	changes := 0
	fields := newMessage.Descriptor().Fields()
	for fi := 0; fi < fields.Len(); fi++ {
		descriptor := fields.Get(fi)
		newRefValue := newMessage.Get(descriptor)
		rawOldValue := oldReflectMessage.Get(descriptor)
		oldValue := rawOldValue.Interface()
		newValue := newRefValue.Interface()

		if descriptor.Cardinality() == protoreflect.Repeated {
			originMap := map[string]protoreflect.Value{}
			targetMap := map[string]protoreflect.Value{}

			switch val := newValue.(type) {
			case protoreflect.List:
				// Convert list to MAP with indexes
				oldList := oldValue.(protoreflect.List)

				for i := 0; i < oldList.Len(); i++ {
					originMap[strconv.Itoa(i)] = oldList.Get(i)
				}
				for i := 0; i < val.Len(); i++ {
					targetMap[strconv.Itoa(i)] = val.Get(i)
				}
			case protoreflect.Map:
				oldMap := oldValue.(protoreflect.Map)

				oldMap.Range(func(key protoreflect.MapKey, value protoreflect.Value) bool {
					originMap[key.String()] = value
					return true
				})
				val.Range(func(key protoreflect.MapKey, value protoreflect.Value) bool {
					targetMap[key.String()] = value
					return true
				})
			}
			if resultMap, mapChanged := mapDiff(descriptor, originMap, targetMap); mapChanged {
				changes++
				diffMessage[string(descriptor.Name())] = resultMap
			}
			continue
		}
		val, diff := diffField(descriptor, oldValue, newValue)
		if diff {
			changes++
			diffMessage[string(descriptor.Name())] = val
		}
	}

	return diffMessage, changes > 0
}

func mapDiff(descriptor protoreflect.FieldDescriptor, originMap, targetMap map[string]protoreflect.Value) (interface{}, bool) {
	resultMap := map[string]interface{}{}
	lchanged := 0
	for key, value := range targetMap {
		oldVal, ok := originMap[key]
		if !ok {
			// No old value,
			putToMappDif(value, resultMap, "+"+key)
			lchanged++
			continue
		}
		val, diff := diffField(descriptor, oldVal.Interface(), value.Interface())
		if diff {
			if diff {
				lchanged++
				resultMap[key] = val
			}
		}
	}
	for key, value := range originMap {
		_, ok := targetMap[key]
		if !ok {
			// No new value, mark as deleted
			putToMappDif(value, resultMap, "-"+key)
			lchanged++
		}
	}
	return resultMap, lchanged > 0
}

func putToMappDif(value protoreflect.Value, resultMap map[string]interface{}, key string) {
	if msg, ok := value.Interface().(protoreflect.Message); ok {
		smsg := msg.Interface()
		resultMap[key] = smsg
	} else {
		resultMap[key] = value.String()
	}
}

func diffField(descriptor protoreflect.FieldDescriptor, oldValue, newValue interface{}) (interface{}, bool) {
	if descriptor.Kind() == protoreflect.MessageKind {
		// A pointer to message, we do not need to compare
		if newMsg, ok := newValue.(protoreflect.Message); ok {
			oldMsg, oldOk := oldValue.(protoreflect.Message)
			if !oldOk {
				// No old message defined
				return newMsg.Interface(), true
			}
			fieldDiff, childFieldChanged := Diff(oldMsg, newMsg)
			if childFieldChanged {
				return fieldDiff, true
			}
			return "=", false
		} else if oldMsg, ok := oldValue.(protoreflect.Message); ok {
			// No new message defined
			// No old message defined
			return oldMsg.Interface(), true
		}
	}
	if !reflect.DeepEqual(oldValue, newValue) {
		// Primitive value is not equal, set new value
		return newValue, true
	}
	return nil, false
}