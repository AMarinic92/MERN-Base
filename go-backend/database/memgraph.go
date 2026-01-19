package database

import (
	"context"
	"fmt"
	"go-backend/models"
	"strings"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)


func processTypes(typeLine string) []string {
    // Splits "Legendary Creature — Elf Shaman" into ["Legendary", "Creature", "Elf", "Shaman"]
    clean := strings.ReplaceAll(typeLine, " — ", " ")
    return strings.Fields(clean)
}

func derefString(s *string) string {
    if s == nil { return "" }
    return *s
}

func derefFloat(f *float64) float64 {
    if f == nil { return 0.0 }
    return *f
}

func GetCardSuggestions(oracleID string) ([]models.Card, error) {
	ctx := context.Background()
	session := GraphDriver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	// 1. Execute the Graph Search
	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		cypher := `
			MATCH (source:Card {id: $id})
			MATCH (source)-[:PRODUCES|HAS_KEYWORD|IS_TYPE]->(attr)
			MATCH (rec:Card)-[:PRODUCES|HAS_KEYWORD|IS_TYPE]->(attr)
			WHERE rec.id <> source.id
			WITH rec, count(DISTINCT attr) AS sharedCount
			ORDER BY sharedCount DESC
			LIMIT 10
			RETURN rec.id AS id
		`
		res, err := tx.Run(ctx, cypher, map[string]interface{}{"id": oracleID})
		if err != nil {
			return nil, err
		}

		var ids []string
		for res.Next(ctx) {
			if id, ok := res.Record().Get("id"); ok {
				ids = append(ids, id.(string))
			}
		}
		return ids, nil
	})

	if err != nil {
		return nil, fmt.Errorf("graph search failed: %w", err)
	}

	suggestedIDs := result.([]string)
	if len(suggestedIDs) == 0 {
		return []models.Card{}, nil
	}

	// 2. Hydrate from Postgres (Bulk fetch for speed)
	var cards []models.Card
	err = DB.Where("id IN ?", suggestedIDs).
		Order(fmt.Sprintf("idx(array['%s'], id)", strings.Join(suggestedIDs, "','"))). // Maintain Graph Rank
		Find(&cards).Error

	return cards, err
}

func syncToMemgraph(cards []*models.Card) error {
    ctx := context.Background()
    session := GraphDriver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
    defer session.Close(ctx)

    // Prepare a "lean" map for Memgraph ingestion
    var batchData []map[string]interface{}
    for _, c := range cards {
        batchData = append(batchData, map[string]interface{}{
            "id":        c.ID,
            "name":      c.Name,
            "cmc":       derefFloat(c.CMC),
            "types":     processTypes(c.TypeLine),
            "keywords":  []string(c.Keywords),
            "mechanics": extractMechanics(derefString(c.OracleText)),
        })
    }

    // The "Power Query": Updates nodes and relationships in one go
    _, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
        query := `
        UNWIND $batch AS data
        MERGE (c:Card {id: data.id})
        SET c.name = data.name, c.cmc = data.cmc
        
        // Connect Types
        WITH c, data
        UNWIND data.types AS tName
        MERGE (t:Type {name: tName})
        MERGE (c)-[:IS_TYPE]->(t)
        
        // Connect Keywords
        WITH c, data
        UNWIND data.keywords AS kName
        MERGE (k:Keyword {name: kName})
        MERGE (c)-[:HAS_KEYWORD]->(k)
        
        // Connect Mechanics
        WITH c, data
        UNWIND data.mechanics AS mName
        MERGE (m:Mechanic {name: mName})
        MERGE (c)-[:PRODUCES]->(m)
        `
        return tx.Run(ctx, query, map[string]interface{}{"batch": batchData})
    })

    return err
}

func extractMechanics(oracleText string) []string {
	mechanics := []string{}
	text := strings.ToLower(oracleText)

	// Simple pattern matching for core synergies
	mapping := map[string][]string{
		"Draw":      {"draw a card", "draws a card"},
		"Ramp":      {"search your library for a land", "add {g}", "add {u}", "add {r}", "add {b}", "add {w}"},
		"Token":     {"create", "token"},
		"Lifegain":  {"gain", "life"},
		"Graveyard": {"return", "graveyard", "exile from your graveyard"},
	}

	for mech, keywords := range mapping {
		for _, kw := range keywords {
			if strings.Contains(text, kw) {
				mechanics = append(mechanics, mech)
				break 
			}
		}
	}
	return mechanics
}

func executeSchema(ctx context.Context) {
	session := GraphDriver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	queries := []string{
		"CREATE CONSTRAINT ON (c:Card) ASSERT c.id IS UNIQUE;",
		"CREATE INDEX ON :Type(name);",
		"CREATE INDEX ON :Keyword(name);",
		"CREATE INDEX ON :Mechanic(name);",
	}

	for _, q := range queries {
		session.Run(ctx, q, nil)
	}
}