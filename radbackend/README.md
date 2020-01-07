# RadBackend readme
Radbackend это пример бэкэнда для Radmicro. 

## Настройка
### config.toml

- SQL
-  Query
db.Prepare проверяет только синтаксис. Количество и валидность аргументов проверяется только при db.Exec


### environment

## Как работает

0. Subscribe Nats QueueName = *Config.Nats.QueueName*, Subject = *Config.Nats.ServiceName* + **.req.\***
1. Ждем сообщения msg  (msg.ReplyTo содержит ключ запроса)
2. Из msg.Subject выделяем остаток строки последней точки, это будет Topic (например Auth или Start)
3. По ключу Topic выбираем query=*Sql.Query\[Topic\]*
4. Если query не существует - переход к 1 (это позволяет разделить бэкэнды по назначению)
5. Если query.Cacheable == false (аккаунтинг, не возвращает данные, не кешируется)
    - query.Prepared.Exec(Attributes)
6. query.Cacheable==true (авторизация, возвращает данные)
    - Берем ключ из ReplyTo
    - Читаем кэш Redis по ключу, 
    - при отсутствии в кэше записи
        - query.Prepared.Exec(Attributes)
        - сохраняем в кэш
    - отвечаем в NATS Subject=msg.ReplyTo и возвращаемся к п.1
7. Переход к п.1



