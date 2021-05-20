// o2.go

package o_o

import (
	"bytes"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"sync"
)

var lck sync.Mutex
//            stack ===> key ==> value=>count
var loc = map[string]map[int]map[int]int{}

// kind: get, put, has, del, misc
func LogLoc(kind string, keySize, valueSize int) {
	lck.Lock()
	defer lck.Unlock()

	// < 1024 - keep
	// floor to nearest k
	adj := func(v int) int {
		if v > 1024 {
			v = v / 1024 * 1024
		}
		return v
	}

	// reduce recursive iavl functions to one
	reduce := func(stack string) string {
		r := regexp.MustCompile("(<= [^=]+iavl[^=]+ )+")
		return r.ReplaceAllString(stack, "<= IAVL ")
	}

	stack := reduce(CallStack(2, 30))
	keySize = adj(keySize)
	valueSize = adj(valueSize)

	add := func(name string) {
		keys, ok := loc[name]
		if !ok {
			keys = map[int]map[int]int{}
			loc[name] = keys
		}
		values, ok := keys[keySize]
		if !ok {
			values = map[int]int{}
			keys[keySize] = values
		}
		count, ok := values[valueSize]
		if !ok {
			values[valueSize] = 1
		} else {
			values[valueSize] = count + 1
		}
	}
	add(stack)
	add(kind)
}

func PrintLoc() string {
	lck.Lock()
	defer lck.Unlock()

	var b bytes.Buffer
	b.WriteString("*** DB Access Pattern\n")

	sortedStringKeys := func(m interface{}) []string {
		var names []string
		for _, v := range reflect.ValueOf(m).MapKeys() {
			names = append(names, v.Interface().(string))
		}
		sort.Strings(names)
		return names
	}

	sortedIntKeys := func(m interface{}) []int {
		var keys []int
		for _, v := range reflect.ValueOf(m).MapKeys() {
			keys = append(keys, v.Interface().(int))
		}
		sort.Ints(keys)
		return keys
	}

	// stack => map[int]map[int]int
	for _, name := range sortedStringKeys(loc) {
		b.WriteString(fmt.Sprintf("%s\n  ", name))
		keysMap := loc[name]
		for _, keySize := range sortedIntKeys(keysMap) {
			valuesMap := keysMap[keySize]
			totalCount := 0
			for _, count := range valuesMap {
				totalCount += count
			}
			b.WriteString(fmt.Sprintf("key(%d,%d)=", keySize, totalCount))
			first := true
			for _, valueSize := range sortedIntKeys(valuesMap) {
				count := valuesMap[valueSize]
				if !first {
					b.WriteString(",")
				} else {
					first = false
				}
				b.WriteString(fmt.Sprintf("%d:%d", valueSize, count))
			}
			b.WriteString("  ")
		}
		b.WriteString("\n")
	}
	fmt.Print(b.String())
	return b.String()
}

// EOF
