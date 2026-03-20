// core/models/service_spice.go
//
// SPICE service model.
package models

// SPICEService stores probe results for a SPICE endpoint.
type SPICEService struct {
	ServiceBase
}

func (s *SPICEService) Type() ServiceType  { return ServiceTypeSPICE }
func (s *SPICEService) Base() *ServiceBase { return &s.ServiceBase }
