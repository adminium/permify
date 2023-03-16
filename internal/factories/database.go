package factories

import (
	"fmt"
	
	"github.com/adminium/permify/internal/config"
	"github.com/adminium/permify/internal/repositories/memory/migrations"
	"github.com/adminium/permify/pkg/database"
	IMDatabase "github.com/adminium/permify/pkg/database/memory"
	PQDatabase "github.com/adminium/permify/pkg/database/postgres"
)

// DatabaseFactory - Create database according to given configuration
func DatabaseFactory(conf config.Database) (db database.Database, err error) {
	switch conf.Engine {
	case database.POSTGRES.String():
		db, err = PQDatabase.New(conf.URI,
			PQDatabase.MaxOpenConnections(conf.MaxOpenConnections),
			PQDatabase.MaxIdleConnections(conf.MaxIdleConnections),
			PQDatabase.MaxConnectionIdleTime(conf.MaxConnectionIdleTime),
			PQDatabase.MaxConnectionLifeTime(conf.MaxConnectionLifetime),
		)
		if err != nil {
			return nil, err
		}
		return
	case database.MEMORY.String():
		db, err = IMDatabase.New(migrations.Schema)
		if err != nil {
			return nil, err
		}
		return
	default:
		return nil, fmt.Errorf("%s connection is unsupported", conf.Engine)
	}
}
