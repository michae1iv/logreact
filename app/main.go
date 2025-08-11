package main

import (
	"context"
	"correlator/api"
	"correlator/config"
	"correlator/db"
	"correlator/logger"
	"correlator/regulars"
	"correlator/rule_manager"
	"correlator/rw"
	"fmt"
	"log"
	"time"
)

/*
	 TODO:
	 1. Создание обработчика сложных правил, должно учитываться:
	 		//1.1. Список вложенных правил (подправила)
			//1.2. Условия для срабатывания данных правил, например (двукратное срабатывание подправила, срабатывание после определенного правила)
			//1.3. Возможность выбора, принимать сработки от множества хостов, или только от одного
	 2. Добавить CSRF Protection на страницу логина
	 //3. Написать тестовые правила
	 4. Добавить проверку сложности пароля
//	 5. Переписать работу бэкенда в формате JSON-Response, JSON-Request
	 6. Фронтэнд на Next.js
//	 7. Добавить уровень критичности для правил
//	 8. Сделать отдельные обработчики, для каждого правила (Может сделать один обработчик под группу правил)
	 9. Сделать возможность настройки группы хостов (Администратор сможет устанавливать группы хостов,
	 		 для того, чтобы легче было писать правила)

//	 10. Добавить конфиги для подключения к БД, кафке
 11. Написать тесты на бэк и фронт

//	 12. Настроить логирование приложения
 13. Проверить возможность работы клиента на Astra
 14. Сделать нормальную настройку прав пользователей (Один суперпользователь, права через базу данных)
// 15. Убрать страницу регистрации
 16. Сделать возможность сброса пароля для суперпользователя (Двухфакторка?)
 17. Сделать нормальную админ-панель
 18. Сделать возможность проверки правила на определенном наборе логов

//	 19. Убрать actions из правил (либо сделать их в виде отдельной сущности, т.е закреплять действия за событиями)
 20. Попробовать создать функцию, которая бы могла брать правила с определленой api
 21. Добавить возможность не создавать алерт на определенное правило
// 22. Добавить .gitignore
*/

func main() {
	cfg, err := config.LoadConfig("config.yaml") // Load configuration from config.yaml
	if err != nil {
		log.Fatalf("Error in loading configuration: %v", err)
	}

	logger.InitLoggers(&cfg.Logging) // Initialising error.log and success.log to write down all information about application

	err = regulars.CompileAllExpressions() // compiling all regulars
	if err != nil {
		logger.ErrorLogger.Fatalf("An error occured while compiling regular expressions: %v", err)
	}

	db.AutoMigrateAndStartDB(&cfg.Database) // starting database

	// Starting GlobalHandler
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	rule_manager.GlobalHandlerObj.IsReady = false
	rule_manager.GlobalHandlerObj.FramePointer = make(map[string]*rule_manager.Frame)
	rule_manager.GlobalHandlerObj.LogChan = make(chan []byte, cfg.RuleHandler.ReadBufferSize)
	rw.WChan = make(chan []byte, cfg.RuleHandler.WriterBufferSize)
	rule_manager.GlobalHandlerObj.RuleChan = make(chan *rule_manager.Rule)
	go rule_manager.GlobalHandlerObj.Start(ctx)

	for !rule_manager.GlobalHandlerObj.IsReady { // while GlobalHandler isn't ready readers and writers must wait
		time.Sleep(100 * time.Millisecond)
	}
	rw.InitReaders(ctx, &cfg.Reader, rule_manager.GlobalHandlerObj.LogChan) // start readers
	go rw.InitWriters(ctx, &cfg.Writer)                                     // start writers

	go api.InitRouter(cfg.Server, cfg.Authentication.Secret) // start Api

	fmt.Println("Ready to Go")

	select {}
}
