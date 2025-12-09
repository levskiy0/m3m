package main

import (
	"m3m/pkg/schema"
)

// GetSchema returns the schema for TypeScript generation
func (p *TelegramPlugin) GetSchema() schema.ModuleSchema {
	return schema.ModuleSchema{
		Name:        "$telegram",
		Description: "Telegram Bot API plugin",
		Methods: []schema.MethodSchema{
			{
				Name:        "startBot",
				Description: "Start a new Telegram bot",
				Params: []schema.ParamSchema{
					{Name: "token", Type: "string", Description: "Bot token from @BotFather"},
					{Name: "setup", Type: "(instance: TelegramBotInstance) => void", Description: "Setup callback"},
				},
			},
			{
				Name:        "stopBot",
				Description: "Stop a bot by token",
				Params: []schema.ParamSchema{
					{Name: "token", Type: "string", Description: "Bot token"},
				},
			},
			{
				Name:        "stopAll",
				Description: "Stop all running bots",
			},
		},
		RawTypes: `interface TelegramUser {
    id: number;
    isBot: boolean;
    firstName: string;
    lastName?: string;
    username?: string;
    languageCode?: string;
}

interface TelegramChat {
    id: number;
    type: string;
    title?: string;
    username?: string;
    firstName?: string;
    lastName?: string;
}

interface TelegramPhotoSize {
    fileId: string;
    fileUniqueId: string;
    width: number;
    height: number;
    fileSize?: number;
}

interface TelegramDocument {
    fileId: string;
    fileUniqueId: string;
    fileName?: string;
    mimeType?: string;
    fileSize?: number;
}

interface TelegramMessage {
    messageId: number;
    date: number;
    text?: string;
    chat: TelegramChat;
    from?: TelegramUser;
    photo?: TelegramPhotoSize[];
    document?: TelegramDocument;
}

interface TelegramCallbackQuery {
    id: string;
    from: TelegramUser;
    data?: string;
    chatInstance: string;
    message?: TelegramMessage;
}

interface TelegramUpdate {
    updateId: number;
    message?: TelegramMessage;
    callbackQuery?: TelegramCallbackQuery;
}

interface InlineKeyboardButton {
    text: string;
    url?: string;
    callbackData?: string;
    callback_data?: string;
}

interface KeyboardButton {
    text: string;
    requestContact?: boolean;
    requestLocation?: boolean;
}

interface SendMessageOptions {
    inlineKeyboard?: InlineKeyboardButton[][];
    keyboard?: KeyboardButton[][];
    removeKeyboard?: boolean;
    resizeKeyboard?: boolean;
    oneTimeKeyboard?: boolean;
    parseMode?: "HTML" | "Markdown" | "MarkdownV2";
    disableWebPagePreview?: boolean;
}

interface SendPhotoOptions {
    caption?: string;
    inlineKeyboard?: InlineKeyboardButton[][];
}

interface SendDocumentOptions {
    caption?: string;
    filename?: string;
    inlineKeyboard?: InlineKeyboardButton[][];
}

interface EditMessageOptions {
    inlineKeyboard?: InlineKeyboardButton[][];
}

interface TelegramContext {
    /** The raw update object */
    update: TelegramUpdate;
    /** Reply with a text message */
    reply(text: string): TelegramMessage;
    /** Reply with a photo */
    replyPhoto(photo: string, caption?: string): TelegramMessage;
    /** Reply with text and reply keyboard */
    replyWithKeyboard(text: string, keyboard: KeyboardButton[][], options?: { resize?: boolean; oneTime?: boolean; placeholder?: string }): TelegramMessage;
    /** Reply with text and inline keyboard */
    replyWithInlineKeyboard(text: string, keyboard: InlineKeyboardButton[][]): TelegramMessage;
    /** Answer callback query (for inline buttons) */
    answerCallback(text?: string, showAlert?: boolean): void;
    /** Edit the message (for callback queries) */
    editMessage(text: string, options?: EditMessageOptions): TelegramMessage;
    /** Delete the current message */
    deleteMessage(): void;
}

interface TelegramBotInstance {
    /** Register a handler for a command or text pattern */
    handle(pattern: string, handler: (ctx: TelegramContext) => void): void;
    /** Register a handler for callback query data */
    handleCallback(data: string, handler: (ctx: TelegramContext) => void): void;
    /** Register a default handler for unmatched messages */
    handleDefault(handler: (ctx: TelegramContext) => void): void;
    /** Send a text message */
    sendMessage(chatId: number, text: string, options?: SendMessageOptions): TelegramMessage;
    /** Send a photo (file path, URL, file_id, or base64) */
    sendPhoto(chatId: number, photo: string, options?: SendPhotoOptions): TelegramMessage;
    /** Send a document (file path, URL, file_id, or base64) */
    sendDocument(chatId: number, document: string, options?: SendDocumentOptions): TelegramMessage;
    /** Send a sticker */
    sendSticker(chatId: number, sticker: string, options?: SendMessageOptions): TelegramMessage;
    /** Send a video (file path, URL, or file_id) */
    sendVideo(chatId: number, video: string, options?: SendPhotoOptions): TelegramMessage;
    /** Send audio (file path, URL, or file_id) */
    sendAudio(chatId: number, audio: string, options?: SendPhotoOptions): TelegramMessage;
    /** Send voice message (file path, URL, or file_id) */
    sendVoice(chatId: number, voice: string, options?: SendPhotoOptions): TelegramMessage;
    /** Edit a message */
    editMessage(chatId: number, messageId: number, text: string, options?: EditMessageOptions): TelegramMessage;
    /** Edit message media (photo) */
    editMessageMedia(chatId: number, messageId: number, photo: string, options?: SendPhotoOptions & EditMessageOptions): TelegramMessage;
    /** Delete a message */
    deleteMessage(chatId: number, messageId: number): void;
    /** Answer a callback query */
    answerCallback(callbackId: string, text?: string, showAlert?: boolean): void;
    /** Get bot info */
    getMe(): TelegramUser;
    /** Get chat member info */
    getChatMember(chatId: number, userId: number): { status: string; user?: TelegramUser };
}`,
	}
}
