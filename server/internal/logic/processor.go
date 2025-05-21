package logic
import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	
	"server/internal/db/models"
)
func LoadAnomalyLogsFromFile(filePath string) ([]*models.AnomalyLog, error) {
	var logs []*models.AnomalyLog

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var log models.AnomalyLog
		line := scanner.Text()
		err := json.Unmarshal([]byte(line), &log)
		if err != nil {
			fmt.Printf("❌ Failed to parse line: %s\n", line)
			continue
		}
		logs = append(logs, &log)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	return logs, nil
}



