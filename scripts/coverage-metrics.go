package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/mod/modfile"
)

// FunctionData holds coverage and complexity information for a function
type FunctionData struct {
	Package    string
	File       string
	Line       int
	Function   string
	Coverage   float64
	Complexity int
}

// Metrics holds the calculated metrics
type Metrics struct {
	TotalFunctions         int
	StatementCoverage      float64
	AvgFunctionCoverage    float64
	MedianFunctionCoverage float64
	HighComplexityCoverage float64
	MaxComplexity          int
	MinComplexity          int
	AvgComplexity          float64
	MedianComplexity       float64
	TotalRiskScore         float64
	AvgRiskPerFunction     float64
	HighRiskFunctions      int // Functions with risk > 100
	HighComplexityCount    int
	TopRiskyFunctions      []*FunctionData
}

func main() {
	coverageFile := flag.String("coverage", ".coverage/coverage-by-function.txt", "Coverage output file")
	complexityFile := flag.String("complexity", ".coverage/complexity.txt", "Cyclomatic complexity output file")
	complexityThreshold := flag.Int("threshold", 7, "Complexity threshold for high-complexity coverage")
	riskThreshold := flag.Float64("risk-threshold", 200.0, "Risk threshold to count high-risk functions")
	outputFormat := flag.String("format", "text", "Output format: text, json, or markdown")
	modfilePath := flag.String("modfile", "go.mod", "Path to go.mod file for module path")
	outputFile := flag.String("output", "stdout", "Output file (default: stdout)")
	flag.Parse()

	// Parse coverage data
	coverageData, totalCoverage, err := parseCoverage(*coverageFile, *modfilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing coverage: %v\n", err)
		os.Exit(1)
	}

	// Parse complexity data
	complexityData, err := parseComplexity(*complexityFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing complexity: %v\n", err)
		os.Exit(1)
	}

	// Combine data
	functions := combineData(coverageData, complexityData)

	// Calculate metrics
	metrics := calculateMetrics(functions, *complexityThreshold, *riskThreshold)
	metrics.StatementCoverage = totalCoverage

	// Output results
	var output string
	switch *outputFormat {
	case "json":
		output = formatJSON(metrics, *complexityThreshold)
	case "markdown":
		output = formatMarkdown(metrics, *complexityThreshold)
	default:
		output = formatText(metrics, *complexityThreshold, *riskThreshold)
	}

	if *outputFile != "stdout" {
		err := os.WriteFile(*outputFile, []byte(output), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Print(output)
	}
}

func parseCoverage(filename string, modfilePath string) (map[string]*FunctionData, float64, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, 0, err
	}
	defer func() {
		_ = file.Close()
	}()

	data := make(map[string]*FunctionData)
	var totalCoverage float64
	scanner := bufio.NewScanner(file)

	// Match: github.com/McTalian/wow-build-tools/internal/build/build.go:58:	Build	56.0%
	re := regexp.MustCompile(`^(.+):(\d+):\s+(\S+)\s+(\d+\.?\d*)%`)
	totalRe := regexp.MustCompile(`^total:\s+\(statements\)\s+(\d+\.?\d*)%`)

	// Read go.mod to get module path
	modBytes, err := os.ReadFile(modfilePath)
	if err != nil {
		return nil, 0, fmt.Errorf("error reading go.mod file: %v", err)
	}
	modulePath := modfile.ModulePath(modBytes)

	for scanner.Scan() {
		line := scanner.Text()

		// Check for total line
		if totalMatches := totalRe.FindStringSubmatch(line); len(totalMatches) == 2 {
			totalCoverage, _ = strconv.ParseFloat(totalMatches[1], 64)
			continue
		}

		matches := re.FindStringSubmatch(line)
		if len(matches) == 5 {
			filePath := matches[1]
			lineNum, _ := strconv.Atoi(matches[2])
			funcName := matches[3]
			coverage, _ := strconv.ParseFloat(matches[4], 64)

			// Normalize file path to be relative to module
			if strings.HasPrefix(filePath, modulePath) {
				filePath = strings.TrimPrefix(filePath, modulePath)
				filePath = strings.TrimPrefix(filePath, "/")
			}

			parts := strings.Split(filePath, "/")
			var pkg, file string
			if len(parts) > 0 {
				file = parts[len(parts)-1]
				pkg = strings.Join(parts[:len(parts)-1], "/")
			}

			key := fmt.Sprintf("%s:%d:%s", filePath, lineNum, funcName)
			data[key] = &FunctionData{
				Package:  pkg,
				File:     file,
				Line:     lineNum,
				Function: funcName,
				Coverage: coverage,
			}
		}
	}

	return data, totalCoverage, scanner.Err()
}

func parseComplexity(filename string) (map[string]int, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = file.Close()
	}()

	data := make(map[string]int)
	scanner := bufio.NewScanner(file)

	// Match: 73 build Build internal/build/build.go:58:1
	re := regexp.MustCompile(`^(\d+)\s+\S+\s+(\S+)\s+(.+):(\d+):`)

	for scanner.Scan() {
		line := scanner.Text()
		matches := re.FindStringSubmatch(line)
		if len(matches) == 5 {
			complexity, _ := strconv.Atoi(matches[1])
			funcName := matches[2]
			if strings.HasPrefix(funcName, "(") {
				// Method, extract method name
				parts := strings.Split(funcName, ").")
				funcName = parts[len(parts)-1]
			}
			filePath := matches[3]
			lineNum := matches[4]

			key := fmt.Sprintf("%s:%s:%s", filePath, lineNum, funcName)
			data[key] = complexity
		}
	}

	return data, scanner.Err()
}

func combineData(coverageData map[string]*FunctionData, complexityData map[string]int) []*FunctionData {
	var functions []*FunctionData

	for key, funcData := range coverageData {
		// Try to find matching complexity data
		// Key format: full/path/file.go:line:funcName
		complexity := 1 // Default to 1 if not found

		if c, ok := complexityData[key]; ok {
			complexity = c
		} else {
			// Try alternate key formats
			parts := strings.Split(key, "/")
			if len(parts) > 0 {
				// Try with just the relative path
				for i := 1; i < len(parts); i++ {
					altKey := strings.Join(parts[i:], "/")
					if c, ok := complexityData[altKey]; ok {
						complexity = c
						break
					}
				}
			}
		}

		funcData.Complexity = complexity
		functions = append(functions, funcData)
	}

	return functions
}

func calculateMetrics(functions []*FunctionData, complexityThreshold int, riskThreshold float64) Metrics {
	var metrics Metrics
	var totalFunctionCoverage float64
	var highComplexityCoverage float64
	var highComplexityCount int
	var totalRiskScore float64
	var highRiskCount int

	// For calculating risk per function
	type functionRisk struct {
		function *FunctionData
		risk     float64
	}
	var allRisks []functionRisk

	metrics.TotalFunctions = len(functions)

	// Collect coverage values for median calculation
	coverages := make([]float64, len(functions))
	complexities := make([]int, len(functions))
	var totalComplexity int

	for i, f := range functions {
		coverages[i] = f.Coverage
		complexities[i] = f.Complexity
		totalComplexity += f.Complexity

		// Track min/max complexity
		if i == 0 || f.Complexity > metrics.MaxComplexity {
			metrics.MaxComplexity = f.Complexity
		}
		if i == 0 || f.Complexity < metrics.MinComplexity {
			metrics.MinComplexity = f.Complexity
		}

		// Average function coverage
		totalFunctionCoverage += f.Coverage

		// High-complexity coverage
		if f.Complexity > complexityThreshold {
			highComplexityCoverage += f.Coverage
			highComplexityCount++
		}

		// Risk score: (100 - coverage) * complexity
		uncoveredPercent := 100.0 - f.Coverage
		risk := uncoveredPercent * float64(f.Complexity)
		totalRiskScore += risk

		allRisks = append(allRisks, functionRisk{function: f, risk: risk})

		// Count high-risk functions
		if risk > riskThreshold {
			highRiskCount++
		}
	}

	if metrics.TotalFunctions > 0 {
		metrics.AvgFunctionCoverage = totalFunctionCoverage / float64(metrics.TotalFunctions)
		metrics.AvgRiskPerFunction = totalRiskScore / float64(metrics.TotalFunctions)
		metrics.AvgComplexity = float64(totalComplexity) / float64(metrics.TotalFunctions)

		// Calculate median coverage
		sort.Float64s(coverages)
		mid := len(coverages) / 2
		if len(coverages)%2 == 0 {
			metrics.MedianFunctionCoverage = (coverages[mid-1] + coverages[mid]) / 2.0
		} else {
			metrics.MedianFunctionCoverage = coverages[mid]
		}

		// Calculate median complexity
		sort.Ints(complexities)
		if len(complexities)%2 == 0 {
			metrics.MedianComplexity = float64(complexities[mid-1]+complexities[mid]) / 2.0
		} else {
			metrics.MedianComplexity = float64(complexities[mid])
		}
	}

	if highComplexityCount > 0 {
		metrics.HighComplexityCoverage = highComplexityCoverage / float64(highComplexityCount)
		metrics.HighComplexityCount = highComplexityCount
	}

	metrics.TotalRiskScore = totalRiskScore
	metrics.HighRiskFunctions = highRiskCount

	// Sort by risk and get top 10
	sort.Slice(allRisks, func(i, j int) bool {
		return allRisks[i].risk > allRisks[j].risk
	})

	topN := min(len(allRisks), 10)

	for i := 0; i < topN; i++ {
		metrics.TopRiskyFunctions = append(metrics.TopRiskyFunctions, allRisks[i].function)
	}

	return metrics
}

func formatText(metrics Metrics, threshold int, riskThreshold float64) string {
	var sb strings.Builder
	sb.WriteString("Coverage Metrics Report\n")
	sb.WriteString("=======================\n")
	sb.WriteString(fmt.Sprintf("Total Functions: %d\n", metrics.TotalFunctions))
	sb.WriteString("\n")

	sb.WriteString("COVERAGE METRICS\n")
	sb.WriteString("----------------\n")
	sb.WriteString(fmt.Sprintf("Statement Coverage:         %.1f%% (total statements covered)\n", metrics.StatementCoverage))
	sb.WriteString(fmt.Sprintf("Mean Function Coverage:     %.1f%% (average across all functions)\n", metrics.AvgFunctionCoverage))
	sb.WriteString(fmt.Sprintf("Median Function Coverage:   %.1f%% (typical function)\n", metrics.MedianFunctionCoverage))
	sb.WriteString(fmt.Sprintf("High-Complexity Coverage:   %.1f%% (complexity > %d, %d functions)\n",
		metrics.HighComplexityCoverage, threshold, metrics.HighComplexityCount))
	sb.WriteString("\n")

	sb.WriteString("COMPLEXITY METRICS\n")
	sb.WriteString("------------------\n")
	sb.WriteString(fmt.Sprintf("Max Complexity:             %d\n", metrics.MaxComplexity))
	sb.WriteString(fmt.Sprintf("Mean Complexity:            %.1f\n", metrics.AvgComplexity))
	sb.WriteString(fmt.Sprintf("Median Complexity:          %.1f\n", metrics.MedianComplexity))
	sb.WriteString(fmt.Sprintf("Min Complexity:             %d\n", metrics.MinComplexity))
	sb.WriteString("\n")

	sb.WriteString("RISK METRICS\n")
	sb.WriteString("------------\n")
	sb.WriteString("Risk = (100 - coverage%) × complexity\n")
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("Total Risk Score:           %.1f\n", metrics.TotalRiskScore))
	sb.WriteString(fmt.Sprintf("Avg Risk Per Function:      %.1f (normalized for codebase size)\n", metrics.AvgRiskPerFunction))
	sb.WriteString(fmt.Sprintf("High-Risk Functions:        %d (risk > %.1f)\n", metrics.HighRiskFunctions, riskThreshold))
	sb.WriteString("\n")

	sb.WriteString("Top 10 Riskiest Functions:\n")
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("%-8s %-12s %-8s %s\n", "Risk", "Coverage", "Complexity", "Function"))
	sb.WriteString(strings.Repeat("-", 80) + "\n")

	for _, f := range metrics.TopRiskyFunctions {
		risk := (100.0 - f.Coverage) * float64(f.Complexity)
		funcPath := fmt.Sprintf("%s:%d %s", f.File, f.Line, f.Function)
		sb.WriteString(fmt.Sprintf("%-8.1f %-12s %-10d %s\n",
			risk,
			fmt.Sprintf("%.1f%%", f.Coverage),
			f.Complexity,
			funcPath))
	}

	return sb.String()
}

func formatJSON(metrics Metrics, threshold int) string {
	// Create a JSON-friendly structure
	type RiskyFunction struct {
		File       string  `json:"file"`
		Line       int     `json:"line"`
		Function   string  `json:"function"`
		Coverage   float64 `json:"coverage"`
		Complexity int     `json:"complexity"`
		Risk       float64 `json:"risk"`
	}

	type JSONOutput struct {
		TotalFunctions         int             `json:"total_functions"`
		StatementCoverage      float64         `json:"statement_coverage"`
		MeanFunctionCoverage   float64         `json:"mean_function_coverage"`
		MedianFunctionCoverage float64         `json:"median_function_coverage"`
		HighComplexityCoverage float64         `json:"high_complexity_coverage"`
		ComplexityThreshold    int             `json:"complexity_threshold"`
		HighComplexityCount    int             `json:"high_complexity_count"`
		MaxComplexity          int             `json:"max_complexity"`
		MinComplexity          int             `json:"min_complexity"`
		AvgComplexity          float64         `json:"avg_complexity"`
		MedianComplexity       float64         `json:"median_complexity"`
		TotalRiskScore         float64         `json:"total_risk_score"`
		AvgRiskPerFunction     float64         `json:"avg_risk_per_function"`
		HighRiskFunctions      int             `json:"high_risk_functions"`
		TopRiskyFunctions      []RiskyFunction `json:"top_risky_functions"`
	}

	var riskyFuncs []RiskyFunction
	for _, f := range metrics.TopRiskyFunctions {
		risk := (100.0 - f.Coverage) * float64(f.Complexity)
		riskyFuncs = append(riskyFuncs, RiskyFunction{
			File:       f.File,
			Line:       f.Line,
			Function:   f.Function,
			Coverage:   f.Coverage,
			Complexity: f.Complexity,
			Risk:       risk,
		})
	}

	output := JSONOutput{
		TotalFunctions:         metrics.TotalFunctions,
		StatementCoverage:      metrics.StatementCoverage,
		MeanFunctionCoverage:   metrics.AvgFunctionCoverage,
		MedianFunctionCoverage: metrics.MedianFunctionCoverage,
		HighComplexityCoverage: metrics.HighComplexityCoverage,
		ComplexityThreshold:    threshold,
		HighComplexityCount:    metrics.HighComplexityCount,
		MaxComplexity:          metrics.MaxComplexity,
		MinComplexity:          metrics.MinComplexity,
		AvgComplexity:          metrics.AvgComplexity,
		MedianComplexity:       metrics.MedianComplexity,
		TotalRiskScore:         metrics.TotalRiskScore,
		AvgRiskPerFunction:     metrics.AvgRiskPerFunction,
		HighRiskFunctions:      metrics.HighRiskFunctions,
		TopRiskyFunctions:      riskyFuncs,
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error marshaling JSON: %v", err)
	}
	return string(data)
}

func formatMarkdown(metrics Metrics, threshold int) string {
	var sb strings.Builder

	sb.WriteString("# 📊 Coverage Metrics Report\n\n")

	// Summary badges
	sb.WriteString("## Summary\n\n")
	sb.WriteString(fmt.Sprintf("**Total Functions:** %d\n\n", metrics.TotalFunctions))

	// Coverage Metrics Table
	sb.WriteString("### Coverage Metrics\n\n")
	sb.WriteString("| Metric | Value | Description |\n")
	sb.WriteString("|--------|-------|-------------|\n")
	sb.WriteString(fmt.Sprintf("| **Statement Coverage** | %.1f%% | Total statements covered |\n", metrics.StatementCoverage))
	sb.WriteString(fmt.Sprintf("| **Mean Function Coverage** | %.1f%% | Average across all functions |\n", metrics.AvgFunctionCoverage))
	sb.WriteString(fmt.Sprintf("| **Median Function Coverage** | %.1f%% | Typical function |\n", metrics.MedianFunctionCoverage))
	sb.WriteString(fmt.Sprintf("| **High-Complexity Coverage** | %.1f%% | Functions with complexity > %d (%d functions) |\n",
		metrics.HighComplexityCoverage, threshold, metrics.HighComplexityCount))
	sb.WriteString("\n")

	// Complexity Metrics Table
	sb.WriteString("### Complexity Metrics\n\n")
	sb.WriteString("| Metric | Value | Description |\n")
	sb.WriteString("|--------|-------|-------------|\n")
	sb.WriteString(fmt.Sprintf("| **Max Complexity** | %d | Most complex function |\n", metrics.MaxComplexity))
	sb.WriteString(fmt.Sprintf("| **Mean Complexity** | %.1f | Average complexity across all functions |\n", metrics.AvgComplexity))
	sb.WriteString(fmt.Sprintf("| **Median Complexity** | %.1f | Typical function complexity |\n", metrics.MedianComplexity))
	sb.WriteString(fmt.Sprintf("| **Min Complexity** | %d | Simplest function |\n", metrics.MinComplexity))
	sb.WriteString("\n")

	// Risk Metrics Table
	sb.WriteString("### Risk Metrics\n\n")
	sb.WriteString("_Risk = (100 - coverage%) × complexity_\n\n")
	sb.WriteString("| Metric | Value | Description |\n")
	sb.WriteString("|--------|-------|-------------|\n")
	sb.WriteString(fmt.Sprintf("| **Total Risk Score** | %.1f | Sum of all function risks |\n", metrics.TotalRiskScore))
	sb.WriteString(fmt.Sprintf("| **Avg Risk Per Function** | %.1f | Normalized for codebase size |\n", metrics.AvgRiskPerFunction))
	sb.WriteString(fmt.Sprintf("| **High-Risk Functions** | %d | Functions with risk > 100 |\n", metrics.HighRiskFunctions))
	sb.WriteString("\n")

	// Top 10 Riskiest Functions
	sb.WriteString("#### 🔴 Top 10 Riskiest Functions\n\n")
	sb.WriteString("| Risk | Coverage | Complexity | Function |\n")
	sb.WriteString("|------|----------|------------|----------|\n")

	for _, f := range metrics.TopRiskyFunctions {
		risk := (100.0 - f.Coverage) * float64(f.Complexity)
		funcPath := fmt.Sprintf("`%s:%d` **%s**", f.File, f.Line, f.Function)
		sb.WriteString(fmt.Sprintf("| %.1f | %.1f%% | %d | %s |\n",
			risk,
			f.Coverage,
			f.Complexity,
			funcPath))
	}

	sb.WriteString("\n---\n\n")
	sb.WriteString("💡 **Focus testing efforts on high-risk functions** to maximize impact on code quality.\n")

	return sb.String()
}
