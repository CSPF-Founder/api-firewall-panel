package endpointstatus

// EndpointStatus represents different endpoint statuses.
type EndpointStatus bool

// Constants for different endpoint statuses.
const (
	EXITED EndpointStatus = false
	UP     EndpointStatus = true
)

// EnumMap maps EndpointStatus constants to their string representations.
var EnumMap = map[EndpointStatus]string{
	UP:     "Running",
	EXITED: "Not Running",
}

// GetString returns the string representation of a EndpointStatus value.
func (j EndpointStatus) GetText() string {
	if str, ok := EnumMap[j]; ok {
		return str
	}
	return "Invalid Endpoint Status"
}
