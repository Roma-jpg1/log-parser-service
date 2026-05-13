package db

import (
	"database/sql"
	"fmt"

	"awesomeProject/internal/models"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(database *sql.DB) *Repository {
	return &Repository{db: database}
}

func (r *Repository) CreateLog(filename string) (int, error) {
	var id int

	err := r.db.QueryRow(
		`INSERT INTO logs (filename, status) VALUES ($1, $2) RETURNING id`,
		filename,
		"pending",
	).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *Repository) MarkLogSuccess(logID int, nodesCount int, portsCount int) error {
	_, err := r.db.Exec(
		`UPDATE logs SET status = $1, nodes_count = $2, ports_count = $3 WHERE id = $4`,
		"success",
		nodesCount,
		portsCount,
		logID,
	)

	return err
}

func (r *Repository) MarkLogFailed(logID int, errorText string) error {
	_, err := r.db.Exec(
		`UPDATE logs SET status = $1, error_text = $2 WHERE id = $3`,
		"failed",
		errorText,
		logID,
	)

	return err
}

func (r *Repository) SaveParsedLog(logID int, parsed *models.ParsedLog) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	nodeIDBySourceID := make(map[string]int)

	for _, node := range parsed.Nodes {
		var nodeID int

		err := tx.QueryRow(
			`INSERT INTO nodes (log_id, source_id, name, type)
			 VALUES ($1, $2, $3, $4)
			 RETURNING id`,
			logID,
			node.SourceID,
			node.Name,
			node.Type,
		).Scan(&nodeID)

		if err != nil {
			return err
		}

		nodeIDBySourceID[node.SourceID] = nodeID
	}

	for _, port := range parsed.Ports {
		nodeID, exists := nodeIDBySourceID[port.SourceID]
		if !exists {
			return fmt.Errorf("node for port not found: source_id=%s", port.SourceID)
		}

		_, err := tx.Exec(
			`INSERT INTO ports (log_id, node_id, source_id, name, status)
			 VALUES ($1, $2, $3, $4, $5)`,
			logID,
			nodeID,
			port.SourceID,
			port.Name,
			port.Status,
		)

		if err != nil {
			return err
		}
	}

	for _, info := range parsed.NodeInfos {
		nodeID, exists := nodeIDBySourceID[info.SourceID]
		if !exists {
			continue
		}

		_, err := tx.Exec(
			`INSERT INTO nodes_info (node_id, key, value)
			 VALUES ($1, $2, $3)`,
			nodeID,
			info.Key,
			info.Value,
		)

		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *Repository) GetLogByID(logID int) (*models.Log, error) {
	var item models.Log

	err := r.db.QueryRow(
		`SELECT id, filename, status, error_text, nodes_count, ports_count, created_at
		 FROM logs
		 WHERE id = $1`,
		logID,
	).Scan(
		&item.ID,
		&item.Filename,
		&item.Status,
		&item.ErrorText,
		&item.NodesCount,
		&item.PortsCount,
		&item.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &item, nil
}

func (r *Repository) GetNodesByLogID(logID int) ([]models.Node, error) {
	rows, err := r.db.Query(
		`SELECT id, log_id, source_id, name, type
		 FROM nodes
		 WHERE log_id = $1
		 ORDER BY id`,
		logID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []models.Node

	for rows.Next() {
		var node models.Node

		err := rows.Scan(
			&node.ID,
			&node.LogID,
			&node.SourceID,
			&node.Name,
			&node.Type,
		)
		if err != nil {
			return nil, err
		}

		nodes = append(nodes, node)
	}

	return nodes, rows.Err()
}

func (r *Repository) GetNodeByID(nodeID int) (*models.Node, error) {
	var node models.Node

	err := r.db.QueryRow(
		`SELECT id, log_id, source_id, name, type
		 FROM nodes
		 WHERE id = $1`,
		nodeID,
	).Scan(
		&node.ID,
		&node.LogID,
		&node.SourceID,
		&node.Name,
		&node.Type,
	)

	if err != nil {
		return nil, err
	}

	return &node, nil
}

func (r *Repository) GetPortsByNodeID(nodeID int) ([]models.Port, error) {
	rows, err := r.db.Query(
		`SELECT id, log_id, node_id, source_id, name, status
		 FROM ports
		 WHERE node_id = $1
		 ORDER BY id`,
		nodeID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ports []models.Port

	for rows.Next() {
		var port models.Port

		err := rows.Scan(
			&port.ID,
			&port.LogID,
			&port.NodeID,
			&port.SourceID,
			&port.Name,
			&port.Status,
		)
		if err != nil {
			return nil, err
		}

		ports = append(ports, port)
	}

	return ports, rows.Err()
}

func (r *Repository) GetNodeInfoByNodeID(nodeID int) ([]models.NodeInfo, error) {
	rows, err := r.db.Query(
		`SELECT id, node_id, key, value
		 FROM nodes_info
		 WHERE node_id = $1
		 ORDER BY key`,
		nodeID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var infos []models.NodeInfo

	for rows.Next() {
		var info models.NodeInfo

		err := rows.Scan(
			&info.ID,
			&info.NodeID,
			&info.Key,
			&info.Value,
		)
		if err != nil {
			return nil, err
		}

		infos = append(infos, info)
	}

	return infos, rows.Err()
}
