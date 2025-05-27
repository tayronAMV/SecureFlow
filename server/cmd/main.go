package main

import (
	// "server/internal/db/models"
	"time"
	"server/internal/api/handlers"
	"fmt"
	"server/internal/logic"

	
)

func main(){
	namespace := "default"
	podname := "demo-test"
	t := time.Now()

	containerName := "nginx"


	kube_client , err:= handlers.GetKubernetesClient()

	if err != nil{
		fmt.Println( err)
		return
	}

	logs ,err := handlers.FetchPodLogs(kube_client , namespace , podname , containerName , t)

	if err != nil{
		fmt.Println(" oh no 2lv  , " ,err)
		return
	}
	

	logs_arr := logic.CleanRawLogs(logs)
	
	struct_logs := logic.ParseLogs(logs_arr)

	items := logic.BuildItemsets(struct_logs)

	res := logic.RunApriori(items )

	

	
	


}
	

