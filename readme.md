Чат(серверная и клиентская часть)
<a href="https://goreportcard.com/badge/github.com/ruslanBik4/chat"> <img src="https://goreportcard.com/badge/github.com/ruslanBik4/chat" /> </a>

 Настройка параметров соединения проихводится с помощью флагов командной строки.
 Есть простой режим отладки.

Сервер.

 Реализует простой чат, передача сообщения от одного участника ко всем остальным
 Есть возможность передавать на сервер файлы с ограничением до 1 гб

Клиент.
  Позволяет передавать и получать сообщения с сервера.

	Перечень доступных команд:
	"file:" - отправить файл,
	":list" - получить список файлов с сервера
	":file" - получить файл с сервера
	":nick" - сменить текущий нил
	":register" - зарегистрироть пароль для ника
	":exit"  - завершить работу