package main

import (
	"bufio"
	"log"
	"log/syslog"
	"net"
	"os"
	"strings"

	"github.com/TRedzepagic/compositelogger/logs"
	_ "github.com/go-sql-driver/mysql"
)

// More information on goroutines :
// 1 : https://www.golang-book.com/books/intro/10
// 2 : https://tour.golang.org/concurrency/1
// 3 : https://golangbot.com/goroutines/

// User creation command used : "CREATE USER 'compositelogger'@'localhost' IDENTIFIED BY 'Mystrongpassword1234$'"
// Database : LOGGER
// Table : LOGS
// You can infer the username and password from this.

// Handler handles the connection and logger symbiosis
func Handler(connection net.Conn, log *logs.CompositeLog) {
	log.Infof("%s has connected to the server", connection.RemoteAddr().String())

	//infinite loop until exited from inside
	for {
		data, err := bufio.NewReader(connection).ReadString('\n')
		if err != nil {
			log.Error("Listening of " + connection.RemoteAddr().String() + " has stopped (disconnected)?...")
			return
		}

		stringdata := strings.TrimSpace(string(data))
		if len(stringdata) == 0 {
			result := "Server responds : Empty input..." + "\n"
			log.Info("Client sent empty string, ignoring")
			connection.Write([]byte(string(result)))
		} else {
			temp := strings.TrimSpace("Address : " + connection.RemoteAddr().String() + " says : " + stringdata)
			log.Info(temp)
			result := "Server has received your message" + "\n"
			connection.Write([]byte(string(result)))
		}
		defer connection.Close()
	}
}

func main() {
	// Logger creation
	filepath := "serverlog"
	filelogger1 := logs.NewFileLogger(filepath)
	defer filelogger1.Close()

	stdoutLog := logs.NewStdLogger()
	defer stdoutLog.Close()

	systemlogger, _ := logs.NewSysLogger(syslog.LOG_NOTICE, log.LstdFlags)

	databaseLog := logs.NewDBLogger(logs.DatabaseConfiguration())
	defer databaseLog.Close()

	wantDebug := false

	log := logs.NewCustomLogger(wantDebug, filelogger1, stdoutLog, systemlogger, databaseLog)

	// Listen server implementation
	// Check for arguments on startup (Exit if no arguments provided (specifically, port number))
	arguments := os.Args
	if len(arguments) == 1 {
		log.Warn("No port number specified... Exiting.....")
		return
	}

	log.Info("Listening...")
	PORT := ":" + arguments[1]
	listener, err := net.Listen("tcp", PORT)
	if err != nil {
		log.Error(err)
		return
	}
	for {
		connection, err := listener.Accept()
		if err != nil {
			log.Error(err)
			return
		}
		go Handler(connection, log)
	}
}
