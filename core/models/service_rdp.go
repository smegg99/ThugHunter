// core/models/service_rdp.go
//
// RDP service model.
package models

// RDPService stores probe results for an RDP endpoint.
type RDPService struct {
	ServiceBase
}

func (s *RDPService) Type() ServiceType  { return ServiceTypeRDP }
func (s *RDPService) Base() *ServiceBase { return &s.ServiceBase }
