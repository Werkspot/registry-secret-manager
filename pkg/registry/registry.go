package registry

type Registry interface {
	Login() (*Credentials, error)
}
