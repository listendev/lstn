package npm

// Deps gets you the package lock dependencies.
func (p *packageLockJSON) Deps() map[string]PackageLockDependency {
	return p.Dependencies
}
