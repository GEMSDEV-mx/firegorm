package firegorm

import (
	"context"
	"fmt"
	"sync"
)

// HookType covers all supported events.
type HookType string

const (
    PreCreate  HookType = "pre_create"
    PostCreate HookType = "post_create"
    PreUpdate  HookType = "pre_update"
    PostUpdate HookType = "post_update"
    PreDelete  HookType = "pre_delete"
    PostDelete HookType = "post_delete"
)

// HookFunc is any function that inspects or mutates 'data' before/after an op.
type HookFunc func(ctx context.Context, data interface{}) error

// HookRegistry manages registration and execution.
type HookRegistry struct {
    mu            sync.RWMutex
    hooks         map[string]map[HookType][]HookFunc      // collection -> hookType -> []func
    enabledAll    bool                                     // global master switch
    enabledTypes  map[HookType]bool                        // per-HookType on/off
    enabledScopes map[string]map[HookType]bool             // per-collection+hookType on/off
}

// DefaultRegistry is the one used by BaseModel.
var DefaultRegistry = NewHookRegistry()

// NewHookRegistry constructs an empty registry.
func NewHookRegistry() *HookRegistry {
    return &HookRegistry{
        hooks:         make(map[string]map[HookType][]HookFunc),
        enabledAll:    true,
        enabledTypes:  make(map[HookType]bool),
        enabledScopes: make(map[string]map[HookType]bool),
    }
}

// RegisterHook registers `fn` under a collection and HookType.
func (r *HookRegistry) RegisterHook(collection string, ht HookType, fn HookFunc) {
    r.mu.Lock()
    defer r.mu.Unlock()
    if r.hooks[collection] == nil {
        r.hooks[collection] = make(map[HookType][]HookFunc)
    }
    r.hooks[collection][ht] = append(r.hooks[collection][ht], fn)
    // enable by default
    if r.enabledScopes[collection] == nil {
        r.enabledScopes[collection] = make(map[HookType]bool)
    }
    r.enabledScopes[collection][ht] = true
    // ensure hookType is enabled in types map
    if _, exists := r.enabledTypes[ht]; !exists {
        r.enabledTypes[ht] = true
    }
}

// EnableAll turns every hook on or off.
func (r *HookRegistry) EnableAll(on bool) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.enabledAll = on
}

// EnableType turns all hooks of that type on or off.
func (r *HookRegistry) EnableType(ht HookType, on bool) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.enabledTypes[ht] = on
}

// EnableScope turns a specific collection+hookType on or off.
func (r *HookRegistry) EnableScope(collection string, ht HookType, on bool) {
    r.mu.Lock()
    defer r.mu.Unlock()
    if r.enabledScopes[collection] == nil {
        r.enabledScopes[collection] = make(map[HookType]bool)
    }
    r.enabledScopes[collection][ht] = on
}

// runHooks executes, in order, all enabled hooks for this point.
func (r *HookRegistry) RunHooks(ctx context.Context, collection string, ht HookType, data interface{}) error {
    r.mu.RLock()
    defer r.mu.RUnlock()

    // master switch
    if !r.enabledAll {
        return nil
    }
    // per-HookType switch
    if on, ok := r.enabledTypes[ht]; !ok || !on {
        return nil
    }
    // per-collection+HookType switch
    if scope, ok := r.enabledScopes[collection]; !ok || !scope[ht] {
        return nil
    }

    // grab the slice
    fns := r.hooks[collection][ht]
    for _, fn := range fns {
        if err := fn(ctx, data); err != nil {
            return fmt.Errorf("hook %s on %s failed: %w", ht, collection, err)
        }
    }
    return nil
}
