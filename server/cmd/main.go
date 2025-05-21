package main

import (
	// "server/internal/db/models"
	"server/internal/logic"
	"fmt"
	
)

func main(){
	logs , err := logic.LoadAnomalyLogsFromFile("/home/amitay2005/Documents/SecureFlow/server/internal/logic/random.txt")
	if err != nil {
		fmt.Println("oh no ! ", err)
		return 
	}

	//step 1 check 

	if logs , ok := logic.Data_Cleaning(logs) ; !ok{
		fmt.Println("on god some shit is bad")
	}else{
		fmt.Println(logs)
	}


	forest := logic.BuildForest(logs ,50 )

	for _,log := range logs{
		avg ,bad_feature := logic.Compute_avg_height(forest,log)
		anomaly_score  := logic.Compute_anomaly_score(logic.CFactor(len(logs)),avg)
		fmt.Println("this is one : " , log)
		fmt.Println("this its anomaly score : ", anomaly_score)
		fmt.Println("and the fucking reason is  ", bad_feature)



	}



}


func Print_Tree(root *logic.TreeNode) {
	if root == nil {
		return
	}
	fmt.Printf("this is val %v , this is threshold %v , to feature %v ",root.Val , root.Threshold , root.Feature_of_threshold)
	Print_Tree(root.Right)
	Print_Tree(root.Left)

}