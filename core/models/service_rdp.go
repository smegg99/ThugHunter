// core/models/scraper_service_rdp.go
package models

type RDPService struct {
	ServiceBase
}

func (s *RDPService) Type() ServiceType  { return ServiceTypeRDP }
func (s *RDPService) Base() *ServiceBase { return &s.ServiceBase }
