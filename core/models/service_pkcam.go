// core/models/scraper_service_pkcam.go
package models

type PKCameraService struct {
	ServiceBase
	HasTwoWayAudio bool `json:"has_two_way_audio"`
}

func (s *PKCameraService) Type() ServiceType  { return ServiceTypePKCamera }
func (s *PKCameraService) Base() *ServiceBase { return &s.ServiceBase }
