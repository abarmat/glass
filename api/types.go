package api

// ServerStatus is a struct to hold the server status
type ServerStatus struct {
	Name              string `json:"name"`
	Version           string `json:"version"`
	CurrentTime       int64  `json:"currentTime"`
	LastImmutableTime int    `json:"lastImmutableTime"`
}

// HistoryEntry is a struct representing an entry in the history log
type HistoryEntry struct {
	ServerName string `json:"serverName"`
	EntityType string `json:"entityType"`
	EntityID   string `json:"entityId"`
	Timestamp  int64  `json:"timestamp"`
}

// SceneEntity is a struct representing a scene
type SceneEntity struct {
	ID        string   `json:"id"`
	Type      string   `json:"type"`
	Timestamp int64    `json:"timestamp"`
	Pointers  []string `json:"pointers"`
	Content   []struct {
		File string `json:"file"`
		Hash string `json:"hash"`
	} `json:"content"`
	Metadata struct {
		Display struct {
			Title   string `json:"title"`
			Favicon string `json:"favicon"`
		} `json:"display"`
		Contact struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		} `json:"contact"`
		Owner string `json:"owner"`
		Scene struct {
			Parcels []string `json:"parcels"`
			Base    string   `json:"base"`
		} `json:"scene"`
		Communications struct {
			Type       string `json:"type"`
			Signalling string `json:"signalling"`
		} `json:"communications"`
		Policy struct {
			ContentRating    string        `json:"contentRating"`
			Fly              bool          `json:"fly"`
			VoiceEnabled     bool          `json:"voiceEnabled"`
			Blacklist        []interface{} `json:"blacklist"`
			TeleportPosition string        `json:"teleportPosition"`
		} `json:"policy"`
		Main string        `json:"main"`
		Tags []interface{} `json:"tags"`
	} `json:"metadata"`
}

// AuditInfo is a struct that holds entity audit information
type AuditInfo struct {
	DeployedTimestamp int64 `json:"deployedTimestamp"`
	AuthChain         []struct {
		Type      string `json:"type"`
		Payload   string `json:"payload"`
		Signature string `json:"signature"`
	} `json:"authChain"`
	Version          string `json:"version"`
	OriginalMetadata struct {
		OriginalVersion string `json:"originalVersion"`
		Data            struct {
			OriginalRootCid   string `json:"originalRootCid"`
			OriginalAuthor    string `json:"originalAuthor"`
			OriginalSignature string `json:"originalSignature"`
			OriginalTimestamp int    `json:"originalTimestamp"`
		} `json:"data"`
	} `json:"originalMetadata"`
}
