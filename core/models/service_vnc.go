// core/models/crawler_service_vnc.go
package models

type VNCService struct {
	ServiceBase
}

func (s *VNCService) Type() ServiceType  { return ServiceTypeVNC }
func (s *VNCService) Base() *ServiceBase { return &s.ServiceBase }