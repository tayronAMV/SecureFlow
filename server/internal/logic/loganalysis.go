package logic
import (
	"regexp"
	"strings"
	"server/internal/db/models"
)

// Compile filters once for performance
var ignorePatterns = []*regexp.Regexp{
	regexp.MustCompile(`\.(js|css|png|jpg|ico|svg)\b`),
	regexp.MustCompile(`/health|/ready|/live|liveness-probe|readiness-probe`),
	regexp.MustCompile(`^\s*$`), // empty line
}

// CleanRawLogs removes irrelevant lines and trims whitespace - step 1 of just normalize the logs 
func CleanRawLogs(raw string) []string {
	lines := strings.Split(raw, "\n")
	cleaned := []string{}

	for _, line := range lines {
		shouldIgnore := false
		for _, re := range ignorePatterns {
			if re.MatchString(line) {
				shouldIgnore = true
				break
			}
		}
		if !shouldIgnore {
			cleaned = append(cleaned, strings.TrimSpace(line))
		}
	}
	return cleaned
}

var (
	logRegex = regexp.MustCompile(`^(?P<time>\S+)\s+\S+\s+\S+\s+\S+\s+\[[^\]]+\]\s+"(?P<method>GET|POST|PUT|DELETE|PATCH|HEAD|OPTIONS)\s+(?P<path>\S+)[^"]*"\s+(?P<status>\d{3})`)
	idRegex  = regexp.MustCompile(`\d+|[a-fA-F0-9\-]{8,}`) // match digits or UUIDs
)

// Main parsing function - need to add regex - only for nginx
func ParseLogs(lines []string) []models.LogItem {
	var parsed []models.LogItem

	for _, line := range lines {
		matches := logRegex.FindStringSubmatch(line)
		if len(matches) == 0 {
			continue
		}

		timestamp := matches[1]
		method := strings.ToUpper(matches[2])
		rawPath := matches[3]
		status := matches[4]

		path := normalizePath(rawPath)
		op := classifyOperation(method)
		target := extractTarget(path)
		itemType := classifyType(path)

		parsed = append(parsed, models.LogItem{
			Timestamp: timestamp,
			Method:    method,
			Path:      path,
			Status:    status,
			Operation: op,
			Target:    target,
			Type:      itemType,
			Raw:       line,
		})
	}

	return parsed
}

// Replace numbers or UUIDs with {id}
func normalizePath(p string) string {
	parts := strings.Split(p, "/")
	for i, part := range parts {
		if idRegex.MatchString(part) {
			parts[i] = "{id}"
		}
	}
	return strings.ToLower(strings.Join(parts, "/"))
}

// Simple method-to-operation mapping
func classifyOperation(method string) string {
	switch method {
	case "GET":
		return "read"
	case "POST":
		return "write"
	case "PUT", "PATCH":
		return "update"
	case "DELETE":
		return "delete"
	default:
		return "unknown"
	}
}

// Extract the second segment as target (e.g., "user" in /api/user/{id})
func extractTarget(path string) string {
	parts := strings.Split(path, "/")
	for _, part := range parts {
		if part != "" && part != "api" && part != "{id}" {
			return part
		}
	}
	return "unknown"
}

// Basic type classification
func classifyType(path string) string {
	if strings.HasPrefix(path, "/api") {
		return "api_call"
	}
	return "static"
}

func BuildItemsets(logs []models.LogItem) [][]string {
	var transactions [][]string
	for _, log := range logs {
		transaction := []string{
			log.Type,        // "api_call"
			log.Operation,   // "read", "write", etc.
			log.Target,      // e.g., "user"
			log.Path,        // normalized path like "/user/{id}"
		}
		transactions = append(transactions, transaction)
	}
	return transactions
}

func Count1Itemsets(transactions [][]string, minSupport int) map[string]int {
	counts := map[string]int{}
	for _, tx := range transactions {
		seen := map[string]bool{}
		for _, item := range tx {
			if !seen[item] {
				counts[item]++
				seen[item] = true
			}
		}
	}
	// Filter by support
	for k, v := range counts {
		if v < minSupport {
			delete(counts, k)
		}
	}
	return counts
}

func CountCandidates(candidates [][]string, transactions [][]string, minSupport int) [][]string {
	result := [][]string{}
	for _, c := range candidates {
		count := 0
		for _, tx := range transactions {
			if containsAll(tx, c) {
				count++
			}
		}
		if count >= minSupport {
			result = append(result, c)
		}
	}
	return result
}


func RunApriori(transactions [][]string) [][]string {
	minSupport := int(0.2 * float64(len(transactions)))
	var result [][]string
	L1 := Count1Itemsets(transactions, minSupport)
	Lk := [][]string{}
	for item := range L1 {
		Lk = append(Lk, []string{item})
		result = append(result, []string{item})
	}

	k := 2
	for len(Lk) > 0 { 
		Ck := GenerateCandidates(Lk, k)
		Lk = CountCandidates(Ck, transactions, minSupport)
		result = append(result, Lk...)
		k++
	}
	return result
}

func mergeIfCompatible(a, b []string, k int) []string {
	set := make(map[string]bool)
	for _, item := range a {
		set[item] = true
	}
	for _, item := range b {
		set[item] = true
	}
	if len(set) == k {
		out := []string{}
		for item := range set {
			out = append(out, item)
		}
		return out
	}
	return nil
}

func containsAll(tx, items []string) bool {
	m := map[string]bool{}
	for _, x := range tx {
		m[x] = true
	}
	for _, y := range items {
		if !m[y] {
			return false
		}
	}
	return true
}

func GenerateCandidates(prev [][]string, k int) [][]string {
	candidates := [][]string{}
	n := len(prev)

	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			candidate := mergeIfCompatible(prev[i], prev[j], k)
			if candidate != nil {
				candidates = append(candidates, candidate)
			}
		}
	}
	return candidates
}

