package model

import "time"

type PanelUser struct {
	ID           int64      `json:"id"`
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	DisplayName  string     `json:"displayName"`
	Role         string     `json:"role"`
	Theme        string     `json:"theme"`
	BannerPreset string     `json:"bannerPreset"`
	AccentColor  string     `json:"accentColor"`
	AvatarURL    string     `json:"avatarUrl"`
	Bio          string     `json:"bio"`
	Status       string     `json:"status,omitempty"`
	StatusText   string     `json:"statusText,omitempty"`
	CreatedBy    int64      `json:"createdBy,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
	LastLoginAt  *time.Time `json:"lastLoginAt,omitempty"`
	PasswordHash string     `json:"-"`
}

type PanelSession struct {
	ID        string    `json:"id"`
	UserID    int64     `json:"userId"`
	CreatedAt time.Time `json:"createdAt"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type PanelRoom struct {
	ID                 int64      `json:"id"`
	Slug               string     `json:"slug"`
	Name               string     `json:"name"`
	Description        string     `json:"description"`
	Icon               string     `json:"icon"`
	Category           string     `json:"category"`
	Scope              string     `json:"scope"`
	SortOrder          int        `json:"sortOrder"`
	AdminOnly          bool       `json:"adminOnly"`
	VIPOnly            bool       `json:"vipOnly"`
	PasswordProtected  bool       `json:"passwordProtected"`
	PasswordHash       string     `json:"-"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
	LastMessageAt      *time.Time `json:"lastMessageAt,omitempty"`
	LastMessagePreview string     `json:"lastMessagePreview,omitempty"`
	PeerUserID         int64      `json:"peerUserId,omitempty"`
}

type PanelAttachment struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	Kind        string `json:"kind"`
	ContentType string `json:"contentType"`
	SizeBytes   int64  `json:"sizeBytes"`
	Extension   string `json:"extension,omitempty"`
	Width       int    `json:"width,omitempty"`
	Height      int    `json:"height,omitempty"`
}

type PanelReaction struct {
	Emoji         string `json:"emoji"`
	Count         int    `json:"count"`
	ViewerReacted bool   `json:"viewerReacted"`
}

type PanelReplyPreview struct {
	MessageID   int64  `json:"messageId"`
	AuthorName  string `json:"authorName"`
	BodyPreview string `json:"bodyPreview"`
}

type PanelMessage struct {
	ID              int64              `json:"id"`
	RoomID          int64              `json:"roomId"`
	AuthorID        int64              `json:"authorId"`
	AuthorName      string             `json:"authorName"`
	AuthorRole      string             `json:"authorRole"`
	Body            string             `json:"body"`
	Kind            string             `json:"kind"`
	IsAI            bool               `json:"isAI"`
	ReplyToID       int64              `json:"replyToId,omitempty"`
	Reply           *PanelReplyPreview `json:"reply,omitempty"`
	Reactions       []PanelReaction    `json:"reactions,omitempty"`
	Attachment      *PanelAttachment   `json:"attachment,omitempty"`
	CreatedAt       time.Time          `json:"createdAt"`
	UpdatedAt       *time.Time         `json:"updatedAt,omitempty"`
	IsPinned        bool               `json:"isPinned,omitempty"`
	ViewerFavorited bool               `json:"viewerFavorited,omitempty"`
	BlockedByViewer bool               `json:"blockedByViewer,omitempty"`
}

type PanelPresence struct {
	UserID           int64     `json:"userId"`
	Username         string    `json:"username"`
	DisplayName      string    `json:"displayName"`
	Role             string    `json:"role"`
	Theme            string    `json:"theme"`
	BannerPreset     string    `json:"bannerPreset"`
	AccentColor      string    `json:"accentColor"`
	AvatarURL        string    `json:"avatarUrl"`
	Bio              string    `json:"bio"`
	Status           string    `json:"status"`
	StatusText       string    `json:"statusText,omitempty"`
	RoomID           int64     `json:"roomId"`
	LastSeenAt       time.Time `json:"lastSeenAt"`
	Online           bool      `json:"online"`
	BlockedByViewer  bool      `json:"blockedByViewer,omitempty"`
	MutedByViewer    bool      `json:"mutedByViewer,omitempty"`
	HasBlockedViewer bool      `json:"hasBlockedViewer,omitempty"`
}

type PanelBootstrap struct {
	SessionID      string            `json:"sessionId,omitempty"`
	Viewer         PanelUser         `json:"viewer"`
	Rooms          []PanelRoom       `json:"rooms"`
	Online         []PanelPresence   `json:"online"`
	Typing         []PanelTyping     `json:"typing,omitempty"`
	RecentLogs     []PanelLogItem    `json:"recentLogs,omitempty"`
	LatestMessages []PanelMessage    `json:"latestMessages,omitempty"`
	Events         []PanelEvent      `json:"events,omitempty"`
	RoomVersions   map[string]int64  `json:"roomVersions,omitempty"`
	BlockedUserIDs []int64           `json:"blockedUserIds,omitempty"`
	MutedUserIDs   []int64           `json:"mutedUserIds,omitempty"`
	RoomAccess     map[string]string `json:"roomAccess"`
	ServerTime     time.Time         `json:"serverTime"`
	Version        int64             `json:"version"`
}

type PanelTyping struct {
	RoomID      int64     `json:"roomId"`
	UserID      int64     `json:"userId"`
	DisplayName string    `json:"displayName"`
	Role        string    `json:"role"`
	ExpiresAt   time.Time `json:"expiresAt"`
}

type PanelLogItem struct {
	ID        int64     `json:"id"`
	Action    string    `json:"action"`
	ActorID   int64     `json:"actorId"`
	ActorName string    `json:"actorName"`
	RoomID    int64     `json:"roomId,omitempty"`
	RoomSlug  string    `json:"roomSlug,omitempty"`
	Detail    string    `json:"detail"`
	CreatedAt time.Time `json:"createdAt"`
}

type PanelEvent struct {
	ID            int64     `json:"id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	RoomID        int64     `json:"roomId,omitempty"`
	RoomName      string    `json:"roomName,omitempty"`
	CreatedBy     int64     `json:"createdBy"`
	CreatedByName string    `json:"createdByName"`
	StartsAt      time.Time `json:"startsAt"`
	CreatedAt     time.Time `json:"createdAt"`
	RSVPCount     int       `json:"rsvpCount"`
	ViewerJoined  bool      `json:"viewerJoined"`
}

type PanelPollOption struct {
	ID          int64  `json:"id"`
	Label       string `json:"label"`
	Votes       int    `json:"votes"`
	ViewerVoted bool   `json:"viewerVoted,omitempty"`
}

type PanelPoll struct {
	ID             int64             `json:"id"`
	RoomID         int64             `json:"roomId"`
	Question       string            `json:"question"`
	CreatedBy      int64             `json:"createdBy"`
	CreatedByName  string            `json:"createdByName"`
	CreatedAt      time.Time         `json:"createdAt"`
	Options        []PanelPollOption `json:"options"`
	ViewerOptionID int64             `json:"viewerOptionId,omitempty"`
	TotalVotes     int               `json:"totalVotes"`
}

type PanelSearchResult struct {
	Query    string         `json:"query"`
	Rooms    []PanelRoom    `json:"rooms"`
	Users    []PanelUser    `json:"users"`
	Messages []PanelMessage `json:"messages"`
}

type PanelSocialProfile struct {
	User        PanelPresence `json:"user"`
	CanDM       bool          `json:"canDm"`
	CanManage   bool          `json:"canManage"`
	CanModerate bool          `json:"canModerate"`
}

type PanelJoinRequest struct {
	ID                int64      `json:"id"`
	Email             string     `json:"email"`
	DisplayName       string     `json:"displayName"`
	Note              string     `json:"note"`
	Status            string     `json:"status"`
	ReviewNote        string     `json:"reviewNote,omitempty"`
	RequestedAt       time.Time  `json:"requestedAt"`
	ReviewedAt        *time.Time `json:"reviewedAt,omitempty"`
	ReviewedBy        int64      `json:"reviewedBy,omitempty"`
	ReviewerName      string     `json:"reviewerName,omitempty"`
	AccessCodeExpires *time.Time `json:"accessCodeExpires,omitempty"`
	ApprovedUserID    int64      `json:"approvedUserId,omitempty"`
	EmailSent         bool       `json:"emailSent,omitempty"`
	AccessCode        string     `json:"accessCode,omitempty"`
}

type PanelOpsSummary struct {
	Users          int       `json:"users"`
	Rooms          int       `json:"rooms"`
	Messages       int       `json:"messages"`
	Events         int       `json:"events"`
	Polls          int       `json:"polls"`
	OnlineUsers    int       `json:"onlineUsers"`
	ActiveSessions int       `json:"activeSessions"`
	UploadFiles    int       `json:"uploadFiles"`
	UploadBytes    int64     `json:"uploadBytes"`
	Version        int64     `json:"version"`
	GeneratedAt    time.Time `json:"generatedAt"`
}

type PanelTerminalResult struct {
	Command  string    `json:"command"`
	Output   string    `json:"output"`
	ExitCode int       `json:"exitCode"`
	RanAt    time.Time `json:"ranAt"`
}
