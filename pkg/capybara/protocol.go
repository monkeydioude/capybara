package capybara

const defaultProtocol = HttpProtocol

type Protocol string

const (
	HttpProtocol Protocol = "http"
	RpcProtocol  Protocol = "rpc"

	// Http3Protocol = "http3"
)
