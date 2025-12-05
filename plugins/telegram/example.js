/**
 * Telegram Bot Example: Yes/No Fortune Teller
 *
 * This bot predicts yes or no answers with images
 *
 * Required:
 * - Set TELEGRAM_TOKEN environment variable
 * - Upload yes.png and no.png to storage
 */

const TOKEN = $env.get("TELEGRAM_TOKEN");

$telegram.startBot(TOKEN, ($bot) => {
    $logger.info("Fortune Teller Bot started!");

    // Handle /start command
    $bot.handle("/start", (ctx) => {
        const user = ctx.update.message.from;
        const name = user.firstName || "stranger";

        ctx.replyWithInlineKeyboard(
            `<b>Welcome, ${name}!</b>\n\n` +
            `I can predict YES or NO for any question.\n\n` +
            `Just ask me anything or press the button below!`,
            [
                [{ text: "Ask a question", callbackData: "ask" }],
                [{ text: "About", callbackData: "about" }]
            ]
        );
    });

    // Handle /help command
    $bot.handle("/help", (ctx) => {
        ctx.reply(
            `<b>How to use:</b>\n\n` +
            `1. Type any question\n` +
            `2. I will answer YES or NO\n` +
            `3. Trust the magic!\n\n` +
            `<b>Commands:</b>\n` +
            `/start - Start the bot\n` +
            `/help - Show this help\n` +
            `/fortune - Get a random fortune`
        );
    });

    // Handle /fortune command
    $bot.handle("/fortune", (ctx) => {
        predictFortune(ctx);
    });

    // Handle "Ask a question" button
    $bot.handleCallback("ask", (ctx) => {
        ctx.answerCallback();
        ctx.editMessage(
            `<b>Ask your question!</b>\n\n` +
            `Type any question and I will predict the answer...`,
            { inlineKeyboard: [[{ text: "Back", callbackData: "back" }]] }
        );
    });

    // Handle "About" button
    $bot.handleCallback("about", (ctx) => {
        ctx.answerCallback();
        ctx.editMessage(
            `<b>Fortune Teller Bot</b>\n\n` +
            `Version: 1.0\n` +
            `Powered by: M3M Platform\n\n` +
            `This bot uses ancient algorithms to predict YES or NO.`,
            { inlineKeyboard: [[{ text: "Back", callbackData: "back" }]] }
        );
    });

    // Handle "Back" button
    $bot.handleCallback("back", (ctx) => {
        ctx.answerCallback();
        ctx.editMessage(
            `<b>Fortune Teller</b>\n\n` +
            `Ask me any question and I will predict YES or NO!`,
            {
                inlineKeyboard: [
                    [{ text: "Ask a question", callbackData: "ask" }],
                    [{ text: "About", callbackData: "about" }]
                ]
            }
        );
    });

    // Handle "Ask again" button
    $bot.handleCallback("again", (ctx) => {
        ctx.answerCallback("Ask another question!", false);
    });

    // Default handler for text messages (questions)
    $bot.handleDefault((ctx) => {
        if (ctx.update.message && ctx.update.message.text) {
            const text = ctx.update.message.text;

            // Skip if it's a command
            if (text.startsWith("/")) {
                ctx.reply("Unknown command. Try /help");
                return;
            }

            // Check if it looks like a question
            if (text.length < 3) {
                ctx.reply("Please ask a proper question!");
                return;
            }

            predictFortune(ctx);
        }
    });

    // Fortune prediction function
    function predictFortune(ctx) {
        const isYes = Math.random() > 0.5;
        const image = $storage.getPath(isYes ? "yes.png" : "no.png");
        const answer = isYes ? "YES" : "NO";

        const messages = isYes ? [
            "The stars say YES!",
            "Absolutely YES!",
            "Without a doubt - YES!",
            "YES! Go for it!",
            "The universe confirms: YES!"
        ] : [
            "The answer is NO...",
            "Unfortunately, NO.",
            "The spirits say NO.",
            "NO, not this time.",
            "The cosmos whispers: NO."
        ];

        const message = messages[Math.floor(Math.random() * messages.length)];
        const chatId = ctx.update.message.chat.id;

        // Try to send with image from storage
        try {
            $bot.sendPhoto(chatId, image, {
                caption: `<b>${answer}</b>\n\n${message}`,
                inlineKeyboard: [
                    [{ text: "Ask again", callbackData: "again" }],
                    [{ text: "Share", url: "https://t.me/share/url?text=I%20got%20my%20fortune!" }]
                ]
            });
        } catch (e) {
            // Fallback to text if image not found
            $logger.warn("Image not found, sending text only: " + e.message);
            ctx.replyWithInlineKeyboard(
                `<b>${answer}</b>\n\n${message}`,
                [
                    [{ text: "Ask again", callbackData: "again" }]
                ]
            );
        }
    }
});

// Graceful shutdown
$service.shutdown(() => {
    $logger.info("Stopping Fortune Teller Bot...");
    $telegram.stopAll();
});
