package keys

import (
	"encoding/hex"
	"fmt"
	
	"github.com/cespare/xxhash"
	
	"github.com/adminium/permify/pkg/cache"
	base "github.com/adminium/permify/pkg/pb/base/v1"
	"github.com/adminium/permify/pkg/tuple"
)

type CommandKeys struct {
	cache cache.Cache
}

// NewCheckCommandKeys new instance of CheckCommandKeys
func NewCheckCommandKeys(cache cache.Cache) CommandKeyManager {
	return &CommandKeys{
		cache: cache,
	}
}

// SetCheckKey - Sets the value for the given key.
func (c *CommandKeys) SetCheckKey(key *base.PermissionCheckRequest, value *base.PermissionCheckResponse) bool {
	checkKey := fmt.Sprintf("check_%s_%s:%s:%s@%s", key.GetTenantId(), key.GetMetadata().GetSchemaVersion(), key.GetMetadata().GetSnapToken(), tuple.EntityAndRelationToString(&base.EntityAndRelation{
		Entity:   key.GetEntity(),
		Relation: key.GetPermission(),
	}), tuple.SubjectToString(key.GetSubject()))
	h := xxhash.New()
	size, err := h.Write([]byte(checkKey))
	if err != nil {
		return false
	}
	k := hex.EncodeToString(h.Sum(nil))
	return c.cache.Set(k, value, int64(size))
}

// GetCheckKey - Gets the value for the given key.
func (c *CommandKeys) GetCheckKey(key *base.PermissionCheckRequest) (*base.PermissionCheckResponse, bool) {
	checkKey := fmt.Sprintf("check_%s_%s:%s:%s@%s", key.GetTenantId(), key.GetMetadata().GetSchemaVersion(), key.GetMetadata().GetSnapToken(), tuple.EntityAndRelationToString(&base.EntityAndRelation{
		Entity:   key.GetEntity(),
		Relation: key.GetPermission(),
	}), tuple.SubjectToString(key.GetSubject()))
	h := xxhash.New()
	_, err := h.Write([]byte(checkKey))
	if err != nil {
		return nil, false
	}
	k := hex.EncodeToString(h.Sum(nil))
	resp, found := c.cache.Get(k)
	if found {
		return resp.(*base.PermissionCheckResponse), true
	}
	return nil, false
}

// NoopCommandKeys -
type NoopCommandKeys struct{}

// NewNoopCheckCommandKeys new noop instance of CheckCommandKeys
func NewNoopCheckCommandKeys() CommandKeyManager {
	return &NoopCommandKeys{}
}

// SetCheckKey sets the value for the given key.
func (c *NoopCommandKeys) SetCheckKey(*base.PermissionCheckRequest, *base.PermissionCheckResponse) bool {
	return true
}

// GetCheckKey gets the value for the given key.
func (c *NoopCommandKeys) GetCheckKey(*base.PermissionCheckRequest) (*base.PermissionCheckResponse, bool) {
	return nil, false
}
