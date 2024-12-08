package capybara

type service struct {
	ID       string   `json:"id"`
	Pattern  string   `json:"pattern"`
	Method   string   `json:"method"`
	Port     int32    `json:"port"`
	Redirect string   `json:"redirect"`
	Protocol Protocol `json:"protocol"`
}

func (s *service) FixProtocol() {
	if s.Protocol == "" {
		s.Protocol = defaultProtocol
	}
}

func (s *service) FixMethod() {
	// Unspecified method in json. Using default
	if s.Method == "" {
		s.Method = defaultMethod
	}
}
