package main

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

const multiHashNum = 6

var muMd5 = &sync.Mutex{}

func interfaceToString(dataRaw interface{}) string {
	return fmt.Sprintf("%v", dataRaw)
}

func ExecutePipeline(jobs ...job) {
	in := make(chan interface{}, 0)
	wg := &sync.WaitGroup{}
	wg.Add(len(jobs))
	for idx, j := range jobs {
		out := make(chan interface{}, MaxInputDataLen)
		go func(j job, in, out chan interface{}, wg *sync.WaitGroup) {
			defer func() {
				close(out)
				wg.Done()
			}()
			j(in, out)
		}(j, in, out, wg)
		if idx != len(jobs)-1 {
			in = out
		}
	}
	wg.Wait()
}

func calcMd5(data string) string {
	defer muMd5.Unlock()
	muMd5.Lock()
	return DataSignerMd5(data)
}

func SingleHash(in, out chan interface{}) {
	wgOuter := &sync.WaitGroup{}
	for dataRaw := range in {
		wgOuter.Add(1)
		go func(dataRaw interface{}) {
			inChMd5 := make(chan string)
			outChMd5 := make(chan string)
			inChCrc32 := make(chan string)
			outChCrc32 := make(chan string)
			defer func() {
				close(inChMd5)
				close(outChMd5)
				close(inChCrc32)
				close(outChCrc32)
				wgOuter.Done()
			}()
			data := interfaceToString(dataRaw)
			go func(in, out chan string) {
				data := <-in
				out <- DataSignerCrc32(calcMd5(data))
			}(inChMd5, outChMd5)
			go func(in, out chan string) {
				data := <-in
				out <- DataSignerCrc32(data)
			}(inChCrc32, outChCrc32)
			inChMd5 <- data
			inChCrc32 <- data
			resultCrc32 := <-outChCrc32
			resultMd5 := <-outChMd5
			out <- fmt.Sprintf("%s~%s", resultCrc32, resultMd5)
		}(dataRaw)
	}
	wgOuter.Wait()
}

func MultiHash(in, out chan interface{}) {
	wgOuter := &sync.WaitGroup{}
	for dataRaw := range in {
		wgOuter.Add(1)
		go func(dataRaw interface{}) {
			defer wgOuter.Done()
			wg := &sync.WaitGroup{}
			mu := &sync.Mutex{}
			data := interfaceToString(dataRaw)
			wg.Add(multiHashNum)
			resultMap := map[int]string{}
			for i := 0; i < multiHashNum; i++ {
				ch := make(chan int)
				go func(in chan int, resultMap map[int]string, data string) {
					defer func() {
						mu.Unlock()
						wg.Done()
					}()
					i := <-in
					result := DataSignerCrc32(fmt.Sprintf("%d%s", i, data))
					mu.Lock()
					resultMap[i] = result
				}(ch, resultMap, data)
				ch <- i
			}
			wg.Wait()
			var result string
			for i := 0; i < multiHashNum; i++ {
				result += resultMap[i]
			}
			out <- result
		}(dataRaw)
	}
	wgOuter.Wait()
}

func CombineResults(in, out chan interface{}) {
	var results []string
	for dataRaw := range in {
		data := interfaceToString(dataRaw)
		results = append(results, data)
	}
	sort.Strings(results)
	out <- strings.Join(results, "_")
}
