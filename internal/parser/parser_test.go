package parser

import (
	"archive/zip"
	"awesomeProject/internal/models"
	"os"
	"path/filepath"
	"testing"
)

func TestParseFiles(t *testing.T) {
	dir := t.TempDir()
	csvPath := filepath.Join(dir, "ibdiagnet2.db_csv")
	sharpPath := filepath.Join(dir, "ibdiagnet2.sharp_an_info")

	writeTestFile(t, csvPath, `START_NODES
NodeDesc,NumPorts,NodeType,ClassVersion,BaseVersion,SystemImageGUID,NodeGUID,PortGUID
"HOST_1",1,1,1,1,0xhost1,0xhost1,0xhost1
"SWITCH_1",65,2,1,1,0xswitch1,0xswitch1,0xswitch1
END_NODES

START_PORTS
NodeGuid,PortGuid,PortNum,PortState
0xhost1,0xhost1,1,4
0xswitch1,0xswitch1,0,1
END_PORTS

START_SWITCHES
NodeGUID,LinearFDBCap,RandomFDBCap
0xswitch1,49152,0
END_SWITCHES

START_SYSTEM_GENERAL_INFORMATION
NodeGuid,SerialNumber,ProductName
0xswitch1,SOS123,"Gorilla"
END_SYSTEM_GENERAL_INFORMATION
`)

	writeTestFile(t, sharpPath, `-------------------------------------------------------------------------------------------
SW_GUID=switch1
-------------------------------------------------------------------------------------------
endianness = 0
enable_endianness_per_job = 1
`)

	parsed, err := ParseFiles(csvPath, sharpPath)
	if err != nil {
		t.Fatalf("ParseFiles returned error: %v", err)
	}

	if len(parsed.Nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(parsed.Nodes))
	}
	if parsed.Nodes[0].SourceID != "host1" || parsed.Nodes[0].Type != "host" {
		t.Fatalf("unexpected first node: %+v", parsed.Nodes[0])
	}
	if parsed.Nodes[1].SourceID != "switch1" || parsed.Nodes[1].Type != "switch" {
		t.Fatalf("unexpected second node: %+v", parsed.Nodes[1])
	}

	if len(parsed.Ports) != 2 {
		t.Fatalf("expected 2 ports, got %d", len(parsed.Ports))
	}
	if parsed.Ports[0].Name != "port-1" || parsed.Ports[0].Status != "active" {
		t.Fatalf("unexpected first port: %+v", parsed.Ports[0])
	}

	assertNodeInfo(t, parsed.NodeInfos, "switch1", "SerialNumber", "SOS123")
	assertNodeInfo(t, parsed.NodeInfos, "switch1", "ProductName", "Gorilla")
	assertNodeInfo(t, parsed.NodeInfos, "switch1", "endianness", "0")
}

func TestParseFilesRejectsUnknownPortNode(t *testing.T) {
	dir := t.TempDir()
	csvPath := filepath.Join(dir, "ibdiagnet2.db_csv")

	writeTestFile(t, csvPath, `START_NODES
NodeDesc,NodeType,NodeGUID
"HOST_1",1,0xhost1
END_NODES

START_PORTS
NodeGuid,PortNum,PortState
0xmissing,1,4
END_PORTS
`)

	_, err := ParseFiles(csvPath, "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestParseArchive(t *testing.T) {
	dir := t.TempDir()
	archivePath := filepath.Join(dir, "log.zip")

	createTestArchive(t, archivePath, map[string]string{
		"logs/ibdiagnet2.db_csv": `START_NODES
NodeDesc,NodeType,NodeGUID
"HOST_1",1,0xhost1
END_NODES

START_PORTS
NodeGuid,PortNum,PortState
0xhost1,1,4
END_PORTS
`,
		"logs/ibdiagnet2.sharp_an_info": `SW_GUID=host1
endianness = 0
`,
	})

	parsed, err := ParseArchive(archivePath)
	if err != nil {
		t.Fatalf("ParseArchive returned error: %v", err)
	}

	if len(parsed.Nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(parsed.Nodes))
	}
	if len(parsed.Ports) != 1 {
		t.Fatalf("expected 1 port, got %d", len(parsed.Ports))
	}
	assertNodeInfo(t, parsed.NodeInfos, "host1", "endianness", "0")
}

func TestParseArchiveRejectsUnsafePath(t *testing.T) {
	dir := t.TempDir()
	archivePath := filepath.Join(dir, "log.zip")

	createTestArchive(t, archivePath, map[string]string{
		"../ibdiagnet2.db_csv": "unsafe",
	})

	_, err := ParseArchive(archivePath)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func writeTestFile(t *testing.T, path string, content string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write test file: %v", err)
	}
}

func createTestArchive(t *testing.T, path string, files map[string]string) {
	t.Helper()

	output, err := os.Create(path)
	if err != nil {
		t.Fatalf("create test archive: %v", err)
	}
	defer output.Close()

	writer := zip.NewWriter(output)
	defer writer.Close()

	for name, content := range files {
		fileWriter, err := writer.Create(name)
		if err != nil {
			t.Fatalf("create archive file: %v", err)
		}
		if _, err := fileWriter.Write([]byte(content)); err != nil {
			t.Fatalf("write archive file: %v", err)
		}
	}
}

func assertNodeInfo(t *testing.T, infos []models.NodeInfo, sourceID string, key string, value string) {
	t.Helper()

	for _, info := range infos {
		if info.SourceID == sourceID && info.Key == key && info.Value == value {
			return
		}
	}

	t.Fatalf("node info %s %s=%s was not found in %+v", sourceID, key, value, infos)
}
