package development

import (
	"fmt"
	
	"github.com/adminium/permify/internal/commands"
	"github.com/adminium/permify/internal/config"
	"github.com/adminium/permify/internal/factories"
	"github.com/adminium/permify/internal/keys"
	"github.com/adminium/permify/internal/services"
	"github.com/adminium/permify/pkg/database"
	"github.com/adminium/permify/pkg/logger"
	"github.com/adminium/permify/pkg/telemetry"
)

// Container - Structure for container instance
type Container struct {
	P services.IPermissionService
	R services.IRelationshipService
	S services.ISchemaService
}

// NewContainer - Creates new container instance
func NewContainer() *Container {
	var err error
	
	var db database.Database
	db, err = factories.DatabaseFactory(config.Database{Engine: database.MEMORY.String()})
	if err != nil {
		fmt.Println(err)
	}
	
	l := logger.New("debug")
	
	// Repositories
	relationshipReader := factories.RelationshipReaderFactory(db, l)
	relationshipWriter := factories.RelationshipWriterFactory(db, l)
	
	schemaReader := factories.SchemaReaderFactory(db, l)
	schemaWriter := factories.SchemaWriterFactory(db, l)
	
	// commands
	checkCommand, _ := commands.NewCheckCommand(keys.NewNoopCheckCommandKeys(), schemaReader, relationshipReader, telemetry.NewNoopMeter())
	expandCommand := commands.NewExpandCommand(schemaReader, relationshipReader)
	lookupSchemaCommand := commands.NewLookupSchemaCommand(schemaReader)
	lookupEntityCommand := commands.NewLookupEntityCommand(checkCommand, schemaReader, relationshipReader)
	
	return &Container{
		P: services.NewPermissionService(checkCommand, expandCommand, lookupSchemaCommand, lookupEntityCommand),
		R: services.NewRelationshipService(relationshipReader, relationshipWriter, schemaReader),
		S: services.NewSchemaService(schemaWriter, schemaReader),
	}
}
