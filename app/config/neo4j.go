package config

import (
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// InitNeo4j initializes the Neo4j driver and returns it.
func InitNeo4j() (neo4j.DriverWithContext, error) {
	return neo4j.NewDriverWithContext("neo4j://neo4j:7687", neo4j.BasicAuth("neo4j", "password", ""))
}
