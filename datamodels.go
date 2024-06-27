package main

type (
	SignalMessageMentions struct {
		Start  int64  `json:"start"`
		Length int64  `json:"length"`
		Author string `json:"author"`
	}
	SignalSendMessageV2 struct {
		Base64Attachments []string                `json:"base64_attachments,omitempty"`
		EditTimestamp     uint64                  `json:"edit_timestamp,omitempty"`
		Mentions          []SignalMessageMentions `json:"mentions,omitempty"`
		Message           string                  `json:"message"`
		Number            string                  `json:"number"`
		QuoteAuthor       string                  `json:"quote_author,omitempty"`
		QuoteMentions     []SignalMessageMentions `json:"quote_mentions,omitempty"`
		QuoteMessage      string                  `json:"quote_message,omitempty"`
		QuoteTimestamp    uint64                  `json:"quote_timestamp,omitempty"`
		Recipients        []string                `json:"recipients"`
		Sticker           string                  `json:"sticker,omitempty"`
		TextMode          string                  `json:"text_mode,omitempty"`
	}

	SignalGroupInfo struct {
		GroupId string `json:"groupId"`
		Type    string `json:"type"`
	}
	SignalAttachments struct {
		ContentType     string `json:"contentType"`
		Filename        string `json:"filename"`
		Id              string `json:"id"`
		Size            uint64 `json:"size"`
		Width           uint   `json:"width"`
		Height          uint   `json:"height"`
		Caption         string `json:"caption,omitempty"`
		UploadTimestamp uint64 `json:"uploadTimestamp,omitempty"`
	}
	SignalDataMessage struct {
		Timestamp        uint64              `json:"timestamp"`
		Message          string              `json:"message"`
		ExpiresInSeconds uint64              `json:"expiresInSeconds"`
		ViewOnce         bool                `json:"viewOnce"`
		Attachments      []SignalAttachments `json:"attachments"`
		GroupInfo        SignalGroupInfo     `json:"groupInfo"`
	}
	SignalEnvelope struct {
		Source       string            `json:"source"`
		SourceNumber string            `json:"sourceNumber"`
		SourceUuid   string            `json:"sourceUuid"`
		SourceName   string            `json:"sourceName"`
		SourceDevice uint              `json:"sourceDevice"`
		Timestamp    uint64            `json:"timestamp"`
		SyncMessage  any               `json:"syncMessage"`
		DataMessage  SignalDataMessage `json:"dataMessage"`
	}
	SignalMessage struct {
		Envelope SignalEnvelope `json:"envelope"`
		Account  string         `json:"account"`
	}

	SignalGroupEntry struct {
		Admins          []string `json:"admins"`
		Blocked         bool     `json:"blocked"`
		Id              string   `json:"id"`
		InternalId      string   `json:"internal_id"`
		InviteLink      string   `json:"invite_link"`
		Members         []string `json:"members"`
		Name            string   `json:"name"`
		PendingInvites  []string `json:"pending_invites"`
		PendingRequests []string `json:"pending_requests"`
	}
)
