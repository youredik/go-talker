# Тестовое задание "Говорун" (Go Junior)

### Задание

1) ``/listen`` - sse эндпоинт, который раз в секунду по sse шлёт текущее некоторое слово всем подключенным к нему клиентам.(плюсом будет если сделаешь без цикла)

2) ``/say`` - post эндпоинт, принимает в себя json ``{word:"некоторое слово"}`` и заменяет этим словом слово, которое рассылает listen.

3) (дополнительное) покрыть это всё тестом, т.е. при помощи какого либо тестового фрейморка написать клиент который проверит как сценарий с 1 говоруном, так и с ``n`` говорунов и ``n`` слушателей


### Как запустить?


1) Команда ``go run main.go``

2) Открыть ``http://127.0.0.1:8080`` в браузере и смотреть сообщения

3) Чтобы сменить слово говоруна, в терминале выполнить команду: ``curl http://127.0.0.1:8080/say --include --header "Content-Type: application/json" --request "POST" --data '{"word": "Changed message"}'``