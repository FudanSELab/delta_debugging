// Copyright 2017 Istio Authors
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

package evaluator

import (
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	pb "istio.io/api/policy/v1beta1"
	pbv "istio.io/api/policy/v1beta1"
	"istio.io/istio/mixer/pkg/attribute"
	"istio.io/istio/mixer/pkg/expr"
	ilt "istio.io/istio/mixer/pkg/il/testing"
)

func TestExpressions(t *testing.T) {
	for _, test := range ilt.TestData {
		if test.E == "" {
			// Skip tests that don't have expression.
			continue
		}

		if test.Fns != nil {
			// Skip tests that have extern functions defined. We cannot inject extern functions into the evaluator.
			// Compiler tests actually also do evaluation.
			continue
		}
		name := "IL/" + test.TestName()
		t.Run(name, func(tt *testing.T) {
			testWithILEvaluator(test, tt)
		})
	}
}

func testWithILEvaluator(test ilt.TestInfo, t *testing.T) {
	evaluator := initEvaluator(t, test.Conf())
	bag := ilt.NewFakeBag(test.I)

	r, err := evaluator.Eval(test.E, bag)
	// Evaluator does in-line compilation. Check for both.
	if test.CompileErr != "" {
		if err == nil {
			t.Errorf("expected compile error was not thrown: %s", test.CompileErr)
		} else if !strings.HasPrefix(err.Error(), test.CompileErr) {
			t.Errorf("Error mismatch: '%s' != '%s'", err.Error(), test.CompileErr)
		}
		return
	}

	if err = test.CheckEvaluationResult(r, err); err != nil {
		t.Errorf(err.Error())
		return
	}

	if !test.CheckReferenced(bag) {
		t.Errorf("Referenced attribute mismatch: '%v' != '%v'", bag.ReferencedList(), test.Referenced)
		return
	}

	// Depending on the type, try testing specialized methods as well.

	switch test.R.(type) {
	case string:
		astr, err := evaluator.EvalString(test.E, bag)
		if e := test.CheckEvaluationResult(astr, err); e != nil {
			t.Errorf(e.Error())
			return
		}

	case bool:
		abool, err := evaluator.EvalPredicate(test.E, bag)
		if e := test.CheckEvaluationResult(abool, err); e != nil {
			t.Errorf(e.Error())
			return
		}
	}
}

func TestEvalString_WrongType(t *testing.T) {
	e := initEvaluator(t, configInt)
	bag := initBag(int64(23))
	r, err := e.EvalString("attr", bag)
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	if r != "23" {
		t.Fatalf("Unexpected result: r: %v, expected: %v", r, "23")
	}
}

func TestEvalString_Error(t *testing.T) {
	e := initEvaluator(t, configString)
	bag := initBag("foo")
	_, err := e.EvalString("bar", bag)
	if err == nil {
		t.Fatal("Was expecting an error")
	}
}

func TestEvalPredicate_WrongType(t *testing.T) {
	e := initEvaluator(t, configBool)
	bag := initBag(int64(23))
	_, err := e.EvalPredicate("attr", bag)
	if err == nil {
		t.Fatal("Was expecting an error")
	}
}

func TestEvalPredicate_Error(t *testing.T) {
	e := initEvaluator(t, configBool)
	bag := initBag(true)
	_, err := e.EvalPredicate("boo", bag)
	if err == nil {
		t.Fatal("Was expecting an error")
	}
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// This test adds concurrent expression evaluation across
// many go routines.
func TestConcurrent(t *testing.T) {
	bags := []attribute.Bag{}
	maxNum := 64

	for i := 0; i < maxNum; i++ {
		v := randString(6)
		bags = append(bags, ilt.NewFakeBag(
			map[string]interface{}{
				"attr": v,
			},
		))
	}

	expression := fmt.Sprintf("attr == \"%s\"", randString(16))
	maxThreads := 10

	e := initEvaluator(t, configString)
	errChan := make(chan error, len(bags)*maxThreads)

	wg := sync.WaitGroup{}
	for j := 0; j < maxThreads; j++ {
		wg.Add(1)
		go func() {
			for _, b := range bags {
				ok, err := e.EvalPredicate(expression, b)
				if err != nil {
					errChan <- err
					continue
				}
				if ok {
					errChan <- errors.New("unexpected ok")
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()

	if len(errChan) > 0 {
		t.Fatalf("Failed with %d errors: %v", len(errChan), <-errChan)
	}
}

func TestEvalType(t *testing.T) {
	e := initEvaluator(t, configBool)
	ty, err := e.EvalType("attr", e.getAttrContext().finder)
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	if ty != pbv.BOOL {
		t.Fatalf("Unexpected type: %v", ty)
	}
}

func TestEvalType_WrongType(t *testing.T) {
	e := initEvaluator(t, configBool)
	_, err := e.EvalType("boo", e.getAttrContext().finder)
	if err == nil {
		t.Fatal("Was expecting an error")
	}
}

func TestAssertType_WrongType(t *testing.T) {
	e := initEvaluator(t, configBool)
	err := e.AssertType("attr", e.getAttrContext().finder, pbv.STRING)
	if err == nil {
		t.Fatal("Was expecting an error")
	}
}

func TestAssertType_EvaluationError(t *testing.T) {
	e := initEvaluator(t, configBool)
	err := e.AssertType("boo", e.getAttrContext().finder, pbv.BOOL)
	if err == nil {
		t.Fatal("Was expecting an error")
	}
}

func TestConfigChange(t *testing.T) {
	e := initEvaluator(t, configInt)
	bag := initBag(int64(23))

	// Prime the cache
	_, err := e.evalResult("attr", bag)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	f := expr.NewFinder(configBool)
	e.ChangeVocabulary(f)
	if !reflect.DeepEqual(e.getAttrContext().finder, f) {
		t.Fatal("Finder is not set correctly")
	}

	bag = initBag(true)
	_, err = e.EvalPredicate("attr", bag)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func Test_Stress(t *testing.T) {
	src := rand.NewSource(time.Now().UnixNano())
	rnd := rand.New(src)

	e := initEvaluator(t, configString)

	exprs := []string{
		`attr`,
		`attr == "foo"`,
		`attr != "bar"`,
		`attr | "baz"`,
	}

	for i := 0; i < 1000000; i++ {

		for j, exp := range exprs {
			str := generateRandomStr(rnd)
			bag := initBag(str)

			r, err := e.Eval(exp, bag)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if j == 0 {
				if r != str {
					t.Fatalf("%v != %v", r, str)
				}
			}
		}
	}
}

func Test_TypeChecker_Uninitialized(t *testing.T) {
	e, err := NewILEvaluator(10)
	if err != nil {
		t.Fatalf("error: %s", err)
	}

	aType, err := e.EvalType("attr", expr.NewFinder(configString))
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	if aType != pbv.STRING {
		t.Fatalf("attr should have been a string: %s", aType)
	}

	aType, err = e.EvalType("attr", expr.NewFinder(configInt))
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	if aType != pbv.INT64 {
		t.Fatalf("attr should have been an int: %s", aType)
	}
}

func generateRandomStr(r *rand.Rand) string {
	size := r.Intn(20) + 1
	bytes := make([]byte, size)

	for i := 0; i < size; i++ {
		b := byte('a') + byte(r.Intn(26))
		bytes[i] = b
	}

	return string(bytes)
}

func initBag(attrValue interface{}) attribute.Bag {
	attrs := make(map[string]interface{})
	attrs["attr"] = attrValue

	return ilt.NewFakeBag(attrs)
}

func initEvaluator(t *testing.T, attrs map[string]*pb.AttributeManifest_AttributeInfo) *IL {
	e, err := NewILEvaluator(10)
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	finder := expr.NewFinder(attrs)
	e.ChangeVocabulary(finder)
	return e
}

var configInt = map[string]*pb.AttributeManifest_AttributeInfo{
	"attr": {
		ValueType: pbv.INT64,
	},
}

var configString = map[string]*pb.AttributeManifest_AttributeInfo{
	"attr": {
		ValueType: pbv.STRING,
	},
}

var configBool = map[string]*pb.AttributeManifest_AttributeInfo{
	"attr": {
		ValueType: pbv.BOOL,
	},
}
