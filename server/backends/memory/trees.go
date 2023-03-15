package memory

import (
	"context"
	"fmt"
	"strconv"

	"github.com/vektah/gqlparser/v2/gqlerror"

	"layout/server/graphql/model"
)

// IDs: We have a global ID for all nodes that have references to/from.
// Since we always ingest data and never remove, we can keep this global and
// increment it as needed.
// For fast retrieval, we also keep a map from ID from nodes that have it.
type nodeID int
type hasID interface {
	getID() nodeID
}
type indexType map[nodeID]hasID

func (n *pkgNamespaceStruct) getID() nodeID { return n.id }
func (n *pkgNameStruct) getID() nodeID      { return n.id }
func (n *pkgVersionStruct) getID() nodeID   { return n.id }
func (n *pkgVersionNode) getID() nodeID     { return n.id }
func (n *srcNamespaceStruct) getID() nodeID { return n.id }
func (n *srcNameStruct) getID() nodeID      { return n.id }
func (n *srcNameNode) getID() nodeID        { return n.id }
func (n *srcMapLink) getID() nodeID         { return n.id }

// Internal data: Packages
type pkgTypeMap map[string]*pkgNamespaceStruct
type pkgNamespaceStruct struct {
	id         nodeID
	typeKey    string
	namespaces pkgNamespaceMap
}
type pkgNamespaceMap map[string]*pkgNameStruct
type pkgNameStruct struct {
	id        nodeID
	parent    nodeID
	namespace string
	names     pkgNameMap
}
type pkgNameMap map[string]*pkgVersionStruct
type pkgVersionStruct struct {
	id         nodeID
	parent     nodeID
	name       string
	versions   pkgVersionList
	srcMapLink nodeID
}
type pkgVersionList []*pkgVersionNode
type pkgVersionNode struct {
	id         nodeID
	parent     nodeID
	version    string
	subpath    string
	srcMapLink nodeID
}

// Be type safe, don't use any / interface{}
type pkgNameOrVersion interface {
	implementsPkgNameOrVersion()
	setSrcMapLink(id nodeID)
	getSrcMapLink() nodeID
}

func (p *pkgVersionStruct) implementsPkgNameOrVersion() {}
func (p *pkgVersionNode) implementsPkgNameOrVersion()   {}
func (p *pkgVersionStruct) setSrcMapLink(id nodeID)     { p.srcMapLink = id }
func (p *pkgVersionNode) setSrcMapLink(id nodeID)       { p.srcMapLink = id }
func (p *pkgVersionStruct) getSrcMapLink() nodeID       { return p.srcMapLink }
func (p *pkgVersionNode) getSrcMapLink() nodeID         { return p.srcMapLink }

// Internal data: Sources
type srcTypeMap map[string]*srcNamespaceStruct
type srcNamespaceStruct struct {
	id         nodeID
	typeKey    string
	namespaces srcNamespaceMap
}
type srcNamespaceMap map[string]*srcNameStruct
type srcNameStruct struct {
	id        nodeID
	parent    nodeID
	namespace string
	names     srcNameList
}
type srcNameList []*srcNameNode
type srcNameNode struct {
	id         nodeID
	parent     nodeID
	name       string
	tag        *string
	commit     *string
	srcMapLink nodeID
}

// Internal data: link between sources and packages (HasSourceAt)
type srcMaps []*srcMapLink
type srcMapLink struct {
	id            nodeID
	sourceID      nodeID
	packageID     nodeID
	justification string
}

var (
	id         nodeID = 0
	index             = indexType{}
	packages          = pkgTypeMap{}
	sources           = srcTypeMap{}
	sourceMaps        = srcMaps{}
)

// In general, we would add a lock around this function
func getNextID() nodeID {
	id = id + 1
	return id
}

func (c *client) IngestPackage(ctx context.Context, input model.PackageInput) (*model.Package, error) {
	namespacesStruct, hasNamespace := packages[input.Type]
	if !hasNamespace {
		namespacesStruct = &pkgNamespaceStruct{
			id:         getNextID(),
			typeKey:    input.Type,
			namespaces: pkgNamespaceMap{},
		}
		index[namespacesStruct.id] = namespacesStruct
	}
	namespaces := namespacesStruct.namespaces

	namesStruct, hasName := namespaces[nilToEmpty(input.Namespace)]
	if !hasName {
		namesStruct = &pkgNameStruct{
			id:        getNextID(),
			parent:    namespacesStruct.id,
			namespace: nilToEmpty(input.Namespace),
			names:     pkgNameMap{},
		}
		index[namesStruct.id] = namesStruct
	}
	names := namesStruct.names

	versionStruct, hasVersions := names[input.Name]
	if !hasVersions {
		versionStruct = &pkgVersionStruct{
			id:       getNextID(),
			parent:   namesStruct.id,
			name:     input.Name,
			versions: pkgVersionList{},
		}
		index[versionStruct.id] = versionStruct
	}
	versions := versionStruct.versions

	newVersion := pkgVersionNode{
		id:      getNextID(),
		parent:  versionStruct.id,
		version: nilToEmpty(input.Version),
		subpath: nilToEmpty(input.Subpath),
	}
	index[newVersion.id] = &newVersion

	// Don't insert duplicates
	duplicate := false
	for _, v := range versions {
		if v.version == newVersion.version && v.subpath == newVersion.subpath {
			duplicate = true
			break
		}
	}
	if !duplicate {
		// Need to append to version and replace field in versionStruct
		versionStruct.versions = append(versions, &newVersion)
		// All others are refs to maps, so no need to update struct
		names[input.Name] = versionStruct
		namespaces[nilToEmpty(input.Namespace)] = namesStruct
		packages[input.Type] = namespacesStruct
	}

	// build return GraphQL type
	return buildPackageResponse(newVersion.id, nil)
}

func (c *client) Packages(ctx context.Context, filter model.PackageFilter) ([]*model.Package, error) {
	if filter.ID != nil {
		id, err := strconv.Atoi(*filter.ID)
		if err != nil {
			return nil, err
		}
		p, err := buildPackageResponse(nodeID(id), &filter)
		if err != nil {
			return nil, err
		}
		return []*model.Package{p}, nil
	}
	out := []*model.Package{}
	for dbType, namespaces := range packages {
		if noMatch(filter.Type, dbType) {
			continue
		}
		pNamespaces := []*model.PackageNamespace{}
		for namespace, names := range namespaces.namespaces {
			if noMatch(filter.Namespace, namespace) {
				continue
			}
			pns := []*model.PackageName{}
			for name, versions := range names.names {
				if noMatch(filter.Name, name) {
					continue
				}
				pvs := []*model.PackageVersion{}
				for _, v := range versions.versions {
					if noMatch(filter.Version, v.version) {
						continue
					}
					if noMatch(filter.Subpath, v.subpath) {
						continue
					}
					pv := model.PackageVersion{
						// IDs are generated as string even though we ask for integers
						// See https://github.com/99designs/gqlgen/issues/2561
						ID:      fmt.Sprintf("%d", v.id),
						Version: v.version,
						Subpath: v.subpath,
					}
					pvs = append(pvs, &pv)
				}
				pn := model.PackageName{
					// IDs are generated as string even though we ask for integers
					// See https://github.com/99designs/gqlgen/issues/2561
					ID:       fmt.Sprintf("%d", versions.id),
					Name:     name,
					Versions: pvs,
				}
				pns = append(pns, &pn)
			}
			pn := model.PackageNamespace{
				// IDs are generated as string even though we ask for integers
				// See https://github.com/99designs/gqlgen/issues/2561
				ID:        fmt.Sprintf("%d", names.id),
				Namespace: namespace,
				Names:     pns,
			}
			pNamespaces = append(pNamespaces, &pn)
		}
		p := model.Package{
			// IDs are generated as string even though we ask for integers
			// See https://github.com/99designs/gqlgen/issues/2561
			ID:         fmt.Sprintf("%d", namespaces.id),
			Type:       dbType,
			Namespaces: pNamespaces,
		}
		out = append(out, &p)
	}
	return out, nil
}

func (c *client) SourceFromPackageVersion(ctx context.Context, pkg *model.PackageVersion) (*model.Source, error) {
	id, err := strconv.Atoi(pkg.ID)
	if err != nil {
		return nil, err
	}
	internalNode, ok := index[nodeID(id)]
	if !ok {
		return nil, gqlerror.Errorf("ID does not match existing node")
	}
	internalPkg, ok := internalNode.(pkgNameOrVersion)
	if !ok {
		return nil, gqlerror.Errorf("ID does not match expected node type")
	}
	srcMapID := internalPkg.getSrcMapLink()
	srcMapNode, ok := index[srcMapID]
	if !ok {
		return nil, nil
	}
	srcMap, ok := srcMapNode.(*srcMapLink)
	if !ok {
		return nil, gqlerror.Errorf("ID does not match expected node type")
	}
	return buildSourceResponse(srcMap.sourceID, nil)
}

func (c *client) IngestSource(ctx context.Context, input model.SourceInput) (*model.Source, error) {
	namespacesStruct, hasNamespace := sources[input.Type]
	if !hasNamespace {
		namespacesStruct = &srcNamespaceStruct{
			id:         getNextID(),
			typeKey:    input.Type,
			namespaces: srcNamespaceMap{},
		}
		index[namespacesStruct.id] = namespacesStruct
	}
	namespaces := namespacesStruct.namespaces

	namesStruct, hasName := namespaces[input.Namespace]
	if !hasName {
		namesStruct = &srcNameStruct{
			id:        getNextID(),
			parent:    namespacesStruct.id,
			namespace: input.Namespace,
			names:     srcNameList{},
		}
		index[namesStruct.id] = namesStruct
	}
	names := namesStruct.names

	newSource := srcNameNode{
		id:     getNextID(),
		parent: namesStruct.id,
		name:   input.Name,
	}
	index[newSource.id] = &newSource
	if input.Tag != nil {
		tag := *input.Tag
		newSource.tag = &tag
	}
	if input.Commit != nil {
		commit := *input.Commit
		newSource.commit = &commit
	}

	// Don't insert duplicates
	duplicate := false
	for _, src := range names {
		if src.name != input.Name {
			continue
		}
		if noMatchPtrInput(input.Tag, src.tag) {
			continue
		}
		if noMatchPtrInput(input.Commit, src.commit) {
			continue
		}
		duplicate = true
		break
	}
	if !duplicate {
		namesStruct.names = append(names, &newSource)
		namespaces[input.Namespace] = namesStruct
		sources[input.Type] = namespacesStruct
	}

	// build return GraphQL type
	return buildSourceResponse(newSource.id, nil)
}

// TODO: add ID fields
func (c *client) Sources(ctx context.Context, filter model.SourceFilter) ([]*model.Source, error) {
	if filter.ID != nil {
		id, err := strconv.Atoi(*filter.ID)
		if err != nil {
			return nil, err
		}
		s, err := buildSourceResponse(nodeID(id), &filter)
		if err != nil {
			return nil, err
		}
		return []*model.Source{s}, nil
	}
	out := []*model.Source{}
	for dbType, namespaces := range sources {
		if noMatch(filter.Type, dbType) {
			continue
		}
		sNamespaces := []*model.SourceNamespace{}
		for namespace, names := range namespaces.namespaces {
			if noMatch(filter.Namespace, namespace) {
				continue
			}
			sns := []*model.SourceName{}
			for _, s := range names.names {
				if noMatch(filter.Name, s.name) {
					continue
				}
				if noMatchPtr(filter.Tag, s.tag) {
					continue
				}
				if noMatchPtr(filter.Commit, s.commit) {
					continue
				}
				newSrc := model.SourceName{
					// IDs are generated as string even though we ask for integers
					// See https://github.com/99designs/gqlgen/issues/2561
					ID:     fmt.Sprintf("%d", s.id),
					Name:   s.name,
					Tag:    s.tag,
					Commit: s.commit,
				}
				sns = append(sns, &newSrc)
			}
			sn := model.SourceNamespace{
				// IDs are generated as string even though we ask for integers
				// See https://github.com/99designs/gqlgen/issues/2561
				ID:        fmt.Sprintf("%d", names.id),
				Namespace: namespace,
				Names:     sns,
			}
			sNamespaces = append(sNamespaces, &sn)
		}
		s := model.Source{
			// IDs are generated as string even though we ask for integers
			// See https://github.com/99designs/gqlgen/issues/2561
			ID:         fmt.Sprintf("%d", namespaces.id),
			Type:       dbType,
			Namespaces: sNamespaces,
		}
		out = append(out, &s)
	}
	return out, nil
}

func (c *client) IngestSourceAt(ctx context.Context, packageArg model.PackageInput, source model.SourceInput, input model.HasSourceAtInput) (*model.HasSourceAt, error) {
	// Note: This assumes that the package and source have already been
	// ingested (and should error otherwise).
	//
	// Note: In general, we may be tempted to convert from input to filter
	// and retrieve the singleton list from the selection resolvers, but
	// this returns a GraphQL struct, not the backend struct.
	//
	// Note: Here we assume that if the source input does not contain
	// tag/commit info then we want to match at package name level,
	// otherwise at package version level. If the input contains more info
	// than what we need, we ignore it.
	matchAtPackageNameLevel := source.Tag != nil || source.Commit != nil

	srcNamespace, srcHasNamespace := sources[source.Type]
	if !srcHasNamespace {
		return nil, gqlerror.Errorf("Source type \"%s\" not found", source.Type)
	}
	srcName, srcHasName := srcNamespace.namespaces[source.Namespace]
	if !srcHasName {
		return nil, gqlerror.Errorf("Source namespace \"%s\" not found", source.Namespace)
	}
	found := false
	var sourceID nodeID
	for _, src := range srcName.names {
		if src.name != source.Name {
			continue
		}
		if noMatchPtrInput(source.Tag, src.tag) {
			continue
		}
		if noMatchPtrInput(source.Commit, src.commit) {
			continue
		}
		if found {
			return nil, gqlerror.Errorf("More than one source matches input")
		}
		sourceID = src.id
		found = true
	}
	if !found {
		return nil, gqlerror.Errorf("No source matches input")
	}

	pkgNamespace, pkgHasNamespace := packages[packageArg.Type]
	if !pkgHasNamespace {
		return nil, gqlerror.Errorf("Package type \"%s\" not found", packageArg.Type)
	}
	pkgName, pkgHasName := pkgNamespace.namespaces[nilToEmpty(packageArg.Namespace)]
	if !pkgHasName {
		return nil, gqlerror.Errorf("Package namespace \"%s\" not found", nilToEmpty(packageArg.Namespace))
	}
	pkgVersion, pkgHasVersion := pkgName.names[packageArg.Name]
	if !pkgHasVersion {
		return nil, gqlerror.Errorf("Package name \"%s\" not found", packageArg.Name)
	}
	var packageID nodeID
	if !matchAtPackageNameLevel {
		packageID = pkgVersion.id
	} else {
		found = false
		for _, version := range pkgVersion.versions {
			if noMatchInput(packageArg.Version, version.version) {
				continue
			}
			if noMatchInput(packageArg.Subpath, version.subpath) {
				continue
			}
			if found {
				return nil, gqlerror.Errorf("More than one package matches input")
			}
			packageID = version.id
			found = true
		}
		if !found {
			return nil, gqlerror.Errorf("No package matches input")
		}
	}

	// store the link
	newSrcMapLink := &srcMapLink{
		id:            getNextID(),
		sourceID:      sourceID,
		packageID:     packageID,
		justification: input.Justification,
	}
	index[newSrcMapLink.id] = newSrcMapLink
	sourceMaps = append(sourceMaps, newSrcMapLink)
	// set the backlinks
	index[packageID].(pkgNameOrVersion).setSrcMapLink(newSrcMapLink.id)
	index[sourceID].(*srcNameNode).srcMapLink = newSrcMapLink.id

	// build return GraphQL type
	p, err := buildPackageResponse(packageID, nil)
	if err != nil {
		return nil, err
	}
	s, err := buildSourceResponse(sourceID, nil)
	if err != nil {
		return nil, err
	}
	out := model.HasSourceAt{
		Package:       p,
		Source:        s,
		Justification: input.Justification,
	}

	return &out, nil
}

// TODO: add ID fields
func (c *client) SourceMap(ctx context.Context, filter model.HasSourceAtFilter) ([]*model.HasSourceAt, error) {
	out := []*model.HasSourceAt{}

	for _, mapLink := range sourceMaps {
		if noMatch(filter.Justification, mapLink.justification) {
			continue
		}
		// Note: Here we may call the selection resolvers since we
		// build GraphQL structs based on what's on the backend. But
		// this will result in then having to build a cartesian product
		// between packages (of count, say, P) and sources (of count,
		// say, C), for a total of P*C nodes than then need to be
		// compared with the, say, M sourceMaps that we select so far.
		// In general M << {P, C}, so this is wasteful.
		p, err := buildPackageResponse(mapLink.packageID, filter.Package)
		if err != nil {
			return nil, err
		}
		if p == nil {
			continue
		}
		s, err := buildSourceResponse(mapLink.sourceID, filter.Source)
		if err != nil {
			return nil, err
		}
		if s == nil {
			continue
		}
		newHSA := model.HasSourceAt{
			Package:       p,
			Source:        s,
			Justification: mapLink.justification,
		}
		out = append(out, &newHSA)
	}

	return out, nil
}

// Builds a model.Package to send as GraphQL response, starting from id.
// The optional filter allows restricting output (on selection operations).
func buildPackageResponse(id nodeID, filter *model.PackageFilter) (*model.Package, error) {
	if filter != nil && filter.ID != nil {
		filteredID, err := strconv.Atoi(*filter.ID)
		if err != nil {
			return nil, err
		}
		if nodeID(filteredID) != id {
			return nil, nil
		}
	}

	node, ok := index[id]
	if !ok {
		return nil, gqlerror.Errorf("ID does not match existing node")
	}

	pvl := []*model.PackageVersion{}
	if versionNode, ok := node.(*pkgVersionNode); ok {
		pv := model.PackageVersion{
			// IDs are generated as string even though we ask for integers
			// See https://github.com/99designs/gqlgen/issues/2561
			ID:      fmt.Sprintf("%d", versionNode.id),
			Version: versionNode.version,
			Subpath: versionNode.subpath,
		}
		if filter != nil && noMatch(filter.Version, pv.Version) {
			return nil, nil
		}
		if filter != nil && noMatch(filter.Subpath, pv.Subpath) {
			return nil, nil
		}
		pvl = append(pvl, &pv)
		node = index[versionNode.parent]
	}

	pnl := []*model.PackageName{}
	if versionStruct, ok := node.(*pkgVersionStruct); ok {
		pn := model.PackageName{
			// IDs are generated as string even though we ask for integers
			// See https://github.com/99designs/gqlgen/issues/2561
			ID:       fmt.Sprintf("%d", versionStruct.id),
			Name:     versionStruct.name,
			Versions: pvl,
		}
		if filter != nil && noMatch(filter.Name, pn.Name) {
			return nil, nil
		}
		pnl = append(pnl, &pn)
		node = index[versionStruct.parent]
	}

	pnsl := []*model.PackageNamespace{}
	if nameStruct, ok := node.(*pkgNameStruct); ok {
		pns := model.PackageNamespace{
			// IDs are generated as string even though we ask for integers
			// See https://github.com/99designs/gqlgen/issues/2561
			ID:        fmt.Sprintf("%d", nameStruct.id),
			Namespace: nameStruct.namespace,
			Names:     pnl,
		}
		if filter != nil && noMatch(filter.Namespace, pns.Namespace) {
			return nil, nil
		}
		pnsl = append(pnsl, &pns)
		node = index[nameStruct.parent]
	}

	namespaceStruct, ok := node.(*pkgNamespaceStruct)
	if !ok {
		return nil, gqlerror.Errorf("ID does not match expected node type")
	}
	p := model.Package{
		// IDs are generated as string even though we ask for integers
		// See https://github.com/99designs/gqlgen/issues/2561
		ID:         fmt.Sprintf("%d", namespaceStruct.id),
		Type:       namespaceStruct.typeKey,
		Namespaces: pnsl,
	}
	if filter != nil && noMatch(filter.Type, p.Type) {
		return nil, nil
	}
	return &p, nil
}

// Builds a model.Source to send as GraphQL response, starting from id.
// The optional filter allows restricting output (on selection operations).
func buildSourceResponse(id nodeID, filter *model.SourceFilter) (*model.Source, error) {
	if filter != nil && filter.ID != nil {
		filteredID, err := strconv.Atoi(*filter.ID)
		if err != nil {
			return nil, err
		}
		if nodeID(filteredID) != id {
			return nil, nil
		}
	}

	node, ok := index[id]
	if !ok {
		return nil, gqlerror.Errorf("ID does not match existing node")
	}

	snl := []*model.SourceName{}
	if nameNode, ok := node.(*srcNameNode); ok {
		sn := model.SourceName{
			// IDs are generated as string even though we ask for integers
			// See https://github.com/99designs/gqlgen/issues/2561
			ID:   fmt.Sprintf("%d", nameNode.id),
			Name: nameNode.name,
		}
		if nameNode.tag != nil {
			sn.Tag = nameNode.tag
		}
		if nameNode.commit != nil {
			sn.Commit = nameNode.commit
		}
		if filter != nil && noMatch(filter.Name, sn.Name) {
			return nil, nil
		}
		if filter != nil && noMatchPtr(filter.Tag, sn.Tag) {
			return nil, nil
		}
		if filter != nil && noMatchPtr(filter.Commit, sn.Commit) {
			return nil, nil
		}
		snl = append(snl, &sn)
		node = index[nameNode.parent]
	}

	snsl := []*model.SourceNamespace{}
	if nameStruct, ok := node.(*srcNameStruct); ok {
		sns := model.SourceNamespace{
			// IDs are generated as string even though we ask for integers
			// See https://github.com/99designs/gqlgen/issues/2561
			ID:        fmt.Sprintf("%d", nameStruct.id),
			Namespace: nameStruct.namespace,
			Names:     snl,
		}
		if filter != nil && noMatch(filter.Namespace, sns.Namespace) {
			return nil, nil
		}
		snsl = append(snsl, &sns)
		node = index[nameStruct.parent]
	}

	namespaceStruct, ok := node.(*srcNamespaceStruct)
	if !ok {
		return nil, gqlerror.Errorf("ID does not match expected node type")
	}
	s := model.Source{
		// IDs are generated as string even though we ask for integers
		// See https://github.com/99designs/gqlgen/issues/2561
		ID:         fmt.Sprintf("%d", namespaceStruct.id),
		Type:       namespaceStruct.typeKey,
		Namespaces: snsl,
	}
	if filter != nil && noMatch(filter.Type, s.Type) {
		return nil, nil
	}
	return &s, nil
}

func noMatch(filter *string, value string) bool {
	if filter != nil {
		return value != *filter
	}
	return false
}

func noMatchInput(filter *string, value string) bool {
	if filter != nil {
		return value != *filter
	}
	return value != ""
}

func noMatchPtr(filter *string, value *string) bool {
	if filter == nil {
		if value == nil {
			return false
		} else {
			return false
		}
	} else {
		if value == nil {
			return true
		} else {
			return *value != *filter
		}
	}
}

func noMatchPtrInput(input *string, value *string) bool {
	if input == nil {
		if value == nil {
			return false
		} else {
			return *value != ""
		}
	} else {
		if value == nil {
			return true
		} else {
			return *value != *input
		}
	}
}

func nilToEmpty(input *string) string {
	if input == nil {
		return ""
	}
	return *input
}
