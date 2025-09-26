package utils

import (
	"log"
	"os"
)

// Délimiteurs utilisés dans les messages formatés (clé=valeur/clé=valeur/...)
var fieldsep = "/"
var keyvalsep = "="

var Cyan string = "\033[1;36m"
var rouge string = "\033[1;31m"
var purple string = "\033[0;35m"
var Raz string = "\033[0;00m"
var Jaune string = "\033[1;33m"

// PID du processus courant
var pid = os.Getpid()
var stderr = log.New(os.Stderr, "", 0)

//var fileMutex sync.Mutex

func Display_d(p_nom *string, where string, what string) {
	stderr.Printf(" + [%.20s %d] %-16s : %s\n", *p_nom, pid, where, what)
}

func Display_w(p_nom *string, where string, what string) {
	stderr.Printf("%s * [%.20s %d] %-16s : %s\n%s", purple, *p_nom, pid, where, what, Raz)
}

func Display_e(p_nom *string, where string, what string) {
	stderr.Printf("%s ! [%.20s %d] %-16s : %s\n%s", rouge, *p_nom, pid, where, what, Raz)
}

func Display_info(p_nom *string, where string, what string) {
	stderr.Printf("%s i [%s %d] %s : %s\n%s", Jaune, *p_nom, pid, where, what, Raz)
}

func Msg_send(sender *string, msg string) {
	Display_d(sender, "msg_send", "émission de "+msg)
}

func Msg_receive(p_nom *string, msg string) {
	where := "msg_receive"
	what := "réception de " + msg
	stderr.Printf("%s ! [%.20s %d] %s : %s\n%s", Cyan, *p_nom, pid, where, what, Raz)
}

func Msg_format(key string, val string) string {
	return fieldsep + keyvalsep + key + keyvalsep + val
}
