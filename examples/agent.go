package examples

// Agent is a sample DTO for DB records.
type Agent struct {
	AgentCode   string `db:"agent_code"`
	WorkingArea string `db:"working_area"`
	AgentName   string `db:"agent_name"`
	Commission  string `db:"commission"`
	PhoneNo     string `db:"phone_no"`
	Country     string `db:"country"`
}
