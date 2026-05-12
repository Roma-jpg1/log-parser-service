package parser

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"

	"awesomeProject/internal/models"
)

const (
	sectionNodes       = "NODES"
	sectionPorts       = "PORTS"
	sectionSwitches    = "SWITCHES"
	sectionSystemInfo  = "SYSTEM_GENERAL_INFORMATION"
	sectionStartPrefix = "START_"
	sectionEndPrefix   = "END_"
)

// ParseFiles reads ibdiagnet2 db_csv and sharp_an_info files and returns
// normalized nodes, ports and node key/value details.
func ParseFiles(csvPath string, sharpPath string) (*models.ParsedLog, error) {
	parsed, err := parseDBCSV(csvPath)
	if err != nil {
		return nil, err
	}

	if sharpPath != "" {
		infos, err := parseSharpAnInfo(sharpPath)
		if err != nil {
			return nil, err
		}
		parsed.NodeInfos = append(parsed.NodeInfos, infos...)
	}

	return parsed, nil
}

func parseDBCSV(path string) (*models.ParsedLog, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open db_csv: %w", err)
	}
	defer file.Close()

	sections, err := readSections(file)
	if err != nil {
		return nil, err
	}

	nodes, err := parseNodesSection(sections[sectionNodes])
	if err != nil {
		return nil, err
	}

	nodeIDs := make(map[string]struct{}, len(nodes))
	for _, node := range nodes {
		nodeIDs[node.SourceID] = struct{}{}
	}

	ports, err := parsePortsSection(sections[sectionPorts], nodeIDs)
	if err != nil {
		return nil, err
	}

	var infos []models.NodeInfo

	switchInfos, err := parseInfoSection(sections[sectionSwitches], "NodeGUID")
	if err != nil {
		return nil, err
	}
	infos = append(infos, switchInfos...)

	systemInfos, err := parseInfoSection(sections[sectionSystemInfo], "NodeGuid")
	if err != nil {
		return nil, err
	}
	infos = append(infos, systemInfos...)

	return &models.ParsedLog{
		Nodes:     nodes,
		Ports:     ports,
		NodeInfos: infos,
	}, nil
}

func readSections(reader io.Reader) (map[string][][]string, error) {
	scanner := bufio.NewScanner(reader)
	sections := make(map[string][][]string)
	var current string
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, sectionStartPrefix) {
			if current != "" {
				return nil, fmt.Errorf("line %d: nested section %q", lineNumber, line)
			}
			current = strings.TrimPrefix(line, sectionStartPrefix)
			if current == "" {
				return nil, fmt.Errorf("line %d: empty section name", lineNumber)
			}
			continue
		}

		if strings.HasPrefix(line, sectionEndPrefix) {
			name := strings.TrimPrefix(line, sectionEndPrefix)
			if current == "" {
				return nil, fmt.Errorf("line %d: end section %q without start", lineNumber, name)
			}
			if name != current {
				return nil, fmt.Errorf("line %d: end section %q does not match start section %q", lineNumber, name, current)
			}
			current = ""
			continue
		}

		if current == "" {
			return nil, fmt.Errorf("line %d: data outside section", lineNumber)
		}

		record, err := parseCSVLine(line)
		if err != nil {
			return nil, fmt.Errorf("line %d: parse csv: %w", lineNumber, err)
		}
		sections[current] = append(sections[current], record)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read db_csv: %w", err)
	}
	if current != "" {
		return nil, fmt.Errorf("section %q is not closed", current)
	}

	return sections, nil
}

func parseCSVLine(line string) ([]string, error) {
	reader := csv.NewReader(strings.NewReader(line))
	reader.TrimLeadingSpace = true

	record, err := reader.Read()
	if err != nil {
		return nil, err
	}
	if _, err := reader.Read(); err != io.EOF {
		return nil, fmt.Errorf("unexpected extra csv record")
	}

	for i := range record {
		record[i] = strings.TrimSpace(record[i])
	}

	return record, nil
}

func parseNodesSection(rows [][]string) ([]models.Node, error) {
	if len(rows) < 2 {
		return nil, fmt.Errorf("%s section is empty", sectionNodes)
	}

	header := indexHeader(rows[0])
	for _, column := range []string{"NodeDesc", "NodeType", "NodeGUID"} {
		if _, ok := header[column]; !ok {
			return nil, fmt.Errorf("%s section does not have %s column", sectionNodes, column)
		}
	}

	nodes := make([]models.Node, 0, len(rows)-1)
	seen := make(map[string]struct{}, len(rows)-1)

	for i, row := range rows[1:] {
		if len(row) != len(rows[0]) {
			return nil, fmt.Errorf("%s row %d: expected %d columns, got %d", sectionNodes, i+2, len(rows[0]), len(row))
		}

		sourceID := normalizeGUID(row[header["NodeGUID"]])
		if sourceID == "" {
			return nil, fmt.Errorf("%s row %d: empty NodeGUID", sectionNodes, i+2)
		}
		if _, exists := seen[sourceID]; exists {
			return nil, fmt.Errorf("%s row %d: duplicate NodeGUID %q", sectionNodes, i+2, sourceID)
		}
		seen[sourceID] = struct{}{}

		nodes = append(nodes, models.Node{
			SourceID: sourceID,
			Name:     row[header["NodeDesc"]],
			Type:     nodeType(row[header["NodeType"]]),
		})
	}

	return nodes, nil
}

func parsePortsSection(rows [][]string, nodeIDs map[string]struct{}) ([]models.Port, error) {
	if len(rows) < 2 {
		return nil, fmt.Errorf("%s section is empty", sectionPorts)
	}

	header := indexHeader(rows[0])
	for _, column := range []string{"NodeGuid", "PortNum", "PortState"} {
		if _, ok := header[column]; !ok {
			return nil, fmt.Errorf("%s section does not have %s column", sectionPorts, column)
		}
	}

	ports := make([]models.Port, 0, len(rows)-1)

	for i, row := range rows[1:] {
		if len(row) != len(rows[0]) {
			return nil, fmt.Errorf("%s row %d: expected %d columns, got %d", sectionPorts, i+2, len(rows[0]), len(row))
		}

		sourceID := normalizeGUID(row[header["NodeGuid"]])
		if _, exists := nodeIDs[sourceID]; !exists {
			return nil, fmt.Errorf("%s row %d: node %q was not found in %s section", sectionPorts, i+2, sourceID, sectionNodes)
		}

		portNum := row[header["PortNum"]]
		if portNum == "" {
			return nil, fmt.Errorf("%s row %d: empty PortNum", sectionPorts, i+2)
		}

		ports = append(ports, models.Port{
			SourceID: sourceID,
			Name:     "port-" + portNum,
			Status:   portState(row[header["PortState"]]),
		})
	}

	return ports, nil
}

func parseInfoSection(rows [][]string, sourceColumn string) ([]models.NodeInfo, error) {
	if len(rows) == 0 {
		return nil, nil
	}
	if len(rows) < 2 {
		return nil, fmt.Errorf("info section with %s column is empty", sourceColumn)
	}

	header := indexHeader(rows[0])
	sourceIndex, ok := header[sourceColumn]
	if !ok {
		return nil, fmt.Errorf("info section does not have %s column", sourceColumn)
	}

	infos := make([]models.NodeInfo, 0, (len(rows)-1)*(len(rows[0])-1))

	for i, row := range rows[1:] {
		if len(row) != len(rows[0]) {
			return nil, fmt.Errorf("info row %d: expected %d columns, got %d", i+2, len(rows[0]), len(row))
		}

		sourceID := normalizeGUID(row[sourceIndex])
		if sourceID == "" {
			return nil, fmt.Errorf("info row %d: empty %s", i+2, sourceColumn)
		}

		for columnIndex, key := range rows[0] {
			if columnIndex == sourceIndex {
				continue
			}
			infos = append(infos, models.NodeInfo{
				SourceID: sourceID,
				Key:      key,
				Value:    row[columnIndex],
			})
		}
	}

	return infos, nil
}

func parseSharpAnInfo(path string) ([]models.NodeInfo, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open sharp_an_info: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var sourceID string
	var infos []models.NodeInfo
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "-") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return nil, fmt.Errorf("sharp_an_info line %d: expected key=value", lineNumber)
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" {
			return nil, fmt.Errorf("sharp_an_info line %d: empty key", lineNumber)
		}

		if key == "SW_GUID" {
			sourceID = normalizeGUID(value)
			if sourceID == "" {
				return nil, fmt.Errorf("sharp_an_info line %d: empty SW_GUID", lineNumber)
			}
			continue
		}

		if sourceID == "" {
			return nil, fmt.Errorf("sharp_an_info line %d: parameter before SW_GUID", lineNumber)
		}

		infos = append(infos, models.NodeInfo{
			SourceID: sourceID,
			Key:      key,
			Value:    value,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read sharp_an_info: %w", err)
	}

	return infos, nil
}

func indexHeader(header []string) map[string]int {
	index := make(map[string]int, len(header))
	for i, column := range header {
		index[column] = i
	}
	return index
}

func normalizeGUID(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "0x")
	return strings.ToLower(value)
}

func nodeType(value string) string {
	switch value {
	case "1":
		return "host"
	case "2":
		return "switch"
	default:
		return "unknown"
	}
}

func portState(value string) string {
	switch value {
	case "1":
		return "down"
	case "2":
		return "init"
	case "3":
		return "armed"
	case "4":
		return "active"
	default:
		return value
	}
}
