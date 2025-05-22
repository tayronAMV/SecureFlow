package logic
import (

	"fmt"
	"server/internal/db/models"
	"server/internal/api/handlers"
	

)
const ANOMALY_THRESHOLD = 0.6

func Anomaly_detection(arr []*models.AnomalyLog){

	cleaned_arr , ok := Data_Cleaning(arr)

	if !ok{
		fmt.Println("the data is not good for algorithm , waiting for next round of samples ")
		return 
	} 

	forest := BuildForest(cleaned_arr ,NUM_TREES_IN_FOREST)

	var suspicous_samples = make(map[*models.AnomalyLog]int)

	for _ , sample := range cleaned_arr {
		avg_height , reason := Compute_avg_height(forest , sample)
		score := Compute_anomaly_score(CFactor(NUM_TREES_IN_FOREST),avg_height)

		if score >= ANOMALY_THRESHOLD {
			suspicous_samples[sample] = reason
		}

	}

	for sample ,_:= range suspicous_samples{
		kube_client , err:= handlers.GetKubernetesClient()

		if err != nil {
			fmt.Println("some is wrong with generating kube client " , err)
			return
		}

		container_logs  , err:= handlers.FetchPodLogs(kube_client , sample.Info.Namespace , sample.Info.PodName , sample.Info.Container_name,sample.Timestamp)

		if err != nil{
			fmt.Printf("problem fetching container logs from %s\n err = %v\n" , sample.Info.PodName , err)
			return
		}

		clean_container_logs := CleanRawLogs(container_logs)

		structured_logs := ParseLogs(clean_container_logs)

		items := BuildItemsets(structured_logs)

		frequent_activity := RunApriori(items)

		
	}




	
	
}
