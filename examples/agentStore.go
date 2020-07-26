package examples

import (
	"github.com/jmoiron/sqlx"
)

// AgentStore An example entity store for agents
type AgentStore struct {
	DBX *sqlx.DB
}

// GetAgents returns agents from store
func (a *AgentStore) GetAgents() ([]Agent, error) {
	// this is some sample db code using sqlx
	var agents = []Agent{}

	tx := a.DBX.MustBegin()
	defer tx.Commit()

	err := tx.Select(&agents, "select * from agents")
	if err != nil {
		return nil, err
	}

	return agents, nil
}
