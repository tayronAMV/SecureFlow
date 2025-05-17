package logic

import (
	"math"
	"math/rand"
	"server/internal/db/models"
	"fmt"
)


const (
	CPU_MIN     = 0.1
	DISKIO_MIN  = 0.2
	MEMORY_MIN  = 0.01
	NETWORK_MIN = 0.02
	SYSCALL_MIN = 10
	NUM_TREES_IN_FOREST = 100 
)

type TreeNode struct {
	val   float64
	feature_of_threshold  int // 0=CPU, 1=DiskIO, 2=Memory, 3=Network , 4 = SysCall
	threshold float64
	Left  *TreeNode
	Right *TreeNode
}

func RandFloatN(n float64) float64 {
	return rand.Float64() * n
}

func max_int(a int , b int ) int {
	if a < b {
		return b
	}
	return a 
}


// step 1 : purpose - just to remove dups for the algo noty getting bloated with the same data
func removeDuplicates(logs []models.AnomalyLog) []models.AnomalyLog {
	seen := make(map[string]bool)
	result := []models.AnomalyLog{}

	for _, log := range logs {
		key := fmt.Sprintf("%.2f-%.2f-%.2f-%.2f-%.2f", log.CPU, log.Memory, log.DiskIO, log.Network, log.Syscall)
		if !seen[key] {
			result = append(result, log)
			seen[key] = true
		}
	}
	return result
}

// step 1  
func dataIsValid(logs []models.AnomalyLog) bool {
	streak := 0
	invalid_count := 0 
	for _, log := range logs {
		if math.IsNaN(log.CPU) || math.IsNaN(log.Memory) || math.IsNaN(log.DiskIO) ||
			math.IsNaN(log.Network) || math.IsNaN(log.Syscall) {
			streak++
			invalid_count++
		}else{
			streak = 0 
		}
		if streak>= 5{
			return false
		}
	}
	return float64(invalid_count) <= (float64(len(logs)) * 0.2)
}

// step 1 , data cleaning 
func Data_Cleaning(logs []models.AnomalyLog) ([]models.AnomalyLog , bool){
	logs = removeDuplicates(logs)
	if !dataIsValid(logs){
		return nil , false
	}
	return logs , true

}

//step 2 , give weight to each metric for correct isolation forest based on the container activity
func weightsCalc(arr []*models.AnomalyLog) []float64 {
	var res [5]float64
	p := float64(len(arr))

	for _, s := range arr {
		if s.CPU > CPU_MIN {
			res[0]++
		}
		if s.DiskIO > DISKIO_MIN {
			res[1]++
		}
		if s.Memory > MEMORY_MIN {
			res[2]++
		}
		if s.Network > NETWORK_MIN {
			res[3]++
		}
		if s.Syscall > SYSCALL_MIN {
			res[4]++
		}
	}

	out := make([]float64, 5)
	for i := 0; i < 5; i++ {
		if res[i] == 0 {
			out[i] = 0 // completely unused metric
		} else {
			out[i] = 1 + (res[i] / p) // W₀ + Σf / p
		}
	}
	return out
}

// select random feature to isolate the containers 
func selectRandomMetric(weights []float64) int {
	sum := 0.0
	for _, w := range weights {
		sum += w
	}
	if sum == 0 {
		return rand.Intn(5) // fallback if all weights are zero
	}

	r := RandFloatN(sum)
	for i := 0; i < len(weights); i++ {
		r -= weights[i]
		if r < 0 {
			return i
		}
	}
	return len(weights) - 1 // final fallback
}
func findFeatureMinMax(arr []*models.AnomalyLog, feature int) (float64, float64) {
	if len(arr) == 0 {
		return 0, 0
	}
	get := func(s *models.AnomalyLog) float64 {
		switch feature {
		case 0:
			return s.CPU
		case 1:
			return s.DiskIO
		case 2:
			return s.Memory
		case 3:
			return s.Network
		case 4:
			return s.Syscall
		default:
			return 0
		}
	}
	min := get(arr[0])
	max := get(arr[0])
	for _, s := range arr[1:] {
		val := get(s)
		if val < min {
			min = val
		}
		if val > max {
			max = val
		}
	}
	return min, max
}

// step 4 , building one Itree 
func buildTree(arr []*models.AnomalyLog, depth int) *TreeNode {
	if len(arr) <= 1 || depth >= int(math.Log2(float64(len(arr)))) {
		return nil
	}

	weights := weightsCalc(arr)
	feature := selectRandomMetric(weights)
	min, max := findFeatureMinMax(arr, feature)
	if min == max {
		return nil
	}

	threshold := min + rand.Float64()*(max-min)

	left := []*models.AnomalyLog{}
	right := []*models.AnomalyLog{}

	for _, s := range arr {
		var val float64
		switch feature {
		case 0:
			val = s.CPU
		case 1:
			val = s.DiskIO
		case 2:
			val = s.Memory
		case 3:
			val = s.Network
		case 4:
			val = s.Syscall
		}
		if val < threshold {
			left = append(left, s)
		} else {
			right = append(right, s)
		}
	}

	node := &TreeNode{
		val:  threshold,
		feature_of_threshold: feature,
		threshold: threshold,
	}
	node.Left = buildTree(left, depth+1)
	node.Right = buildTree(right, depth+1)
	return node
}

//step 4 , building forest of Itrees 
func buildForest(arr []*models.AnomalyLog, numTrees int) []*TreeNode {
	forest := make([]*TreeNode, 0, numTrees)
	for i := 0; i < numTrees; i++ {
		sample := sampleSubset(arr, 100) // or any subsample size < len(arr)
		tree := buildTree(sample, 0)
		if tree != nil {
			forest = append(forest, tree)
		}
	}
	return forest
}


//step 4 , take a sample of logs ( like 30 out of 100) and build trees from that for true randomness
func sampleSubset(data []*models.AnomalyLog, size int) []*models.AnomalyLog {
	sample := make([]*models.AnomalyLog, size)
	for i := range size {
		sample[i] = data[rand.Intn(len(data))] // sample with replacement
	}
	return sample
}

//step 5 , calc height of a one tree for one sample , h(x) for a specific anomaly log 
// step 6 , also returns reason of isolation  , for fiding reason of anomaly  
func getTreeHeight(root *TreeNode ,log *models.AnomalyLog , depth int) (int , int) {
	if root == nil{
		return depth , -1 
	}

	// if cur node is a leaf 
	if root.Left == nil && root.Right == nil{
		return depth , root.feature_of_threshold
	}
	
	switch root.feature_of_threshold {
	case 0:
		if root.threshold > log.CPU {
			return getTreeHeight(root.Left , log , depth+1 )
		}
		return getTreeHeight(root.Right , log , depth+1 )
	case 1 :
		if root.threshold > log.DiskIO {
			return getTreeHeight(root.Left , log , depth+1 )
		}
		return getTreeHeight(root.Right , log , depth+1 )
	
	case 2: 
		if root.threshold > log.Memory {
			return getTreeHeight(root.Left , log , depth+1 )
		}
		return getTreeHeight(root.Right , log , depth+1 )

	case 3: 
		if root.threshold > log.Network {
			return getTreeHeight(root.Left , log , depth+1 )
		}
		return getTreeHeight(root.Right , log , depth+1 )

	case 4: 
		if root.threshold > log.Syscall {
			return getTreeHeight(root.Left , log , depth+1 )
		}
		return getTreeHeight(root.Right , log , depth+1 )
	}


	return depth, -1

}


// step 5 ,compute E(h(x)) , the avg height of the tree for every sample , avarge of h(x) for every sample ! 
func compute_avg_height(iforest []*TreeNode , log *models.AnomalyLog) float64{
	sum := 0.0
	for i:= range iforest{
		h , _ := getTreeHeight(iforest[i] , log , 1 )
		sum += float64(h)
	}

	return sum/float64(len(iforest))
}



// step 5 , compute c(n) normaliztion factor for the E(h(x)) for score to be between 0 -1 
func cFactor(n int) float64 {
	if n <= 1 {
		return 0
	}
	return 2*(math.Log(float64(n-1)) + 0.5772) - (2.0*float64(n-1))/float64(n)
}

// step 5 , give anomaly score for each sample
func compute_anomaly_score(cn float64 , avg_height float64) float64 {
	if cn == 0 {
		return 1 
	}
	return math.Pow(2, -avg_height/cn)
}






