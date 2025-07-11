package handlers

const (
	// Help Message
	MsgHelp = "Вот что я умею:\n\n*Команды:*\n/start — приветствие и главное меню\n/help — показать это сообщение\n\n*Кнопки:*\n\"%s\" — краткая сводка текущей погоды\n\"%s\" — картинка с котом и комплиментом\n\"%s\" — меню планов и напоминаний\n\"%s\" — время ваших отношений\n\"%s\" — поможет тебе принять решение\n\"%s\" — меню твоих настроек"

	// Start messages
	MsgWelcome             = "Привет\\! Я *Вкущуща* — твой романтический помощник 💌\n\n"
	MsgStartBase           = "Можешь сразу нажать на одну из кнопок или позвать команду /help, чтобы ознакомиться с моими функциями подробнее.\n\n\n"
	MsgStartDeveloper      = "От разработчика:\nМеня зовут Алексей @Petryanin\n"
	MsgStartSource         = "Исходный код открыт и выложен на [github](https://github.com/Petryanin/love-bot)"
	MsgStartGratitude      = "Я очень старался и буду благодарен за ⭐"
	MsgCommandNotAvailable = "В данный момент команда недоступна😢"
	MsgSettingsStart       = "Для начала мне нужно узнать тебя получше. Пожалуйста, отправь мне название своего города или твою геолокацию.\n\nЭто поможет мне предоставлять актуальную сводку погоды и учитывать часовой пояс при напоминаниях"
	MsgUnknownMessageType  = "🧐Не могу распознать такой тип сообщения, попробуй ещё"
	MsgPartnerPrompt       = "\n\nТеперь отправь мне Telegram-ник твоего партнёра.\n\nЭто поможет мне учитывать ваши совместные планы."
	MsgSettingsComplete    = "\n\nСупер, всё готово!👍\n\n"

	// Settings menu messages
	MsgCurrentSettings        = "*Ваши текущие настройки:*\n\n\\- город: *%s*\n\\- часовой пояс: *%s*\n\\- партнер: @%s\n\\- котики: *%s*"
	MsgCatTimeDisabled        = "Отключены"
	MsgCatTimeEnabled         = "Каждый день в %s"
	MsgUserInfoError          = "Упс, не удалось получить информацию по пользователю 😿\nПопробуй позже"
	MsgCityPrompt             = "Пожалуйста, отправь мне название своего города или твою геолокацию.\n\nЭто поможет мне предоставлять актуальную сводку погоды и учитывать часовой пояс при напоминаниях"
	MsgPartnerPromptSettings  = "Пожалуйста, введи Telegram-ник твоего партнёра.\n\nЭто поможет мне учитывать ваши совместные планы."
	MsgCatPrompt              = "Пожалуйста, введи время, в которое тебе ежедневно будут прилетать котики (в формате HH:MM) или нажми «%s», чтобы отказаться от подписки."
	MsgCatTimeUpdateError     = "Упс, произошла ошибка. Обратись к администратору"
	MsgCatTimeDisabledSuccess = "👌Хорошо, больше не буду присылать тебе котиков"
	MsgInvalidTimeFormat      = "😿Неверный формат. Введи время в виде HH:MM, например «18:30»"
	MsgCatTimeSaveError       = "Упс, не удалось сохранить время 😿\nПопробуй позже"
	MsgCatTimeSaved           = "✅ Время сохранено! Ежедневно в %s тебе будут прилетать котики😻"

	// Settings and Start messages
	MsgLocationNotFound = "🧐Не смог распознать геолокацию, попробуй ещё"
	MsgCitySaved        = "Город сохранён: *%s*\nЧасовой пояс: *%s*"
	MsgInvalidUsername  = "🧐Не смог расположить имя пользователя. Проверь правильность написания и попробуй ещё"
	MsgPartnerNotFound  = "Упс, не смог добавить твоего партнера 😿\n\nПроверь правильность написания и попробуй ещё.\n\nЕсли не поможет, то возможно это связано с тем, что у нас ещё не было диалога. Буду ждать, пока вы оба зарегистрируетесь 🤗\n\nЕсли и это не поможет, обратись к администратору"
	MsgPartnerSaved     = "Партнёр сохранён: %s"

	// Compliment messages
	MsgComplimentImageError = "Упс, комплимент где-то зажевался😿. Пишу что-то от себя..."

	// Weather messages
	MsgWeatherError        = "Упс, не удалось получить погоду 😿\nПопробуй позже"
	MsgWeatherSummaryError = "Не удалось получить погоду 😕"

	// Magic ball messages
	MsgMagicBallThinking1    = "🔮 Я вижу тьму"
	MsgMagicBallThinking2    = "🔮 Образы начинают проясняться"
	MsgMagicBallThinking3    = "🔮 Судьба медленно раскрывается"
	MsgMagicBallCharging     = "🔮 Зарядка шара"
	MsgMagicBallAccumulating = "🔮 Энергия накапливается"
	MsgMagicBallAwakening    = "🔮 Оракул пробуждается"
	MsgMagicBallProphesying  = "🔮 Пророчество формируется"

	// Plan messages
	MsgPlanMenu             = "Меню планов: о чем вам напомнить?"
	MsgPlanAdd              = "Введи текст напоминания"
	MsgPlanTimeInFuture     = "🧐Время должно быть в будущем, попробуй ещё"
	MsgPlanAskRemindTime    = "Когда напомнить?"
	MsgPlanAskEventTime     = "Когда это событие произойдёт?"
	MsgPlanTimeFormatError  = "🧐Не смог распознать формат, попробуй ещё"
	MsgPlanRemindTimeError  = "🧐Время напоминания не может быть больше времени события, попробуй ещё"
	MsgPlanSaveError        = "😥Ошибка при сохранении"
	MsgPlanSaved            = "✅План сохранён!\n\n%s %s\n(напомню %s)"
	MsgPlanPartner          = "Твоя Вкущуща создала новый план: %s на %s"
	MsgPlanDeleted          = "👌Напоминание успешно удалено"
	MsgPlanNoPlans          = "У вас нет планов"
	MsgPlanList             = "%s\n\nВыбери план для подробностей:"
	MsgPlanChangeRemindTime = "Введи время повторного напоминания"
	MsgPlanNotFound         = "🧐Такого плана не найдено"
	MsgPlanScheduleError    = "😥Ошибка, не удалось сохранить новое время"
	MsgPlanRemindAgain      = "%s\n\nХорошо, напомню снова в это время: %s"
	MsgPlanReminder         = "📢Напоминание: %s (%s)\n\n Напомнить снова через:"
)
