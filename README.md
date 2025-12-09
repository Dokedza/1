Веб-сервис для доступности сайтов

отравляешь список сайтов - получаешь статусы
хранит историю
генерирует отчёты
сохраняет данные при остановке

запустить go run main.go

сервер запустится на http://localhost:8080

curl http://localhost:8080/health - проверка здоровья сервиса

curl -X POST http://localhost:8080/api/check \
  -H "Content-Type: application/json" \
  -d '{"links": ["google.com", "yandex.ru"]}' - для проверки ссылки

  curl http://localhost:8080/api/status/1 -узнать результат

  curl -X POST http://localhost:8080/api/report \
  -H "Content-Type: application/json" \
  -d '{"links_list": [1]}' \
  --output report.pdf - получить PDF отчёт

  Проверка происходит асинхронно, есть несколько потоков для проверки, данные сохраняются в файл, остановка без потерь
