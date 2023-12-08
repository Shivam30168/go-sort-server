package main

import (
	"encoding/json"
	"net/http"
	"sort"
	"sync"
	"time"
)

type RequestPayload struct {
	ToSort [][]int `json:"to_sort"`
}


type ResponsePayload struct {
	SortedArrays [][]int `json:"sorted_arrays"`
	TimeNs       int64   `json:"time_ns"`
}

func main() {
	http.HandleFunc("/process-single", ProcessSingle)
	http.HandleFunc("/process-concurrent", ProcessConcurrent)
	http.ListenAndServe(":8000", nil)
}

func ProcessSingle(w http.ResponseWriter, r *http.Request) {
	processRequest(w, r, false)
}

func ProcessConcurrent(w http.ResponseWriter, r *http.Request) {
	processRequest(w, r, true)
}

func processRequest(w http.ResponseWriter, r *http.Request, concurrent bool) {
	var reqPayload RequestPayload
	err := json.NewDecoder(r.Body).Decode(&reqPayload)
	if err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	startTime := time.Now()

	var sortedArrays [][]int
	if concurrent {
		sortedArrays = sortConcurrently(reqPayload.ToSort)
	} else {
		sortedArrays = sortSequentially(reqPayload.ToSort)
	}

	duration := time.Since(startTime).Nanoseconds()

	respPayload := ResponsePayload{
		SortedArrays: sortedArrays,
		TimeNs:       duration,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(respPayload)
}

func sortSequentially(arrays [][]int) [][]int {
	var sortedArrays [][]int

	for _, arr := range arrays {
		sorted := make([]int, len(arr))
		copy(sorted, arr)
		sort.Ints(sorted)
		sortedArrays = append(sortedArrays, sorted)
	}

	return sortedArrays
}

func sortConcurrently(arrays [][]int) [][]int {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var sortedArrays [][]int

	for _, arr := range arrays {
		wg.Add(1)
		go func(arr []int) {
			defer wg.Done()

			sorted := make([]int, len(arr))
			copy(sorted, arr)
			sort.Ints(sorted)

			mu.Lock()
			sortedArrays = append(sortedArrays, sorted)
			mu.Unlock()
		}(arr)
	}

	wg.Wait()

	return sortedArrays
}
