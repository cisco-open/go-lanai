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

// Package internal is an internal package that help to test generated mocks
package internal

import (
    "fmt"
    "github.com/golang/mock/gomock"
    "github.com/onsi/gomega"
    . "github.com/onsi/gomega"
    "reflect"
)

// AssertGoMockGenerated check given "mock" implemented interface T and invocation of interface method is recorded by ctrl
func AssertGoMockGenerated[T any](g *gomega.WithT, mock interface{}, ctrl *gomock.Controller) {
    var targetInterface T
    rtI := reflect.TypeOf(&targetInterface).Elem()
    rv := reflect.ValueOf(mock)

    // get recorder by invoke EXPECT()
    expectFn := rv.MethodByName("EXPECT")
    g.Expect(expectFn.IsZero()).To(BeFalse(), "mock %T should have EXPECT()", mock)
    g.Expect(expectFn.Type().NumIn()).To(BeZero(), "mock %T should have EXPECT() with 0 input", mock)
    out := InvokeFunc(expectFn, []reflect.Value{})
    g.Expect(out).To(HaveLen(1), "mock %T should have EXPECT() with 1 output", mock)
    rExpect := out[0]

    // go through interfaces
    for i := 0; i < rtI.NumMethod(); i++ {
        name := rtI.Method(i).Name
        rm, ok := rv.Type().MethodByName(name)
        g.Expect(ok).To(BeTrue(), "actual mock should implement method [%s]", name)
        AssertGoMockGeneratedMethod(g, rm, rv, rExpect)
    }
}

func AssertGoMockGeneratedMethod(g *gomega.WithT, method reflect.Method, receiver reflect.Value, expect reflect.Value) {
    var out []reflect.Value
    g.Expect(method.IsExported()).To(BeTrue(), "method [%s] should be exported", method.Name)
    actualFn := receiver.MethodByName(method.Name)
    g.Expect(actualFn.IsZero()).To(BeFalse(), "mock should have matching method [%s]", method.Name)
    expectFn := expect.MethodByName(method.Name)
    g.Expect(expectFn.IsZero()).To(BeFalse(), "EXPECT() should have matching method [%s]", method.Name)

    // prepare input params
    ft := method.Func.Type()
    actualIn := make([]reflect.Value, 0, ft.NumIn())
    expectIn := make([]reflect.Value, 0, ft.NumIn())
    // Note: the first param in method is receiver
    for i, isVariadic, lastIdx := 1, ft.IsVariadic(), ft.NumIn()-1; i <= lastIdx; i++ {
        v := MockValue(ft.In(i), isVariadic && i == lastIdx)
        actualIn = append(actualIn, v)
        fmt.Printf("method=%s, IsVariadic=%v, in=%T\n", method.Name, ft.IsVariadic(), v.Interface())
        if isVariadic && i == lastIdx {
            // this is a varargs, the expectIn should be []interface{}{gomock.Any()}
            expectIn = append(expectIn, reflect.ValueOf([]interface{}{gomock.Any()}))
        } else {
            expectIn = append(expectIn, reflect.ValueOf(gomock.Eq(v.Interface())))
        }
    }

    // mock behavior
    out = InvokeFunc(expectFn, expectIn)
    g.Expect(out).To(HaveLen(1), "EXPECT().%s() should return 1 item", method.Name)
    g.Expect(out[0].Interface()).To(BeAssignableToTypeOf(&gomock.Call{}), "EXPECT().%s() should return %T", method.Name, &gomock.Call{})
    mockCall := out[0].Interface().(*gomock.Call)
    mockedRet := make([]interface{}, ft.NumOut())
    for i := 0; i < ft.NumOut(); i++ {
        mockedRet[i] = MockValue(ft.Out(i), false).Interface()
    }
    mockCall.Return(mockedRet...)

    // call actual method
    out = InvokeFunc(actualFn, actualIn)
    g.Expect(out).To(HaveLen(ft.NumOut()), "method [%s] should return correct number of parameters", method.Name)
}

func InvokeFunc(fn reflect.Value, in []reflect.Value) []reflect.Value {
    if fn.Type().IsVariadic() {
        return fn.CallSlice(in)
    } else {
        return fn.Call(in)
    }
}

// MockValue mock a value of given types.
// if this is a slice and varargs, the slice contains one element
func MockValue(typ reflect.Type, isVarargs bool) reflect.Value {
    switch typ.Kind() {
    case reflect.Pointer:
        return reflect.New(typ.Elem())
    case reflect.Slice:
        if isVarargs {
            return reflect.MakeSlice(typ, 1, 1)
        }
        return reflect.MakeSlice(typ, 0, 0)
    default:
        return reflect.Indirect(reflect.New(typ))
    }
}