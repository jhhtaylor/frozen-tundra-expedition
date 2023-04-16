package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
)

type PartyRole string

const (
	Scout    PartyRole = "Scout"
	Healer   PartyRole = "Healer"
	Gatherer PartyRole = "Gatherer"
)

type Map struct {
	StepAllowance   int
	ResourceInfo    []ResourceInfo
	Quota           []int
	QuotaMultiplier int
	MapSize         [2]int
	Grid            [][]string
}

type ResourceInfo struct {
	Type       string
	Occurrence int
	Locations  [][2]int
}

type Submission struct {
	Party []PartyRole `json:"Party"`
	Path  [][2]int    `json:"Path"`
}

func main() {
	// Read map data from file
	data, err := ioutil.ReadFile("map.txt")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Parse map data
	m, err := parseMap(string(data))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Choose a party
	party := []PartyRole{Scout, Gatherer}

	// Find the optimal path
	path := findPath(m, party)

	// Generate the submission
	submission := Submission{
		Party: party,
		Path:  path,
	}

	// Print the submission in JSON format
	jsonBytes, err := json.MarshalIndent(submission, "", "  ")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println(string(jsonBytes))
}

func parseMap(data string) (Map, error) {
	lines := strings.Split(data, "\n")

	stepAllowanceStr := strings.Split(lines[0], "=")
	stepAllowance, err := strconv.Atoi(stepAllowanceStr[1])
	if err != nil {
		return Map{}, fmt.Errorf("failed to parse step allowance: %w", err)
	}

	resourceInfoCount, err := strconv.Atoi(lines[1][strings.Index(lines[1], ",")+1:])
	if err != nil {
		return Map{}, fmt.Errorf("failed to parse resource info count: %w", err)
	}

	resourceInfo := make([]ResourceInfo, resourceInfoCount)
	for i := 0; i < resourceInfoCount; i++ {
		parts := strings.Split(lines[2+i], ",")
		resourceType := parts[0]
		occurrence, err := strconv.Atoi(parts[1])
		if err != nil {
			return Map{}, fmt.Errorf("failed to parse occurrence: %w", err)
		}

		locations := make([][2]int, occurrence)
		for j := 0; j < occurrence; j++ {
			locationParts := strings.Split(parts[2+j], ",")
			x, err := strconv.Atoi(locationParts[0])
			if err != nil {
				return Map{}, fmt.Errorf("failed to parse location x: %w", err)
			}
			y, err := strconv.Atoi(locationParts[1])
			if err != nil {
				return Map{}, fmt.Errorf("failed to parse location y: %w", err)
			}
			locations[j] = [2]int{x, y}
		}

		resourceInfo[i] = ResourceInfo{
			Type:       resourceType,
			Occurrence: occurrence,
			Locations:  locations,
		}
	}

	quotaLine := 2 + resourceInfoCount
	quotaParts := strings.Split(lines[quotaLine], "=")[1]
	quotaStr := strings.Split(quotaParts, ",")
	quota := make([]int, len(quotaStr))
	for i, part := range quotaStr {
		value, err := strconv.Atoi(part)
		if err != nil {
			return Map{}, fmt.Errorf("failed to parse quota: %w", err)
		}
		quota[i] = value
	}

	quotaMultStr := strings.Split(lines[quotaLine+1], "=")
	quotaMult, err := strconv.Atoi(quotaMultStr[1])
	if err != nil {
		return Map{}, fmt.Errorf("failed to parse quota multiplier: %w", err)
	}

	mapSizeStr := strings.Split(lines[quotaLine+2], "=")[1]
	mapSizeParts := strings.Split(mapSizeStr, "x")
	mapSize := [2]int{}
	for i, part := range mapSizeParts {
		value, err := strconv.Atoi(part)
		if err != nil {
			return Map{}, fmt.Errorf("failed to parse map size: %w", err)
		}
		mapSize[i] = value
	}

	grid := make([][]string, mapSize[0])
	for i := 0; i < mapSize[0]; i++ {
		grid[i] = strings.Split(lines[quotaLine+3+i], "")
	}

	return Map{
		StepAllowance:   stepAllowance,
		ResourceInfo:    resourceInfo,
		Quota:           quota,
		QuotaMultiplier: quotaMult,
		MapSize:         mapSize,
		Grid:            grid,
	}, nil
}

type state struct {
	position [2]int
	steps    int
	resCount []int
	path     [][2]int
	health   int
}

func findPath(m Map, party []PartyRole) [][2]int {
	start := [2]int{0, 0}
	visited := make(map[[2]int]bool)
	roleEffects := make(map[string]int)
	healerPresent := false

	for _, role := range party {
		effectMultiplier, ok := roleEffects[string(role)]
		if !ok {
			effectMultiplier = 1
		}
		roleEffects[string(role)] = effectMultiplier

		if role == Healer {
			healerPresent = true
		}
	}

	var bestPath [][2]int
	var bestScore int

	var dfs func(curr state) int
	dfs = func(curr state) int {
		if curr.steps > m.StepAllowance {
			return 0
		}

		visited[curr.position] = true
		defer func() { visited[curr.position] = false }()

		score := 0
		for i, count := range curr.resCount {
			effectMultiplier, ok := roleEffects[m.ResourceInfo[i].Type]
			if !ok {
				effectMultiplier = 1
			}
			score += count * m.Quota[i] * effectMultiplier
		}

		for _, neighbor := range getNeighbors(curr.position, m, visited) {
			newHealth := curr.health - 1
			if healerPresent {
				newHealth = min(newHealth+1, 10)
			}
			if newHealth > 0 {
				newState := curr.move(neighbor, m, roleEffects)
				newState.health = newHealth
				newScore := dfs(newState)
				if newScore > bestScore {
					bestScore = newScore
					bestPath = newState.path
				}
			}
		}
		return bestScore
	}

	dfs(state{
		position: start,
		steps:    0,
		resCount: make([]int, len(m.ResourceInfo)),
		path:     [][2]int{start},
		health:   10,
	})

	return bestPath
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func getNeighbors(pos [2]int, m Map, visited map[[2]int]bool) [][2]int {
	directions := [][2]int{
		{-1, 0}, // Up
		{1, 0},  // Down
		{0, -1}, // Left
		{0, 1},  // Right
	}
	neighbors := [][2]int{}

	for _, d := range directions {
		newPos := [2]int{pos[0] + d[0], pos[1] + d[1]}
		if newPos[0] >= 0 && newPos[0] < len(m.Grid) && newPos[1] >= 0 && newPos[1] < len(m.Grid[0]) && !visited[newPos] {
			neighbors = append(neighbors, newPos)
		}
	}

	return neighbors
}

func (s state) move(newPos [2]int, m Map, roleEffects map[string]int) state {
	newState := state{
		position: newPos,
		steps:    s.steps + 1,
		resCount: make([]int, len(s.resCount)),
		path:     append(s.path, newPos),
	}

	copy(newState.resCount, s.resCount)

	cellStr := m.Grid[newPos[0]][newPos[1]]
	cell, err := strconv.Atoi(cellStr)
	if err != nil {
		fmt.Printf("Error converting cell string to integer: %v\n", err)
		return newState
	}

	if cell > 0 {
		effectMultiplier := roleEffects[m.ResourceInfo[cell-1].Type]
		newState.resCount[cell-1] += effectMultiplier
	}

	return newState
}

func calculateScore(path [][2]int, m Map, party []PartyRole) int {
	roleEffects := make(map[string]int)

	for _, role := range party {
		switch role {
		case Scout:
			roleEffects["Scout"] = 2
		case Healer:
			roleEffects["Healer"] = 1 // You may need to adjust this value based on the Healer's effect
		case Gatherer:
			roleEffects["Gatherer"] = 3
		}
	}

	travelScore := 0
	resourceCounts := make([]int, len(m.ResourceInfo))

	validCharPattern := regexp.MustCompile(`^[0-9]$`)
	for i := 0; i < len(path)-1; i++ {
		currPos := path[i]
		nextPos := path[i+1]

		// Check if the cell string is a valid character
		if !validCharPattern.MatchString(m.Grid[currPos[0]][currPos[1]]) {
			fmt.Printf("Invalid character encountered: %s\n", m.Grid[currPos[0]][currPos[1]])
			continue
		}

		// Calculate travel score
		difficulty, err := strconv.Atoi(m.Grid[currPos[0]][currPos[1]])
		if err != nil {
			fmt.Printf("Error converting cell string to integer: %v\n", err)
			continue
		}
		travelScore += difficulty

		// Update resource counts
		resourceType := m.Grid[nextPos[0]][nextPos[1]]
		for i, resInfo := range m.ResourceInfo {
			if resInfo.Type == resourceType {
				resourceCounts[i]++
				break
			}
		}
	}

	resourceScore := 0
	for i, count := range resourceCounts {
		roleEffect, ok := roleEffects[m.ResourceInfo[i].Type]
		if !ok {
			roleEffect = 1
		}
		resourceScore += count * m.Quota[i] * roleEffect
	}

	totalScore := resourceScore - travelScore
	return totalScore
}
