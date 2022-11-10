package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
)

func ExecutePipeline(args ...job) {
	wg := &sync.WaitGroup{}
	in := make(chan interface{})

	for _, arg := range args {
		wg.Add(1)
		out := make(chan interface{})
		go func(group *sync.WaitGroup, in, out chan interface{}, a job) {
			defer group.Done()
			a(in, out)
			close(out)
		}(wg, in, out, arg)
		in = out
	}
	wg.Wait()
}

func SingleHash(in, out chan interface{}) {
	mu := &sync.Mutex{}
	w := &sync.WaitGroup{}
	for val := range in {
		w.Add(1)
		go func(w *sync.WaitGroup, val interface{}) {
			defer w.Done()
			s := fmt.Sprint(val)
			fmt.Printf("SingleHash data %s\n", s)

			mu.Lock()
			mdHash := DataSignerMd5(s)
			mu.Unlock()

			fmt.Printf("SingleHash md5(data) %s\n", mdHash)

			dataToCrc := []string{
				fmt.Sprint(val),
				mdHash,
			}

			wg := &sync.WaitGroup{}
			hashSlice := make([]string, 2)

			for index, str := range dataToCrc {
				wg.Add(1)
				go func(group *sync.WaitGroup, strToHash string, sl []string, i int) {
					defer group.Done()
					hash := DataSignerCrc32(strToHash)
					sl[i] = hash
				}(wg, str, hashSlice, index)
			}
			wg.Wait()

			out <- strings.Join(hashSlice, "~")
		}(w, val)
	}
	w.Wait()
}

func MultiHash(in, out chan interface{}) {
	w := &sync.WaitGroup{}
	for val := range in {
		strVal := fmt.Sprint(val)

		w.Add(1)
		go func(w *sync.WaitGroup) {
			defer w.Done()
			sliceStr := make([]string, 6)
			wg := &sync.WaitGroup{}
			for th := 0; th < 6; th++ {
				concatStr := strconv.Itoa(th) + strVal
				wg.Add(1)
				go func(group *sync.WaitGroup, strToHash string, idx int, sl []string) {
					defer group.Done()
					hash := DataSignerCrc32(strToHash)
					fmt.Printf("%s MultiHash: crc32(th+step1)) %d %s\n", strVal, idx, hash)
					sliceStr[idx] = hash
				}(wg, concatStr, th, sliceStr)
			}
			wg.Wait()

			resStr := strings.Join(sliceStr, "")
			out <- resStr
			fmt.Printf("%s MultiHash result: %s\n", strVal, resStr)
		}(w)
	}
	w.Wait()
}

func CombineResults(in, out chan interface{}) {
	var strSlice []string
	for val := range in {
		strSlice = append(strSlice, fmt.Sprint(val))
	}

	sort.Strings(strSlice)

	resStr := strings.Join(strSlice, "_")

	out <- resStr
	fmt.Printf("CombineResults %s", resStr)
}
