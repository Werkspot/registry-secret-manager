package registry

// Registry represents a container registry.
type Registry interface {
	Login() (*Credentials, error)
}
