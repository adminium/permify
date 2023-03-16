package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	
	"github.com/Masterminds/squirrel"
	
	"go.opentelemetry.io/otel/codes"
	
	"github.com/adminium/permify/internal/repositories"
	"github.com/adminium/permify/internal/repositories/postgres/snapshot"
	"github.com/adminium/permify/internal/repositories/postgres/types"
	"github.com/adminium/permify/internal/repositories/postgres/utils"
	"github.com/adminium/permify/pkg/database"
	db "github.com/adminium/permify/pkg/database/postgres"
	"github.com/adminium/permify/pkg/logger"
	base "github.com/adminium/permify/pkg/pb/base/v1"
	"github.com/adminium/permify/pkg/token"
)

type RelationshipReader struct {
	database *db.Postgres
	// options
	txOptions sql.TxOptions
	// logger
	logger logger.Interface
}

// NewRelationshipReader - Creates a new RelationshipReader
func NewRelationshipReader(database *db.Postgres, logger logger.Interface) *RelationshipReader {
	return &RelationshipReader{
		database:  database,
		txOptions: sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true},
		logger:    logger,
	}
}

// QueryRelationships - Query relationships for a given filter
func (r *RelationshipReader) QueryRelationships(ctx context.Context, tenantID string, filter *base.TupleFilter, snap string) (it *database.TupleIterator, err error) {
	ctx, span := tracer.Start(ctx, "relationship-reader.query-relationships")
	defer span.End()
	
	var st token.SnapToken
	st, err = snapshot.EncodedToken{Value: snap}.Decode()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	
	var tx *sql.Tx
	tx, err = r.database.DB.BeginTx(ctx, &r.txOptions)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	
	defer utils.Rollback(tx, r.logger)
	
	var args []interface{}
	
	builder := r.database.Builder.Select("entity_type, entity_id, relation, subject_type, subject_id, subject_relation").From(RelationTuplesTable).Where(squirrel.Eq{"tenant_id": tenantID})
	builder = utils.FilterQueryForSelectBuilder(builder, filter)
	
	builder = utils.SnapshotQuery(builder, st.(snapshot.Token).Value.Uint)
	
	var query string
	query, args, err = builder.ToSql()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, errors.New(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String())
	}
	
	var rows *sql.Rows
	rows, err = tx.QueryContext(ctx, query, args...)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}
	defer rows.Close()
	
	collection := database.NewTupleCollection()
	for rows.Next() {
		rt := repositories.RelationTuple{}
		err = rows.Scan(&rt.EntityType, &rt.EntityID, &rt.Relation, &rt.SubjectType, &rt.SubjectID, &rt.SubjectRelation)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, err
		}
		collection.Add(rt.ToTuple())
	}
	if err = rows.Err(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	
	err = tx.Commit()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	
	return collection.CreateTupleIterator(), nil
}

// ReadRelationships - Read relationships for a given filter and pagination
func (r *RelationshipReader) ReadRelationships(ctx context.Context, tenantID string, filter *base.TupleFilter, snap string, pagination database.Pagination) (collection *database.TupleCollection, ct database.EncodedContinuousToken, err error) {
	ctx, span := tracer.Start(ctx, "relationship-reader.read-relationships")
	defer span.End()
	
	var st token.SnapToken
	st, err = snapshot.EncodedToken{Value: snap}.Decode()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, nil, err
	}
	
	var tx *sql.Tx
	tx, err = r.database.DB.BeginTx(ctx, &r.txOptions)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, nil, err
	}
	
	defer utils.Rollback(tx, r.logger)
	
	builder := r.database.Builder.Select("id, entity_type, entity_id, relation, subject_type, subject_id, subject_relation").From(RelationTuplesTable).Where(squirrel.Eq{"tenant_id": tenantID})
	builder = utils.FilterQueryForSelectBuilder(builder, filter)
	
	builder = utils.SnapshotQuery(builder, st.(snapshot.Token).Value.Uint)
	
	if pagination.Token() != "" {
		var t database.ContinuousToken
		t, err = utils.EncodedContinuousToken{Value: pagination.Token()}.Decode()
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, nil, err
		}
		var v uint64
		v, err = strconv.ParseUint(t.(utils.ContinuousToken).Value, 10, 64)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, nil, errors.New(base.ErrorCode_ERROR_CODE_INVALID_CONTINUOUS_TOKEN.String())
		}
		builder = builder.Where(squirrel.GtOrEq{"id": v})
	}
	
	builder = builder.OrderBy("id").Limit(uint64(pagination.PageSize() + 1))
	
	var query string
	var args []interface{}
	
	query, args, err = builder.ToSql()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, utils.NewNoopContinuousToken().Encode(), errors.New(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String())
	}
	
	var rows *sql.Rows
	rows, err = tx.QueryContext(ctx, query, args...)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, utils.NewNoopContinuousToken().Encode(), errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}
	defer rows.Close()
	
	var lastID uint64
	
	tuples := make([]*base.Tuple, 0, pagination.PageSize()+1)
	for rows.Next() {
		rt := repositories.RelationTuple{}
		err = rows.Scan(&rt.ID, &rt.EntityType, &rt.EntityID, &rt.Relation, &rt.SubjectType, &rt.SubjectID, &rt.SubjectRelation)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, nil, err
		}
		lastID = rt.ID
		tuples = append(tuples, rt.ToTuple())
	}
	if err = rows.Err(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, nil, err
	}
	
	err = tx.Commit()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, nil, err
	}
	
	if len(tuples) > int(pagination.PageSize()) {
		return database.NewTupleCollection(tuples[:pagination.PageSize()]...), utils.NewContinuousToken(strconv.FormatUint(lastID, 10)).Encode(), nil
	}
	
	return database.NewTupleCollection(tuples...), utils.NewNoopContinuousToken().Encode(), nil
}

// GetUniqueEntityIDsByEntityType - Gets all unique entity ids for a given entity type
func (r *RelationshipReader) GetUniqueEntityIDsByEntityType(ctx context.Context, tenantID string, typ, snap string) (ids []string, err error) {
	ctx, span := tracer.Start(ctx, "relationship-reader.get-unique-entity-ids-by-entity-type")
	defer span.End()
	
	var st token.SnapToken
	st, err = snapshot.EncodedToken{Value: snap}.Decode()
	if err != nil {
		return nil, err
	}
	
	var tx *sql.Tx
	tx, err = r.database.DB.BeginTx(ctx, &r.txOptions)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	
	defer utils.Rollback(tx, r.logger)
	
	var args []interface{}
	
	builder := r.database.Builder.Select("entity_id").Distinct().From(RelationTuplesTable).Where(squirrel.Eq{"entity_type": typ, "tenant_id": tenantID})
	builder = utils.SnapshotQuery(builder, st.(snapshot.Token).Value.Uint)
	
	var query string
	query, args, err = builder.ToSql()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, errors.New(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String())
	}
	
	var rows *sql.Rows
	rows, err = tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}
	defer rows.Close()
	
	var result []string
	for rows.Next() {
		var id string
		err = rows.Scan(&id)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, err
		}
		result = append(result, id)
	}
	if err = rows.Err(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	
	err = tx.Commit()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	
	return result, nil
}

// HeadSnapshot - Gets the latest token
func (r *RelationshipReader) HeadSnapshot(ctx context.Context, tenantID string) (token.SnapToken, error) {
	ctx, span := tracer.Start(ctx, "relationship-reader.head-snapshot")
	defer span.End()
	
	var xid types.XID8
	builder := r.database.Builder.Select("id").From(TransactionsTable).Where(squirrel.Eq{"tenant_id": tenantID}).OrderBy("id DESC").Limit(1)
	query, args, err := builder.ToSql()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, errors.New(base.ErrorCode_ERROR_CODE_SQL_BUILDER.String())
	}
	
	row := r.database.DB.QueryRowContext(ctx, query, args...)
	err = row.Scan(&xid)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		if errors.Is(err, sql.ErrNoRows) {
			return snapshot.Token{Value: types.XID8{Uint: 0}}, nil
		}
		return nil, err
	}
	
	return snapshot.Token{Value: xid}, nil
}
