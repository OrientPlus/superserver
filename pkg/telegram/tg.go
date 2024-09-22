package telegram

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type tg interface {
	UpdateCh(config UpdateConfig) UpdatesChannel
}

type UpdatesChannel <-chan tgbotapi.Update

type UpdateConfig struct {
	Offset  int
	Limit   int
	Timeout int
}

type Update struct {
	UpdateID           int                 `json:"update_id"`
	Message            *Message            `json:"message,omitempty"`
	EditedMessage      *Message            `json:"edited_message,omitempty"`
	ChannelPost        *Message            `json:"channel_post,omitempty"`
	EditedChannelPost  *Message            `json:"edited_channel_post,omitempty"`
	InlineQuery        *InlineQuery        `json:"inline_query,omitempty"`
	ChosenInlineResult *ChosenInlineResult `json:"chosen_inline_result,omitempty"`
	CallbackQuery      *CallbackQuery      `json:"callback_query,omitempty"`
	ShippingQuery      *ShippingQuery      `json:"shipping_query,omitempty"`
	PreCheckoutQuery   *PreCheckoutQuery   `json:"pre_checkout_query,omitempty"`
	Poll               *Poll               `json:"poll,omitempty"`
	PollAnswer         *PollAnswer         `json:"poll_answer,omitempty"`
	MyChatMember       *ChatMemberUpdated  `json:"my_chat_member"`
	ChatMember         *ChatMemberUpdated  `json:"chat_member"`
	ChatJoinRequest    *ChatJoinRequest    `json:"chat_join_request"`
}
