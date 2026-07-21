package system

import (
	"errors"
	"fmt"
	"sort"

	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
)

var errAuthorityHierarchyCycle = errors.New("角色层级存在循环")

// authorityHierarchy indexes the flat authority records once and keeps tree
// assembly independent from database access.
type authorityHierarchy struct {
	authorities map[uint]system.SysAuthority
	children    map[uint][]uint
}

func newAuthorityHierarchy(authorities []system.SysAuthority) (*authorityHierarchy, error) {
	hierarchy := &authorityHierarchy{
		authorities: make(map[uint]system.SysAuthority, len(authorities)),
		children:    make(map[uint][]uint),
	}
	for _, authority := range authorities {
		if _, exists := hierarchy.authorities[authority.AuthorityId]; exists {
			return nil, fmt.Errorf("角色 ID %d 重复", authority.AuthorityId)
		}
		authority.Children = nil
		hierarchy.authorities[authority.AuthorityId] = authority
		if authority.ParentId != nil {
			hierarchy.children[*authority.ParentId] = append(hierarchy.children[*authority.ParentId], authority.AuthorityId)
		}
	}
	for parentID := range hierarchy.children {
		sort.Slice(hierarchy.children[parentID], func(i, j int) bool {
			return hierarchy.children[parentID][i] < hierarchy.children[parentID][j]
		})
	}
	return hierarchy, nil
}

func (h *authorityHierarchy) authority(authorityID uint) (system.SysAuthority, bool) {
	authority, exists := h.authorities[authorityID]
	return authority, exists
}

func (h *authorityHierarchy) childIDs(parentID uint) []uint {
	return h.children[parentID]
}

func (h *authorityHierarchy) forest(rootIDs []uint) ([]system.SysAuthority, error) {
	forest := make([]system.SysAuthority, 0, len(rootIDs))
	path := make(map[uint]struct{})
	for _, rootID := range rootIDs {
		root, err := h.tree(rootID, path)
		if err != nil {
			return nil, err
		}
		forest = append(forest, root)
	}
	return forest, nil
}

func (h *authorityHierarchy) tree(authorityID uint, path map[uint]struct{}) (system.SysAuthority, error) {
	if _, cycling := path[authorityID]; cycling {
		return system.SysAuthority{}, fmt.Errorf("%w: %d", errAuthorityHierarchyCycle, authorityID)
	}
	authority, exists := h.authorities[authorityID]
	if !exists {
		return system.SysAuthority{}, fmt.Errorf("角色 %d 不存在", authorityID)
	}
	path[authorityID] = struct{}{}
	defer delete(path, authorityID)

	children := h.childIDs(authorityID)
	authority.Children = make([]system.SysAuthority, 0, len(children))
	for _, childID := range children {
		child, err := h.tree(childID, path)
		if err != nil {
			return system.SysAuthority{}, err
		}
		authority.Children = append(authority.Children, child)
	}
	return authority, nil
}

func (h *authorityHierarchy) descendantIDs(authorityID uint) ([]uint, error) {
	if _, exists := h.authorities[authorityID]; !exists {
		return nil, fmt.Errorf("角色 %d 不存在", authorityID)
	}
	descendants := make([]uint, 0)
	path := make(map[uint]struct{})
	var visit func(uint) error
	visit = func(parentID uint) error {
		if _, cycling := path[parentID]; cycling {
			return fmt.Errorf("%w: %d", errAuthorityHierarchyCycle, parentID)
		}
		path[parentID] = struct{}{}
		defer delete(path, parentID)
		for _, childID := range h.childIDs(parentID) {
			descendants = append(descendants, childID)
			if err := visit(childID); err != nil {
				return err
			}
		}
		return nil
	}
	if err := visit(authorityID); err != nil {
		return nil, err
	}
	return descendants, nil
}
