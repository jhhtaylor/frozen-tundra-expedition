package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

type Tile struct {
	Type         string
	TravelDiff   int
	Resource     string
	ResourceAmt  int
	ResourceUsed bool
}

type Map struct {
	Size          [2]int
	Tiles         [][]Tile
	StepAllowance int
	//Resources     map[string]int
	Resources     []Resource
	ResourceQuota map[string]int
	QuotaMult     int
}

type Party struct {
	Scout         bool
	Healer        bool
	Gatherer      bool
	StepBonus     int
	ResourceBonus int
}

type Resource struct {
	Name string
	Row  int
	Col  int
}

func (p *Party) SetParty(members []string) {
	p.Scout = false
	p.Healer = false
	p.Gatherer = false
	p.StepBonus = 1
	p.ResourceBonus = 1
	for _, member := range members {
		switch member {
		case "Scout":
			p.Scout = true
		case "Healer":
			p.Healer = true
			p.StepBonus = int(math.Ceil(float64(p.StepBonus) * 1.2))
		case "Gatherer":
			p.Gatherer = true
			p.ResourceBonus = 2
		}
	}
}

func (m *Map) LoadMap(filePath string) {
	data, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error reading map file:", err)
		return
	}
	defer data.Close()
	scanner := bufio.NewScanner(data)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, " ")
		if parts[0] == "Step_allowance" {
			m.StepAllowance = atoi(parts[1])
		} else if parts[0] == "Quota" {
			quotas := strings.Split(parts[1], ",")
			m.ResourceQuota["Coal"] = atoi(quotas[0])
			m.ResourceQuota["Fish"] = atoi(quotas[1])
			m.ResourceQuota["Scrap_metal"] = atoi(quotas[2])
		} else if parts[0] == "Quota_multiplier" {
			m.QuotaMult = atoi(parts[1])
		} else if parts[0] == "map_size" {
			size := strings.Split(parts[1], "x")
			m.Size[0], m.Size[1] = atoi(size[0]), atoi(size[1])
			m.Tiles = make([][]Tile, m.Size[0])
			for i := range m.Tiles {
				m.Tiles[i] = make([]Tile, m.Size[1])
			}
		} else if m.Size[0] > 0 {
			for i := range m.Tiles {
				for j := range m.Tiles[i] {
					m.Tiles[i][j].Type = parts[i*m.Size[1]+j]
					switch m.Tiles[i][j].Type {
					case "S":
						m.Tiles[i][j].TravelDiff = 1
					case "I":
						m.Tiles[i][j].TravelDiff = 5
					case "TS":
						m.Tiles[i][j].TravelDiff = 10
					case "M":
						m.Tiles[i][j].TravelDiff = 15
					}
				}
			}
		} else if strings.Contains(parts[0], ",") {
			resource := strings.Split(parts[0], ",")[0]
			resourceCount := atoi(strings.Split(parts[0], ",")[1])
			for i := 1; i <= resourceCount; i++ {
				coords := strings.Split(parts[i], ",")
				row, col := atoi(coords[0]), atoi(coords[1])
				m.Resources = append(m.Resources, Resource{Name: resource, Row: row, Col: col})
			}
		}
	}
}

// func (m *Map) LoadMap(filePath string) {
// 	file, err := os.Open(filePath)
// 	if err != nil {
// 		fmt.Println("Error reading map file:", err)
// 		return
// 	}
// 	defer file.Close()

// 	scanner := bufio.NewScanner(file)
// 	scanner.Split(bufio.ScanLines)

// 	lines := make([]string, 0)
// 	for scanner.Scan() {
// 		lines = append(lines, scanner.Text())
// 	}

// 	if err := scanner.Err(); err != nil {
// 		fmt.Println("Error reading map file:", err)
// 		return
// 	}

// 	m.Resources = make(map[string]int)
// 	m.ResourceQuota = make(map[string]int)
// 	for _, line := range lines {
// 		if strings.HasPrefix(line, "map_size") {
// 			size := strings.Split(strings.TrimPrefix(line, "map_size="), "x")
// 			m.Size[0], _ = strconv.Atoi(size[0])
// 			m.Size[1], _ = strconv.Atoi(size[1])
// 			m.Tiles = make([][]Tile, m.Size[0])
// 			for i := range m.Tiles {
// 				m.Tiles[i] = make([]Tile, m.Size[1])
// 			}
// 		} else if strings.HasPrefix(line, "Step_allowance") {
// 			m.StepAllowance, _ = strconv.Atoi(strings.TrimPrefix(line, "Step_allowance="))
// 		} else if strings.Contains(line, ",") {
// 			parts := strings.Split(line, " ")
// 			if strings.Contains(parts[0], ",") {
// 				quota := strings.Split(parts[0], ",")
// 				m.ResourceQuota[quota[0]] = atoi(quota[1])
// 				m.ResourceQuota[quota[2]] = atoi(quota[3])
// 				m.ResourceQuota[quota[4]] = atoi(quota[5])
// 			} else if strings.HasPrefix(parts[0], "Quota_multiplier") {
// 				m.QuotaMult, _ = strconv.Atoi(strings.TrimPrefix(parts[0], "Quota_multiplier="))
// 			} else {
// 				resource := strings.Split(parts[0], ",")
// 				m.Resources[resource[0]] = atoi(resource[1])
// 				location := strings.Split(parts[1], ",")
// 				row, _ := strconv.Atoi(location[0])
// 				col, _ := strconv.Atoi(location[1])
// 				m.Tiles[row][col].Resource = resource[0]
// 				m.Tiles[row][col].ResourceAmt = m.Resources[resource[0]]
// 			}
// 		} else if m.Size[0] > 0 {
// 			i := 0
// 			for i < m.Size[0] {
// 				tiles := strings.Split(line, ",")
// 				for j, tile := range tiles {
// 					m.Tiles[i][j].Type = tile
// 					switch tile {
// 					case "S":
// 						m.Tiles[i][j].TravelDiff = 1
// 					case "I":
// 						m.Tiles[i][j].TravelDiff = 5
// 					case "TS":
// 						m.Tiles[i][j].TravelDiff = 10
// 					case "M":
// 						m.Tiles[i][j].TravelDiff = 15
// 					}
// 				}
// 				i++
// 			}
// 		}
// 	}
// }

func (m *Map) PrintMap() {
	fmt.Println("Map size:", m.Size[0], "x", m.Size[1])
	fmt.Println("Step allowance:", m.StepAllowance)
	fmt.Println("Resources:", m.Resources)
	fmt.Println("Resource quota:", m.ResourceQuota)
	fmt.Println("Quota multiplier:", m.QuotaMult)
	fmt.Println("Tiles:")
	for _, row := range m.Tiles {
		for _, tile := range row {
			fmt.Print(tile.Type, " ")
		}
		fmt.Println()
	}
}

func (m *Map) FindPath(p *Party) [][]int {
	//var path [][]int
	// Implement pathfinding algorithm here
	path := [][]int{{0, 0}, {0, 1}, {0, 2}, {0, 3}, {1, 3}, {2, 3}, {2, 4}}
	return path
}

func (m *Map) CalculateScore(path [][]int, p *Party) int {
	score := 0
	stepCount := 0
	resources := make(map[string]int)
	for i := 0; i < len(path)-1; i++ {
		r1, c1 := path[i][0], path[i][1]
		r2, c2 := path[i+1][0], path[i+1][1]
		tile := m.Tiles[r2][c2]
		if tile.Resource != "" && !tile.ResourceUsed {
			resources[tile.Resource] += tile.ResourceAmt * p.ResourceBonus
			tile.ResourceUsed = true
		}
		stepCount += int(math.Abs(float64(r1-r2)) + math.Abs(float64(c1-c2)))
		score += tile.TravelDiff
	}
	if p.Scout {
		score = int(float64(score) / 2)
	}
	if stepCount > m.StepAllowance*p.StepBonus {
		score -= (stepCount - m.StepAllowance*p.StepBonus)
	}
	for resource, quota := range m.ResourceQuota {
		if resources[resource] >= quota {
			score += resources[resource] * m.QuotaMult
		}
	}
	return score
}

func atoi(str string) int {
	num, err := strconv.Atoi(str)
	if err != nil {
		fmt.Println("Error converting string to int:", err)
		return 0
	}
	return num
}

func main() {
	var party Party
	var map1 Map
	map1.LoadMap("./../../maps/map.txt")
	map1.PrintMap()
	party.SetParty([]string{"Scout", "Gatherer"})
	path := map1.FindPath(&party)
	fmt.Println("Path:", path)
	fmt.Println("Score:", map1.CalculateScore(path, &party))

	// Output the results in json format
	result := map[string]interface{}{
		"Party": party,
		"Path":  path,
		"Score": map1.CalculateScore(path, &party),
	}
	resultJson, _ := json.Marshal(result)
	fmt.Println(string(resultJson))
}
