// core/scanner/services.go
package scanner

// enabledProbes is the ordered list of service probes to run on every host.
// Add new probes here to extend the scanner.
var enabledProbes = []Probe{
	&VNCProbe{},
}
