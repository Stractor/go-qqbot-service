package service

type QQ_MSG_POST_TYPE string

const (
	POST_TYPE_MESSAGE    QQ_MSG_POST_TYPE = "message"
	POST_TYPE_META_EVENT QQ_MSG_POST_TYPE = "meta_event" //不想做meta_event的处理
)

type QQ_MSG_MESSAGE_TYPE string

const (
	MESSAGE_TYPE_GROUP   QQ_MSG_MESSAGE_TYPE = "group"
	MESSAGE_TYPE_PRIVATE QQ_MSG_MESSAGE_TYPE = "private"
)

type QQ_MSG_SUB_TYPE string

const (
	//现在只处理这两个
	QQ_SUB_TYPE_NORMAL QQ_MSG_SUB_TYPE = "normal" // 群聊直接发言和@都是normal（可能群主不一样？）
	QQ_SUB_TYPE_FRIEND QQ_MSG_SUB_TYPE = "friend" // 私聊，好友发言
)

type QQMessageType struct {
	PostType    string          `json:"post_type"`
	MessageType string          `json:"message_type"`
	Time        int64           `json:"time"`
	SelfID      int64           `json:"self_id"`
	SubType     string          `json:"sub_type"`
	Message     string          `json:"message"`
	RawMessage  string          `json:"raw_message"`
	Sender      QQMessageSender `json:"sender"`
	MessageID   int64           `json:"message_id"`
	Font        int             `json:"font"`
	GroupID     int64           `json:"group_id"`
	UserID      int64           `json:"user_id"`
	MessageSeq  int64           `json:"message_seq"`
}

type QQMessageSender struct {
	Age      int    `json:"age"`
	Area     string `json:"area"`
	Card     string `json:"card"`
	Level    string `json:"level"`
	Nickname string `json:"nickname"`
	Role     string `json:"role"`
	Sex      string `json:"sex"`
	Title    string `json:"title"`
	UserID   int64  `json:"user_id"`
}

type QQMessageSendModel struct {
	Action string              `json:"action"`
	Params QQMessageSendDetail `json:"params"`
}

type QQMessageSendDetail struct {
	MessageType QQ_MSG_MESSAGE_TYPE `json:"message_ype"`
	UserId      int64               `json:"user_id"`
	Message     string              `json:"message"`
}
