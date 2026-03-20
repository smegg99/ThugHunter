// core/models/service_pkcam.go
//
// P2P/K camera service model.
package models

// PKCameraService stores probe results for a P2P/K camera endpoint.
type PKCameraService struct {
	ServiceBase
	HasTwoWayAudio bool `json:"has_two_way_audio"`
}

func (s *PKCameraService) Type() ServiceType  { return ServiceTypePKCamera }
func (s *PKCameraService) Base() *ServiceBase { return &s.ServiceBase }
