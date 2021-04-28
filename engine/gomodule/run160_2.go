package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"sync"
	"test/engine"
	"time"
)

var wg sync.WaitGroup
var mx sync.Mutex

var thread_num = 16
var math_num = 2
var use_lite = false

// var request_num = 16
var request_num = 40000
var max_batch_size = 500

var all_times []time.Duration

var ch = make(chan int, thread_num)

var in_data [][]int64
var res0 [][]float32
var res1 [][]float32
var res2 [][]float32

func alloc_space(max_batch_size int, thread_num int) {
	for i := 0; i < thread_num; i++ {
		data2 := make([]int64, max_batch_size*160)
		in_data = append(in_data, data2)

		data3 := make([]float32, max_batch_size)
		res0 = append(res0, data3)

		data4 := make([]float32, max_batch_size)
		res1 = append(res1, data4)

		data5 := make([]float32, max_batch_size)
		res2 = append(res2, data5)
	}
}

func main() {
	engine.InitEngine("models/160inut/", "", "", thread_num, math_num, use_lite)

	// allocate memory.
	alloc_space(max_batch_size, thread_num)
	batch_size := 500
	features := read_json_data()
	fmt.Println("prepare input data done.")

	for i := 0; i < thread_num; i++ {
		ch <- i
	}

	for i := 0; i < request_num; i++ {
		key_id := <-ch
		wg.Add(1)
		go func(key_id int) {
			start_time := time.Now()
			prepare_input(batch_size, key_id, features)
			engine.Run160Model(key_id, batch_size, in_data[key_id], res0[key_id], res1[key_id], res2[key_id])
			last := time.Now().Sub(start_time)

			// print avg output.
			// print_out_avg(key_id, batch_size)

			// time info.
			mx.Lock()
			all_times = append(all_times, last)

			defer func() {
				wg.Done()
				mx.Unlock()
				ch <- key_id
			}()
		}(key_id)
	}

	wg.Wait()
	engine.TimeInfo(all_times)
}

func read_json_data() map[string][]int64 {
	features := make(map[string][]int64)
	file_bytes, _ := ioutil.ReadFile("data/160_model_input.txt")
	json.Unmarshal(file_bytes, &features)
	return features
}

func prepare_input(batch_size int, thread_id int, features map[string][]int64) {
	for i := 0; i < 160; i++ {
		name := "field_" + strconv.Itoa(i)
		copy(in_data[thread_id][i*batch_size:(i+1)*batch_size], features[name])
	}
}

func print_out_avg(key_id, batch_size int) {
	var avg0 float32 = 0.
	var avg1 float32 = 0.
	var avg2 float32 = 0.
	for i := 0; i < batch_size; i++ {
		avg0 += res0[key_id][i]
		avg1 += res1[key_id][i]
		avg2 += res2[key_id][i]
	}
	fmt.Println(avg0/float32(batch_size), avg1/float32(batch_size), avg2/float32(batch_size))
}