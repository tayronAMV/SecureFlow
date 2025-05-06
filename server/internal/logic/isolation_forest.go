package logic 

import (
	"math"
	"math/rand"

)

type sample struct {
	CPU     float64
	DiskIO  float64
	Memory  float64
	Network float64
}

const (
	CPU_MIN     = 0.1
	DiskIO_MIN  = 0.2
	Memory_MIN  = 0.01
	Network_MIN = 0.02
)

type TreeNode struct {
	val   float64
	feat  int // 0=CPU, 1=DiskIO, 2=Memory, 3=Network
	Left  *TreeNode
	Right *TreeNode
}

func max_int(a int , b int ) int {
	if a < b {
		return b
	}
	return a 
}

var Forest []*TreeNode

func weightsCalc(arr []*sample) []int {
	var res [4]float64
	p := float64(len(arr))
	for _, s := range arr {
		if s.CPU > CPU_MIN {
			res[0]++
		}
		if s.DiskIO > DiskIO_MIN {
			res[1]++
		}
		if s.Memory > Memory_MIN {
			res[2]++
		}
		if s.Network > Network_MIN {
			res[3]++
		}
	}

	out := make([]int, 4)
	for i := 0; i < 4; i++ {
		if res[i] == 0 {
			out[i] = 0
		} else {
			out[i] = int(1 + (res[i] / p))
		}
	}
	return out
}

func selectRandomMetric(weights []int) int {
	sum := 0
	for _, w := range weights {
		sum += w
	}
	if sum == 0 {
		return rand.Intn(4) // fallback if all weights are zero
	}

	r := rand.Intn(sum)
	for i := 0; i < 4; i++ {
		r -= weights[i]
		if r < 0 {
			return i
		}
	}
	return 3 // fallback
}

func findFeatureMinMax(arr []*sample, feature int) (float64, float64) {
	if len(arr) == 0 {
		return 0, 0
	}
	get := func(s *sample) float64 {
		switch feature {
		case 0:
			return s.CPU
		case 1:
			return s.DiskIO
		case 2:
			return s.Memory
		case 3:
			return s.Network
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

func buildTree(arr []*sample, depth int) *TreeNode {
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

	left := []*sample{}
	right := []*sample{}

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
		}
		if val < threshold {
			left = append(left, s)
		} else {
			right = append(right, s)
		}
	}

	node := &TreeNode{
		val:  threshold,
		feat: feature,
	}
	node.Left = buildTree(left, depth+1)
	node.Right = buildTree(right, depth+1)
	return node
}

func getHeight(root *TreeNode) int{
	if root == nil{
		return 0
	}

	right := getHeight(root.Right)
	left := getHeight(root.Left)

	return max_int(right , left) + 1 

}


func avg_height_len(n int) float64 {
	if n <= 1 {
		return 0
	}
	h := math.Log(float64(n-1)) + 0.5772156649 // Harmonic number approximation
	return 2*h - (2 * float64(n-1) / float64(n))
}


	
	