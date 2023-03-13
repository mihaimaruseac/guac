package memory

import (
	"context"

	"github.com/vektah/gqlparser/v2/gqlerror"

	"layout/server/graphql/model"
)

// Internal data: Packages
type pkgTypeMap map[string]pkgNamespaceMap
type pkgNamespaceMap map[string]pkgNameMap
type pkgNameMap map[string]*pkgVersionStruct
type pkgVersionStruct struct {
	source   *srcMapLink
	versions pkgVersionList
}
type pkgVersionList []*pkgVersionNode
type pkgVersionNode struct {
	version string
	subpath string
	source  *srcMapLink
}

// Be type safe, don't use any / interface{}
type pkgNameOrVersion interface {
	implementsPkgNameOrVersion()
	setSource(link *srcMapLink)
}

func (p *pkgVersionStruct) implementsPkgNameOrVersion() {}
func (p *pkgVersionNode) implementsPkgNameOrVersion()   {}
func (p *pkgVersionStruct) setSource(link *srcMapLink)  { p.source = link }
func (p *pkgVersionNode) setSource(link *srcMapLink)    { p.source = link }

// Internal data: Sources
type srcTypeMap map[string]srcNamespaceMap
type srcNamespaceMap map[string]srcNameList
type srcNameList []*srcNameNode
type srcNameNode struct {
	name   string
	tag    *string
	commit *string
	pkg    *srcMapLink
}

// Internal data: link between sources and packages (HasSourceAt)
type srcMaps []*srcMapLink
type srcMapLink struct {
	justification string
	source        *srcNameNode
	pkg           pkgNameOrVersion
}

var packages = pkgTypeMap{}
var sources = srcTypeMap{}
var sourceMaps = srcMaps{}

func (c *client) IngestPackage(ctx context.Context, input model.PackageInput) (*model.Package, error) {
	namespaces, hasNamespace := packages[input.Type]
	names, hasName := namespaces[nilToEmpty(input.Namespace)]
	versionStruct, hasVersions := names[input.Name]
	versions := pkgVersionList{}
	if hasVersions {
		versions = versionStruct.versions
	}
	newVersion := pkgVersionNode{
		version: nilToEmpty(input.Version),
		subpath: nilToEmpty(input.Subpath),
	}

	// Don't insert duplicates
	duplicate := false
	for _, v := range versions {
		if v.version == newVersion.version && v.subpath == newVersion.subpath {
			duplicate = true
			break
		}
	}
	if !duplicate {
		versions = append(versions, &newVersion)
		if !hasNamespace {
			packages[input.Type] = pkgNamespaceMap{}
		}
		if !hasName {
			packages[input.Type][nilToEmpty(input.Namespace)] = pkgNameMap{}
		}
		if !hasVersions {
			versionStruct = &pkgVersionStruct{}
		}
		versionStruct.versions = versions
		packages[input.Type][nilToEmpty(input.Namespace)][input.Name] = versionStruct
	}

	// build return GraphQL type
	out := packageFromInput(input)
	return out, nil
}

func (c *client) Packages(ctx context.Context, filter model.PackageFilter) ([]*model.Package, error) {
	out := []*model.Package{}
	for dbType, namespaces := range packages {
		if noMatch(filter.Type, dbType) {
			continue
		}
		pNamespaces := []*model.PackageNamespace{}
		for namespace, names := range namespaces {
			if noMatch(filter.Namespace, namespace) {
				continue
			}
			pns := []*model.PackageName{}
			for name, versions := range names {
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
						Version: v.version,
						Subpath: v.subpath,
					}
					pvs = append(pvs, &pv)
				}
				pn := model.PackageName{
					Name:     name,
					Versions: pvs,
				}
				pns = append(pns, &pn)
			}
			pn := model.PackageNamespace{
				Namespace: namespace,
				Names:     pns,
			}
			pNamespaces = append(pNamespaces, &pn)
		}
		p := model.Package{
			Type:       dbType,
			Namespaces: pNamespaces,
		}
		out = append(out, &p)
	}
	return out, nil
}

func (c *client) IngestSource(ctx context.Context, input model.SourceInput) (*model.Source, error) {
	namespaces, hasNamespace := sources[input.Type]
	names := namespaces[input.Namespace]
	newSource := srcNameNode{
		name: input.Name,
	}
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
		names = append(names, &newSource)
		if !hasNamespace {
			sources[input.Type] = srcNamespaceMap{}
		}
		sources[input.Type][input.Namespace] = names
	}

	// build return GraphQL type
	out := sourceFromInput(input)
	return out, nil
}

func (c *client) Sources(ctx context.Context, filter model.SourceFilter) ([]*model.Source, error) {
	out := []*model.Source{}
	for dbType, namespaces := range sources {
		if noMatch(filter.Type, dbType) {
			continue
		}
		sNamespaces := []*model.SourceNamespace{}
		for namespace, names := range namespaces {
			if noMatch(filter.Namespace, namespace) {
				continue
			}
			sns := []*model.SourceName{}
			for _, s := range names {
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
					Name:   s.name,
					Tag:    s.tag,
					Commit: s.commit,
				}
				sns = append(sns, &newSrc)
			}
			sn := model.SourceNamespace{
				Namespace: namespace,
				Names:     sns,
			}
			sNamespaces = append(sNamespaces, &sn)
		}
		s := model.Source{
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
	srcName, srcHasName := srcNamespace[source.Namespace]
	if !srcHasName {
		return nil, gqlerror.Errorf("Source namespace \"%s\" not found", source.Namespace)
	}
	var srcPtr *srcNameNode
	srcPtr = nil
	for _, src := range srcName {
		if src.name != source.Name {
			continue
		}
		if noMatchPtrInput(source.Tag, src.tag) {
			continue
		}
		if noMatchPtrInput(source.Commit, src.commit) {
			continue
		}
		if srcPtr != nil {
			return nil, gqlerror.Errorf("More than one source matches input")
		}
		srcPtr = src
	}
	if srcPtr == nil {
		return nil, gqlerror.Errorf("No source matches input")
	}

	pkgNamespace, pkgHasNamespace := packages[packageArg.Type]
	if !pkgHasNamespace {
		return nil, gqlerror.Errorf("Package type \"%s\" not found", packageArg.Type)
	}
	pkgName, pkgHasName := pkgNamespace[nilToEmpty(packageArg.Namespace)]
	if !pkgHasName {
		return nil, gqlerror.Errorf("Package namespace \"%s\" not found", nilToEmpty(packageArg.Namespace))
	}
	pkgVersion, pkgHasVersion := pkgName[packageArg.Name]
	if !pkgHasVersion {
		return nil, gqlerror.Errorf("Package name \"%s\" not found", packageArg.Name)
	}
	var pkgPtr pkgNameOrVersion
	if !matchAtPackageNameLevel {
		pkgPtr = pkgVersion
	} else {
		pkgPtr = nil
		for _, version := range pkgVersion.versions {
			if noMatchInput(packageArg.Version, version.version) {
				continue
			}
			if noMatchInput(packageArg.Subpath, version.subpath) {
				continue
			}
			if pkgPtr != nil {
				return nil, gqlerror.Errorf("More than one package matches input")
			}
			pkgPtr = version
		}
	}
	if pkgPtr == nil {
		return nil, gqlerror.Errorf("No package matches input")
	}

	// store the link
	newSrcMapLink := &srcMapLink{
		justification: input.Justification,
		source:        srcPtr,
		pkg:           pkgPtr,
	}
	pkgPtr.setSource(newSrcMapLink)
	srcPtr.pkg = newSrcMapLink
	sourceMaps = append(sourceMaps, newSrcMapLink)

	// build return GraphQL type
	pkg := packageFromInput(packageArg)
	src := sourceFromInput(source)
	out := model.HasSourceAt{
		Package:       pkg,
		Source:        src,
		Justification: input.Justification,
	}

	return &out, nil
}

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
		p := packageMatchingFilter(filter.Package, mapLink.pkg)
		if p == nil {
			continue
		}
		s := sourceMatchingFilter(filter.Source, mapLink.source)
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

func packageFromInput(input model.PackageInput) *model.Package {
	pv := model.PackageVersion{
		Version: nilToEmpty(input.Version),
		Subpath: nilToEmpty(input.Subpath),
	}
	pn := model.PackageName{
		Name:     input.Name,
		Versions: []*model.PackageVersion{&pv},
	}
	pns := model.PackageNamespace{
		Namespace: nilToEmpty(input.Namespace),
		Names:     []*model.PackageName{&pn},
	}
	return &model.Package{
		Type:       input.Type,
		Namespaces: []*model.PackageNamespace{&pns},
	}
}

func packageMatchingFilter(filter *model.PackageFilter, packageArg pkgNameOrVersion) *model.Package {
	var out *model.Package
	out = nil

	for dbType, namespaces := range packages {
		if filter != nil && noMatch(filter.Type, dbType) {
			return nil
		}
		foundNamespace := false
		pNamespaces := []*model.PackageNamespace{}
		for namespace, names := range namespaces {
			if filter != nil && noMatch(filter.Namespace, namespace) {
				return nil
			}
			foundName := false
			pns := []*model.PackageName{}
			for name, versions := range names {
				if filter != nil && noMatch(filter.Name, name) {
					return nil
				}
				if packageArg == versions {
					pn := model.PackageName{
						Name: name,
					}
					pns = append(pns, &pn)
					foundName = true
				} else {
					pvs := []*model.PackageVersion{}
					foundVersion := false
					for _, v := range versions.versions {
						if filter != nil && noMatch(filter.Version, v.version) {
							return nil
						}
						if filter != nil && noMatch(filter.Subpath, v.subpath) {
							return nil
						}
						if packageArg != v {
							continue
						}
						pv := model.PackageVersion{
							Version: v.version,
							Subpath: v.subpath,
						}
						pvs = append(pvs, &pv)
						foundVersion = true
					}
					if foundVersion {
						pn := model.PackageName{
							Name:     name,
							Versions: pvs,
						}
						pns = append(pns, &pn)
						foundName = true
					}
				}
			}
			if foundName {
				pn := model.PackageNamespace{
					Namespace: namespace,
					Names:     pns,
				}
				pNamespaces = append(pNamespaces, &pn)
				foundNamespace = true
			}
		}
		if foundNamespace {
			out = &model.Package{
				Type:       dbType,
				Namespaces: pNamespaces,
			}
		}
	}

	return out
}

func sourceFromInput(input model.SourceInput) *model.Source {
	sn := model.SourceName{
		Name:   input.Name,
		Tag:    input.Tag,
		Commit: input.Commit,
	}
	sns := model.SourceNamespace{
		Namespace: input.Namespace,
		Names:     []*model.SourceName{&sn},
	}
	return &model.Source{
		Type:       input.Type,
		Namespaces: []*model.SourceNamespace{&sns},
	}
}

func sourceMatchingFilter(filter *model.SourceFilter, source *srcNameNode) *model.Source {
	var out *model.Source
	out = nil

	for dbType, namespaces := range sources {
		if filter != nil && noMatch(filter.Type, dbType) {
			continue
		}
		foundNamespace := false
		sNamespaces := []*model.SourceNamespace{}
		for namespace, names := range namespaces {
			if filter != nil && noMatch(filter.Namespace, namespace) {
				continue
			}
			foundName := false
			sns := []*model.SourceName{}
			for _, s := range names {
				if filter != nil && noMatch(filter.Name, s.name) {
					continue
				}
				if filter != nil && noMatchPtr(filter.Tag, s.tag) {
					continue
				}
				if filter != nil && noMatchPtr(filter.Commit, s.commit) {
					continue
				}
				if source != s {
					continue
				}
				newSrc := model.SourceName{
					Name:   s.name,
					Tag:    s.tag,
					Commit: s.commit,
				}
				sns = append(sns, &newSrc)
				foundName = true
			}
			if foundName {
				sn := model.SourceNamespace{
					Namespace: namespace,
					Names:     sns,
				}
				sNamespaces = append(sNamespaces, &sn)
				foundNamespace = true
			}
		}
		if foundNamespace {
			out = &model.Source{
				Type:       dbType,
				Namespaces: sNamespaces,
			}
		}
	}

	return out
}
