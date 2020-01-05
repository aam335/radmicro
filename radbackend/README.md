# RadBackend readme
Radbackend это пример бэкэнда для Radmicro. 

## Настройка
### config.toml
### environment

## Как работает
1. Ждет сообщения NATS с Subject ServiceName.Topic (ReplyTo содержит ключ запроса)
2. Читает кэш Redis по ключу , при наличии записи отвечает в ReplyTo и возвращается к п.1
3. Спрашивает по Topic запросу SQL
4. Отвечает в ReplyTo
5. Сохраняет в Redis
6. Переход к п.1



