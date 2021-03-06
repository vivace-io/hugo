// Copyright 2017 The Hugo Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package collections

import (
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/spf13/hugo/config"
	"github.com/spf13/hugo/deps"
	"github.com/spf13/hugo/helpers"
	"github.com/spf13/hugo/hugofs"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type tstNoStringer struct{}

func TestAfter(t *testing.T) {
	t.Parallel()

	ns := New(&deps.Deps{})

	for i, test := range []struct {
		index  interface{}
		seq    interface{}
		expect interface{}
	}{
		{int(2), []string{"a", "b", "c", "d"}, []string{"c", "d"}},
		{int32(3), []string{"a", "b"}, false},
		{int64(2), []int{100, 200, 300}, []int{300}},
		{100, []int{100, 200}, false},
		{"1", []int{100, 200, 300}, []int{200, 300}},
		{int64(-1), []int{100, 200, 300}, false},
		{"noint", []int{100, 200, 300}, false},
		{1, nil, false},
		{nil, []int{100}, false},
		{1, t, false},
		{1, (*string)(nil), false},
	} {
		errMsg := fmt.Sprintf("[%d] %v", i, test)

		result, err := ns.After(test.index, test.seq)

		if b, ok := test.expect.(bool); ok && !b {
			require.Error(t, err, errMsg)
			continue
		}

		require.NoError(t, err, errMsg)
		assert.Equal(t, test.expect, result, errMsg)
	}
}

func TestDelimit(t *testing.T) {
	t.Parallel()

	ns := New(&deps.Deps{})

	for i, test := range []struct {
		seq       interface{}
		delimiter interface{}
		last      interface{}
		expect    template.HTML
	}{
		{[]string{"class1", "class2", "class3"}, " ", nil, "class1 class2 class3"},
		{[]int{1, 2, 3, 4, 5}, ",", nil, "1,2,3,4,5"},
		{[]int{1, 2, 3, 4, 5}, ", ", nil, "1, 2, 3, 4, 5"},
		{[]string{"class1", "class2", "class3"}, " ", " and ", "class1 class2 and class3"},
		{[]int{1, 2, 3, 4, 5}, ",", ",", "1,2,3,4,5"},
		{[]int{1, 2, 3, 4, 5}, ", ", ", and ", "1, 2, 3, 4, and 5"},
		// test maps with and without sorting required
		{map[string]int{"1": 10, "2": 20, "3": 30, "4": 40, "5": 50}, "--", nil, "10--20--30--40--50"},
		{map[string]int{"3": 10, "2": 20, "1": 30, "4": 40, "5": 50}, "--", nil, "30--20--10--40--50"},
		{map[string]string{"1": "10", "2": "20", "3": "30", "4": "40", "5": "50"}, "--", nil, "10--20--30--40--50"},
		{map[string]string{"3": "10", "2": "20", "1": "30", "4": "40", "5": "50"}, "--", nil, "30--20--10--40--50"},
		{map[string]string{"one": "10", "two": "20", "three": "30", "four": "40", "five": "50"}, "--", nil, "50--40--10--30--20"},
		{map[int]string{1: "10", 2: "20", 3: "30", 4: "40", 5: "50"}, "--", nil, "10--20--30--40--50"},
		{map[int]string{3: "10", 2: "20", 1: "30", 4: "40", 5: "50"}, "--", nil, "30--20--10--40--50"},
		{map[float64]string{3.3: "10", 2.3: "20", 1.3: "30", 4.3: "40", 5.3: "50"}, "--", nil, "30--20--10--40--50"},
		// test maps with a last delimiter
		{map[string]int{"1": 10, "2": 20, "3": 30, "4": 40, "5": 50}, "--", "--and--", "10--20--30--40--and--50"},
		{map[string]int{"3": 10, "2": 20, "1": 30, "4": 40, "5": 50}, "--", "--and--", "30--20--10--40--and--50"},
		{map[string]string{"1": "10", "2": "20", "3": "30", "4": "40", "5": "50"}, "--", "--and--", "10--20--30--40--and--50"},
		{map[string]string{"3": "10", "2": "20", "1": "30", "4": "40", "5": "50"}, "--", "--and--", "30--20--10--40--and--50"},
		{map[string]string{"one": "10", "two": "20", "three": "30", "four": "40", "five": "50"}, "--", "--and--", "50--40--10--30--and--20"},
		{map[int]string{1: "10", 2: "20", 3: "30", 4: "40", 5: "50"}, "--", "--and--", "10--20--30--40--and--50"},
		{map[int]string{3: "10", 2: "20", 1: "30", 4: "40", 5: "50"}, "--", "--and--", "30--20--10--40--and--50"},
		{map[float64]string{3.5: "10", 2.5: "20", 1.5: "30", 4.5: "40", 5.5: "50"}, "--", "--and--", "30--20--10--40--and--50"},
	} {
		errMsg := fmt.Sprintf("[%d] %v", i, test)

		var result template.HTML
		var err error

		if test.last == nil {
			result, err = ns.Delimit(test.seq, test.delimiter)
		} else {
			result, err = ns.Delimit(test.seq, test.delimiter, test.last)
		}

		require.NoError(t, err, errMsg)
		assert.Equal(t, test.expect, result, errMsg)
	}
}

func TestDictionary(t *testing.T) {
	t.Parallel()

	ns := New(&deps.Deps{})

	for i, test := range []struct {
		values []interface{}
		expect interface{}
	}{
		{[]interface{}{"a", "b"}, map[string]interface{}{"a": "b"}},
		{[]interface{}{"a", 12, "b", []int{4}}, map[string]interface{}{"a": 12, "b": []int{4}}},
		// errors
		{[]interface{}{5, "b"}, false},
		{[]interface{}{"a", "b", "c"}, false},
	} {
		errMsg := fmt.Sprintf("[%d] %v", i, test.values)

		result, err := ns.Dictionary(test.values...)

		if b, ok := test.expect.(bool); ok && !b {
			require.Error(t, err, errMsg)
			continue
		}

		require.NoError(t, err, errMsg)
		assert.Equal(t, test.expect, result, errMsg)
	}
}

func TestEchoParam(t *testing.T) {
	t.Parallel()

	ns := New(&deps.Deps{})

	for i, test := range []struct {
		a      interface{}
		key    interface{}
		expect interface{}
	}{
		{[]int{1, 2, 3}, 1, int64(2)},
		{[]uint{1, 2, 3}, 1, uint64(2)},
		{[]float64{1.1, 2.2, 3.3}, 1, float64(2.2)},
		{[]string{"foo", "bar", "baz"}, 1, "bar"},
		{[]TstX{{A: "a", B: "b"}, {A: "c", B: "d"}, {A: "e", B: "f"}}, 1, ""},
		{map[string]int{"foo": 1, "bar": 2, "baz": 3}, "bar", int64(2)},
		{map[string]uint{"foo": 1, "bar": 2, "baz": 3}, "bar", uint64(2)},
		{map[string]float64{"foo": 1.1, "bar": 2.2, "baz": 3.3}, "bar", float64(2.2)},
		{map[string]string{"foo": "FOO", "bar": "BAR", "baz": "BAZ"}, "bar", "BAR"},
		{map[string]TstX{"foo": {A: "a", B: "b"}, "bar": {A: "c", B: "d"}, "baz": {A: "e", B: "f"}}, "bar", ""},
		{map[string]interface{}{"foo": nil}, "foo", ""},
		{(*[]string)(nil), "bar", ""},
	} {
		errMsg := fmt.Sprintf("[%d] %v", i, test)

		result := ns.EchoParam(test.a, test.key)

		assert.Equal(t, test.expect, result, errMsg)
	}
}

func TestFirst(t *testing.T) {
	t.Parallel()

	ns := New(&deps.Deps{})

	for i, test := range []struct {
		limit  interface{}
		seq    interface{}
		expect interface{}
	}{
		{int(2), []string{"a", "b", "c"}, []string{"a", "b"}},
		{int32(3), []string{"a", "b"}, []string{"a", "b"}},
		{int64(2), []int{100, 200, 300}, []int{100, 200}},
		{100, []int{100, 200}, []int{100, 200}},
		{"1", []int{100, 200, 300}, []int{100}},
		{int64(-1), []int{100, 200, 300}, false},
		{"noint", []int{100, 200, 300}, false},
		{1, nil, false},
		{nil, []int{100}, false},
		{1, t, false},
		{1, (*string)(nil), false},
	} {
		errMsg := fmt.Sprintf("[%d] %v", i, test)

		result, err := ns.First(test.limit, test.seq)

		if b, ok := test.expect.(bool); ok && !b {
			require.Error(t, err, errMsg)
			continue
		}

		require.NoError(t, err, errMsg)
		assert.Equal(t, test.expect, result, errMsg)
	}
}

func TestIn(t *testing.T) {
	t.Parallel()

	ns := New(&deps.Deps{})

	for i, test := range []struct {
		l1     interface{}
		l2     interface{}
		expect bool
	}{
		{[]string{"a", "b", "c"}, "b", true},
		{[]interface{}{"a", "b", "c"}, "b", true},
		{[]interface{}{"a", "b", "c"}, "d", false},
		{[]string{"a", "b", "c"}, "d", false},
		{[]string{"a", "12", "c"}, 12, false},
		{[]string{"a", "b", "c"}, nil, false},
		{[]int{1, 2, 4}, 2, true},
		{[]interface{}{1, 2, 4}, 2, true},
		{[]interface{}{1, 2, 4}, nil, false},
		{[]interface{}{nil}, nil, false},
		{[]int{1, 2, 4}, 3, false},
		{[]float64{1.23, 2.45, 4.67}, 1.23, true},
		{[]float64{1.234567, 2.45, 4.67}, 1.234568, false},
		{"this substring should be found", "substring", true},
		{"this substring should not be found", "subseastring", false},
		{nil, "foo", false},
	} {
		errMsg := fmt.Sprintf("[%d] %v", i, test)

		result := ns.In(test.l1, test.l2)
		assert.Equal(t, test.expect, result, errMsg)
	}
}

func TestIntersect(t *testing.T) {
	t.Parallel()

	ns := New(&deps.Deps{})

	for i, test := range []struct {
		l1, l2 interface{}
		expect interface{}
	}{
		{[]string{"a", "b", "c", "c"}, []string{"a", "b", "b"}, []string{"a", "b"}},
		{[]string{"a", "b"}, []string{"a", "b", "c"}, []string{"a", "b"}},
		{[]string{"a", "b", "c"}, []string{"d", "e"}, []string{}},
		{[]string{}, []string{}, []string{}},
		{[]string{"a", "b"}, nil, []interface{}{}},
		{nil, []string{"a", "b"}, []interface{}{}},
		{nil, nil, []interface{}{}},
		{[]string{"1", "2"}, []int{1, 2}, []string{}},
		{[]int{1, 2}, []string{"1", "2"}, []int{}},
		{[]int{1, 2, 4}, []int{2, 4}, []int{2, 4}},
		{[]int{2, 4}, []int{1, 2, 4}, []int{2, 4}},
		{[]int{1, 2, 4}, []int{3, 6}, []int{}},
		{[]float64{2.2, 4.4}, []float64{1.1, 2.2, 4.4}, []float64{2.2, 4.4}},
		// errors
		{"not array or slice", []string{"a"}, false},
		{[]string{"a"}, "not array or slice", false},

		// []interface{} ∩ []interface{}
		{[]interface{}{"a", "b", "c"}, []interface{}{"a", "b", "b"}, []interface{}{"a", "b"}},
		{[]interface{}{1, 2, 3}, []interface{}{1, 2, 2}, []interface{}{1, 2}},
		{[]interface{}{int8(1), int8(2), int8(3)}, []interface{}{int8(1), int8(2), int8(2)}, []interface{}{int8(1), int8(2)}},
		{[]interface{}{int16(1), int16(2), int16(3)}, []interface{}{int16(1), int16(2), int16(2)}, []interface{}{int16(1), int16(2)}},
		{[]interface{}{int32(1), int32(2), int32(3)}, []interface{}{int32(1), int32(2), int32(2)}, []interface{}{int32(1), int32(2)}},
		{[]interface{}{int64(1), int64(2), int64(3)}, []interface{}{int64(1), int64(2), int64(2)}, []interface{}{int64(1), int64(2)}},
		{[]interface{}{float32(1), float32(2), float32(3)}, []interface{}{float32(1), float32(2), float32(2)}, []interface{}{float32(1), float32(2)}},
		{[]interface{}{float64(1), float64(2), float64(3)}, []interface{}{float64(1), float64(2), float64(2)}, []interface{}{float64(1), float64(2)}},

		// []interface{} ∩ []T
		{[]interface{}{"a", "b", "c"}, []string{"a", "b", "b"}, []interface{}{"a", "b"}},
		{[]interface{}{1, 2, 3}, []int{1, 2, 2}, []interface{}{1, 2}},
		{[]interface{}{int8(1), int8(2), int8(3)}, []int8{1, 2, 2}, []interface{}{int8(1), int8(2)}},
		{[]interface{}{int16(1), int16(2), int16(3)}, []int16{1, 2, 2}, []interface{}{int16(1), int16(2)}},
		{[]interface{}{int32(1), int32(2), int32(3)}, []int32{1, 2, 2}, []interface{}{int32(1), int32(2)}},
		{[]interface{}{int64(1), int64(2), int64(3)}, []int64{1, 2, 2}, []interface{}{int64(1), int64(2)}},
		{[]interface{}{float32(1), float32(2), float32(3)}, []float32{1, 2, 2}, []interface{}{float32(1), float32(2)}},
		{[]interface{}{float64(1), float64(2), float64(3)}, []float64{1, 2, 2}, []interface{}{float64(1), float64(2)}},

		// []T ∩ []interface{}
		{[]string{"a", "b", "c"}, []interface{}{"a", "b", "b"}, []string{"a", "b"}},
		{[]int{1, 2, 3}, []interface{}{1, 2, 2}, []int{1, 2}},
		{[]int8{1, 2, 3}, []interface{}{int8(1), int8(2), int8(2)}, []int8{1, 2}},
		{[]int16{1, 2, 3}, []interface{}{int16(1), int16(2), int16(2)}, []int16{1, 2}},
		{[]int32{1, 2, 3}, []interface{}{int32(1), int32(2), int32(2)}, []int32{1, 2}},
		{[]int64{1, 2, 3}, []interface{}{int64(1), int64(2), int64(2)}, []int64{1, 2}},
		{[]float32{1, 2, 3}, []interface{}{float32(1), float32(2), float32(2)}, []float32{1, 2}},
		{[]float64{1, 2, 3}, []interface{}{float64(1), float64(2), float64(2)}, []float64{1, 2}},
	} {
		errMsg := fmt.Sprintf("[%d] %v", i, test)

		result, err := ns.Intersect(test.l1, test.l2)

		if b, ok := test.expect.(bool); ok && !b {
			require.Error(t, err, errMsg)
			continue
		}

		assert.NoError(t, err, errMsg)
		assert.Equal(t, test.expect, result, errMsg)
	}
}

func TestIsSet(t *testing.T) {
	t.Parallel()

	ns := New(newDeps(viper.New()))

	for i, test := range []struct {
		a      interface{}
		key    interface{}
		expect bool
		isErr  bool
		errStr string
	}{
		{[]interface{}{1, 2, 3, 5}, 2, true, false, ""},
		{[]interface{}{1, 2, 3, 5}, 22, false, false, ""},

		{map[string]interface{}{"a": 1, "b": 2}, "b", true, false, ""},
		{map[string]interface{}{"a": 1, "b": 2}, "bc", false, false, ""},

		{time.Now(), "Day", false, false, ""},
		{nil, "nil", false, false, ""},
	} {
		errMsg := fmt.Sprintf("[%d] %v", i, test)

		result, err := ns.IsSet(test.a, test.key)
		if test.isErr {
			assert.EqualError(t, err, test.errStr, errMsg)
			continue
		}

		assert.NoError(t, err, errMsg)
		assert.Equal(t, test.expect, result, errMsg)
	}
}

func TestLast(t *testing.T) {
	t.Parallel()

	ns := New(&deps.Deps{})

	for i, test := range []struct {
		limit  interface{}
		seq    interface{}
		expect interface{}
	}{
		{int(2), []string{"a", "b", "c"}, []string{"b", "c"}},
		{int32(3), []string{"a", "b"}, []string{"a", "b"}},
		{int64(2), []int{100, 200, 300}, []int{200, 300}},
		{100, []int{100, 200}, []int{100, 200}},
		{"1", []int{100, 200, 300}, []int{300}},
		// errors
		{int64(-1), []int{100, 200, 300}, false},
		{"noint", []int{100, 200, 300}, false},
		{1, nil, false},
		{nil, []int{100}, false},
		{1, t, false},
		{1, (*string)(nil), false},
	} {
		errMsg := fmt.Sprintf("[%d] %v", i, test)

		result, err := ns.Last(test.limit, test.seq)

		if b, ok := test.expect.(bool); ok && !b {
			require.Error(t, err, errMsg)
			continue
		}

		require.NoError(t, err, errMsg)
		assert.Equal(t, test.expect, result, errMsg)
	}
}

func TestQuerify(t *testing.T) {
	t.Parallel()

	ns := New(&deps.Deps{})

	for i, test := range []struct {
		params []interface{}
		expect interface{}
	}{
		{[]interface{}{"a", "b"}, "a=b"},
		{[]interface{}{"a", "b", "c", "d", "f", " &"}, `a=b&c=d&f=+%26`},
		// errors
		{[]interface{}{5, "b"}, false},
		{[]interface{}{"a", "b", "c"}, false},
	} {
		errMsg := fmt.Sprintf("[%d] %v", i, test.params)

		result, err := ns.Querify(test.params...)

		if b, ok := test.expect.(bool); ok && !b {
			require.Error(t, err, errMsg)
			continue
		}

		require.NoError(t, err, errMsg)
		assert.Equal(t, test.expect, result, errMsg)
	}
}

func TestSeq(t *testing.T) {
	t.Parallel()

	ns := New(&deps.Deps{})

	for i, test := range []struct {
		args   []interface{}
		expect interface{}
	}{
		{[]interface{}{-2, 5}, []int{-2, -1, 0, 1, 2, 3, 4, 5}},
		{[]interface{}{1, 2, 4}, []int{1, 3}},
		{[]interface{}{1}, []int{1}},
		{[]interface{}{3}, []int{1, 2, 3}},
		{[]interface{}{3.2}, []int{1, 2, 3}},
		{[]interface{}{0}, []int{}},
		{[]interface{}{-1}, []int{-1}},
		{[]interface{}{-3}, []int{-1, -2, -3}},
		{[]interface{}{3, -2}, []int{3, 2, 1, 0, -1, -2}},
		{[]interface{}{6, -2, 2}, []int{6, 4, 2}},
		// errors
		{[]interface{}{1, 0, 2}, false},
		{[]interface{}{1, -1, 2}, false},
		{[]interface{}{2, 1, 1}, false},
		{[]interface{}{2, 1, 1, 1}, false},
		{[]interface{}{2001}, false},
		{[]interface{}{}, false},
		{[]interface{}{0, -1000000}, false},
		{[]interface{}{tstNoStringer{}}, false},
		{nil, false},
	} {
		errMsg := fmt.Sprintf("[%d] %v", i, test)

		result, err := ns.Seq(test.args...)

		if b, ok := test.expect.(bool); ok && !b {
			require.Error(t, err, errMsg)
			continue
		}

		require.NoError(t, err, errMsg)
		assert.Equal(t, test.expect, result, errMsg)
	}
}

func TestShuffle(t *testing.T) {
	t.Parallel()

	ns := New(&deps.Deps{})

	for i, test := range []struct {
		seq     interface{}
		success bool
	}{
		{[]string{"a", "b", "c", "d"}, true},
		{[]int{100, 200, 300}, true},
		{[]int{100, 200, 300}, true},
		{[]int{100, 200}, true},
		{[]string{"a", "b"}, true},
		{[]int{100, 200, 300}, true},
		{[]int{100, 200, 300}, true},
		{[]int{100}, true},
		// errors
		{nil, false},
		{t, false},
		{(*string)(nil), false},
	} {
		errMsg := fmt.Sprintf("[%d] %v", i, test)

		result, err := ns.Shuffle(test.seq)

		if !test.success {
			require.Error(t, err, errMsg)
			continue
		}

		require.NoError(t, err, errMsg)

		resultv := reflect.ValueOf(result)
		seqv := reflect.ValueOf(test.seq)

		assert.Equal(t, resultv.Len(), seqv.Len(), errMsg)
	}
}

func TestShuffleRandomising(t *testing.T) {
	t.Parallel()

	ns := New(&deps.Deps{})

	// Note that this test can fail with false negative result if the shuffle
	// of the sequence happens to be the same as the original sequence. However
	// the propability of the event is 10^-158 which is negligible.
	seqLen := 100
	rand.Seed(time.Now().UTC().UnixNano())

	for _, test := range []struct {
		seq []int
	}{
		{rand.Perm(seqLen)},
	} {
		result, err := ns.Shuffle(test.seq)
		resultv := reflect.ValueOf(result)

		require.NoError(t, err)

		allSame := true
		for i, v := range test.seq {
			allSame = allSame && (resultv.Index(i).Interface() == v)
		}

		assert.False(t, allSame, "Expected sequence to be shuffled but was in the same order")
	}
}

func TestSlice(t *testing.T) {
	t.Parallel()

	ns := New(&deps.Deps{})

	for i, test := range []struct {
		args []interface{}
	}{
		{[]interface{}{"a", "b"}},
		// errors
		{[]interface{}{5, "b"}},
		{[]interface{}{tstNoStringer{}}},
	} {
		errMsg := fmt.Sprintf("[%d] %v", i, test.args)

		result := ns.Slice(test.args...)

		assert.Equal(t, test.args, result, errMsg)
	}
}

func TestUnion(t *testing.T) {
	t.Parallel()

	ns := New(&deps.Deps{})

	for i, test := range []struct {
		l1     interface{}
		l2     interface{}
		expect interface{}
		isErr  bool
	}{
		{[]string{"a", "b", "c", "c"}, []string{"a", "b", "b"}, []string{"a", "b", "c"}, false},
		{[]string{"a", "b"}, []string{"a", "b", "c"}, []string{"a", "b", "c"}, false},
		{[]string{"a", "b", "c"}, []string{"d", "e"}, []string{"a", "b", "c", "d", "e"}, false},
		{[]string{}, []string{}, []string{}, false},
		{[]string{"a", "b"}, nil, []string{"a", "b"}, false},
		{nil, []string{"a", "b"}, []string{"a", "b"}, false},
		{nil, nil, make([]interface{}, 0), true},
		{[]string{"1", "2"}, []int{1, 2}, make([]string, 0), false},
		{[]int{1, 2}, []string{"1", "2"}, make([]int, 0), false},
		{[]int{1, 2, 3}, []int{3, 4, 5}, []int{1, 2, 3, 4, 5}, false},
		{[]int{1, 2, 3}, []int{1, 2, 3}, []int{1, 2, 3}, false},
		{[]int{1, 2, 4}, []int{2, 4}, []int{1, 2, 4}, false},
		{[]int{2, 4}, []int{1, 2, 4}, []int{2, 4, 1}, false},
		{[]int{1, 2, 4}, []int{3, 6}, []int{1, 2, 4, 3, 6}, false},
		{[]float64{2.2, 4.4}, []float64{1.1, 2.2, 4.4}, []float64{2.2, 4.4, 1.1}, false},
		// errors
		{"not array or slice", []string{"a"}, false, true},
		{[]string{"a"}, "not array or slice", false, true},
	} {
		errMsg := fmt.Sprintf("[%d] %v", i, test)

		result, err := ns.Union(test.l1, test.l2)
		if test.isErr {
			assert.Error(t, err, errMsg)
			continue
		}

		assert.NoError(t, err, errMsg)
		assert.Equal(t, test.expect, result, errMsg)
	}
}

func (x *TstX) TstRp() string {
	return "r" + x.A
}

func (x TstX) TstRv() string {
	return "r" + x.B
}

func (x TstX) unexportedMethod() string {
	return x.unexported
}

func (x TstX) MethodWithArg(s string) string {
	return s
}

func (x TstX) MethodReturnNothing() {}

func (x TstX) MethodReturnErrorOnly() error {
	return errors.New("some error occurred")
}

func (x TstX) MethodReturnTwoValues() (string, string) {
	return "foo", "bar"
}

func (x TstX) MethodReturnValueWithError() (string, error) {
	return "", errors.New("some error occurred")
}

func (x TstX) String() string {
	return fmt.Sprintf("A: %s, B: %s", x.A, x.B)
}

type TstX struct {
	A, B       string
	unexported string
}

func newDeps(cfg config.Provider) *deps.Deps {
	l := helpers.NewLanguage("en", cfg)
	l.Set("i18nDir", "i18n")
	return &deps.Deps{
		Cfg:         cfg,
		Fs:          hugofs.NewMem(l),
		ContentSpec: helpers.NewContentSpec(l),
		Log:         jww.NewNotepad(jww.LevelError, jww.LevelError, os.Stdout, ioutil.Discard, "", log.Ldate|log.Ltime),
	}
}
