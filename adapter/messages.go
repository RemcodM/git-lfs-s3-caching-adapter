package adapter

// inputMessage represents a message from Git LFS to the standalone transfer
// agent. Not all fields will be filled in on all requests.
type inputMessage struct {
	Event     string `json:"event"`
	Operation string `json:"operation"`
	Remote    string `json:"remote"`
	Oid       string `json:"oid"`
	Size      int64  `json:"size"`
	Path      string `json:"path"`
}

// completeMessage represents a completion response.
type progressMessage struct {
	Event          string `json:"event"`
	Oid            string `json:"oid"`
	BytesSoFar     int64  `json:"bytesSoFar"`
	BytesSinceLast int64  `json:"bytesSinceLast"`
}

// errorMessage represents an optional error message that may occur in a
// completion response.
type errorMessage struct {
	Message string `json:"message"`
}

// outputErrorMessage represents an error message that may occur during startup.
type outputErrorMessage struct {
	Error errorMessage `json:"error"`
}

// completeMessage represents a completion response.
type completeMessage struct {
	Event string        `json:"event"`
	Oid   string        `json:"oid"`
	Path  string        `json:"path,omitempty"`
	Error *errorMessage `json:"error,omitempty"`
}
