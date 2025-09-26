package utils

import "strings"

// Findval extrait la valeur associée à une clé dans une chaîne de message encodée.

// Format attendu du message :
// - Le premier caractère indique le séparateur entre les paires clé=valeur (par exemple `/`)
// - Ensuite viennent plusieurs paires clé=valeur séparées par ce caractère。
// - Ce format est utilisé dans la communication FIFO entre processus.

// Paramètres :
//   - p_nom : pointeur vers le nom de l’application
//   - msg   : le message complet reçu
//   - key   : la clé dont on cherche la valeur

// Retourne :
//   - La valeur correspondant à la clé si trouvée, sinon chaîne vide ""
//   - En cas d’erreur de format, un avertissement est affiché

func Findval(p_nom *string, msg string, key string) string {
	if len(msg) < 4 {
		Display_w(p_nom, "findval", "message trop court: "+msg)
		return ""
	}
	sep := msg[0:1]
	tab_allkeyvals := strings.Split(msg[1:], sep)
	for _, keyval := range tab_allkeyvals {
		if len(keyval) < 2 {
			return ""
		}
		equ := keyval[0:1]
		tabkeyval := strings.Split(keyval[1:], equ)
		if tabkeyval[0] == key {
			return tabkeyval[1]
		}
	}
	return ""
}
