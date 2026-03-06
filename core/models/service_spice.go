// core/models/crawler_service_spice.go
package models

type SPICEService struct {
	ServiceBase
}

func (s *SPICEService) Type() ServiceType  { return ServiceTypeSPICE }
func (s *SPICEService) Base() *ServiceBase { return &s.ServiceBase }
